package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/config"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// changelogDefaultLang is English and stays English: everything that greps
// `### Added`, parses version headings or drafts release notes does so in
// English, so the Russian variant has to be asked for out loud.
const changelogDefaultLang = "en"

// changelogLangs are the languages the shipped changelog artifacts cover.
// It is a function rather than a package-level slice so nothing can mutate
// the set.
func changelogLangs() []string { return []string{changelogDefaultLang, "ru"} }

type changelogMeta struct {
	Today          string `json:"today"`
	Repo           string `json:"repo"`
	InitialVersion string `json:"initialVersion"`
}

type changelogData struct {
	Meta     changelogMeta              `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d changelogData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta changelogMeta }{Meta: d.Meta}
}

func newChangelogCmd() *cobra.Command {
	var (
		repoFlag string
		version  string
		from     string
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Generate a CHANGELOG.md scaffold (Keep a Changelog format)",
		Long: `Generate a CHANGELOG.md scaffold in Keep a Changelog format.

The bare invocation generates. The release and validate subcommands maintain
a changelog that already exists — they are not catalog entries of their own,
so ` + "`srekit changelog`" + ` keeps the behaviour it always had.`,
		Example: `  # Write CHANGELOG.md, repository slug from the git remote
  srekit changelog

  # Inspect the structured JSON shape
  srekit changelog --repo acme/api --json | jq '.sections[].id'

  # Round-trip: fill the unreleased section and re-render Markdown
  srekit changelog --repo acme/api --json > cl.json
  # ...edit cl.json...
  srekit changelog --from cl.json

  # Cut a release: move [Unreleased] under a dated heading
  srekit changelog release --version 1.2.0

  # Check an existing changelog against the format
  srekit changelog validate

  # Russian change types (Добавлено, Изменено, ...) instead of English
  srekit changelog --lang ru`,
		// The generator takes no positional arguments; NoArgs on the parent
		// turns a mistyped subcommand into an error instead of a silently
		// ignored word. Cobra resolves a real subcommand before consulting
		// this, so `changelog release` is unaffected.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// The language is resolved first: an unrecognized value must fail
			// before anything is read, rendered or written.
			lang, err := resolveChangelogLang(cmd)
			if err != nil {
				return err
			}

			// The payload is read before slug resolution: meta.repo in the
			// file is one of the three ways a slug can arrive, so detection
			// must not fail before the file has been consulted.
			input, err := sections.ReadPayload(cmd.InOrStdin(), from)
			if err != nil {
				return err
			}

			repo := sections.PickNonEmpty(repoFlag, input.Meta["repo"])
			if repo == "" {
				r, detectErr := meta.DetectRepo()
				if detectErr != nil {
					return fmt.Errorf("could not detect repo from git remote: %w (pass --repo OWNER/REPO)", detectErr)
				}
				repo = r.Slug()
			}

			// The version flag carries a default, so "user asked for this"
			// has to be read off Changed rather than off the value.
			initialVersion := version
			if !cmd.Flags().Changed("version") {
				initialVersion = sections.PickNonEmpty(input.Meta["initialVersion"], version)
			}

			m := changelogMeta{
				Today:          sections.PickNonEmpty(input.Meta["today"], clock.Now().Format("2006-01-02")),
				Repo:           repo,
				InitialVersion: initialVersion,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadChangelogManifest(cmd, loader, lang)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, input.Sections, struct{ Meta changelogMeta }{Meta: m})
			if err != nil {
				return err
			}

			data := changelogData{Meta: m, Sections: rendered}
			opts := out.RenderOptions(cmd, "CHANGELOG.md")
			opts.Lang = lang
			return render.Render(cmd.OutOrStdout(), loader, "changelog", data, opts)
		},
	}
	cmd.Flags().StringVar(&repoFlag, "repo", "", "OWNER/REPO for compare links (default: detect from git remote)")
	cmd.Flags().StringVar(&version, "version", "0.1.0", "initial version label")
	cmd.Flags().StringVar(&from, "from", "", "read sections from JSON file (- for stdin)")
	// Persistent, so `release` and `validate` run under the same selection —
	// but only generation acts on it. Both maintenance subcommands read the
	// change-type vocabulary out of the document in front of them, never out
	// of this flag; see changelog.Vocabularies.
	cmd.PersistentFlags().String("lang", "",
		"language of the generated change types: en or ru (default: en, or changelog_lang in config)")
	out.Bind(cmd, "write to file (default: CHANGELOG.md)")
	cmd.AddCommand(newChangelogReleaseCmd(), newChangelogValidateCmd())
	return cmd
}

// resolveChangelogLang applies the flag → `changelog_lang` → `en`
// precedence and rejects anything outside the shipped set. Both callers
// call it before touching a file, so a typo cannot get as far as writing
// one.
func resolveChangelogLang(cmd *cobra.Command) (string, error) {
	source := "--lang"
	lang := ""
	if f := cmd.Flags().Lookup("lang"); f != nil {
		lang = strings.TrimSpace(f.Value.String())
	}
	if lang == "" {
		source = "changelog_lang"
		lang = strings.TrimSpace(config.Global().GetString("changelog_lang"))
	}
	if lang == "" {
		return changelogDefaultLang, nil
	}
	lang = strings.ToLower(lang)
	if !slices.Contains(changelogLangs(), lang) {
		return "", fmt.Errorf("%s %q: unknown language; accepted values are %s",
			source, lang, strings.Join(changelogLangs(), ", "))
	}
	return lang, nil
}

func loadChangelogManifest(cmd *cobra.Command, loader *tmpl.Loader, lang string) (*sections.Manifest, error) {
	artifactName := tmpl.ArtifactVariantNameFor("changelog", lang)
	artifactBytes, err := loader.LoadArtifactBytesLang("changelog", lang)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", artifactName, err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", artifactName, err)
	}
	warnStaleLegacyFiles(cmd, loader, artifactName, "changelog.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
