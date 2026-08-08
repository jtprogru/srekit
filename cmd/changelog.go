package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

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
  srekit changelog validate`,
		// The generator takes no positional arguments; NoArgs on the parent
		// turns a mistyped subcommand into an error instead of a silently
		// ignored word. Cobra resolves a real subcommand before consulting
		// this, so `changelog release` is unaffected.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			manifest, err := loadChangelogManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, input.Sections, struct{ Meta changelogMeta }{Meta: m})
			if err != nil {
				return err
			}

			data := changelogData{Meta: m, Sections: rendered}
			opts := out.RenderOptions(cmd, "CHANGELOG.md")
			return render.Render(cmd.OutOrStdout(), loader, "changelog", data, opts)
		},
	}
	cmd.Flags().StringVar(&repoFlag, "repo", "", "OWNER/REPO for compare links (default: detect from git remote)")
	cmd.Flags().StringVar(&version, "version", "0.1.0", "initial version label")
	cmd.Flags().StringVar(&from, "from", "", "read sections from JSON file (- for stdin)")
	out.Bind(cmd, "write to file (default: CHANGELOG.md)")
	cmd.AddCommand(newChangelogReleaseCmd(), newChangelogValidateCmd())
	return cmd
}

func loadChangelogManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("changelog")
	if err != nil {
		return nil, fmt.Errorf("load changelog.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse changelog.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "changelog.yaml", "changelog.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
