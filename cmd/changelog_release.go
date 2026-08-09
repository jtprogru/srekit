package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/changelog"
	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/meta"
)

// defaultChangelogPath is the target both maintenance subcommands assume
// when given no positional argument.
const defaultChangelogPath = "CHANGELOG.md"

func newChangelogReleaseCmd() *cobra.Command {
	var (
		version string
		date    string
		yanked  bool
		edit    cliflags.Edit
	)
	cmd := &cobra.Command{
		Use:   "release [FILE]",
		Short: "Cut a release: move [Unreleased] under a dated version heading",
		Long: `Move everything under ## [Unreleased] into a new ## [X.Y.Z] - YYYY-MM-DD
heading, leave [Unreleased] empty, and update the link reference block so
[Unreleased] compares the new tag against HEAD.

Change-type subsections holding nothing but the scaffold's bare "-" are
dropped, so a released version never ships an empty ### Deprecated.

The command edits text. It does not commit, tag or push — review the diff,
then commit and tag yourself.`,
		Example: `  # Preview first
  srekit changelog release --version 1.2.0 --dry-run

  # Cut it
  srekit changelog release --version 1.2.0

  # Backfill a historical, withdrawn release
  srekit changelog release --version 0.0.5 --date 2014-12-13 --yanked

  # A changelog somewhere else
  srekit changelog release --version 1.2.0 docs/CHANGELOG.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// --lang is inherited from the parent and governs generation, not
			// parsing — the vocabulary below is detected from the document.
			// It is still validated here so an unrecognized value fails in
			// the whole command group rather than only where it is used.
			if _, err := resolveChangelogLang(cmd); err != nil {
				return err
			}

			// The date is checked before the file is read so a typo cannot
			// get as far as touching the user's document.
			if date == "" {
				date = clock.Now().Format("2006-01-02")
			} else if !changelog.IsISODate(date) {
				return fmt.Errorf("--date %q: expected YYYY-MM-DD", date)
			}

			path := defaultChangelogPath
			if len(args) == 1 {
				path = args[0]
			}

			src, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("%s does not exist; run `srekit changelog` to create one", path)
				}
				return err
			}

			doc := changelog.Scan(src, changelog.Vocabularies())

			// --json reports what was parsed and stops. It is a way to look
			// at the document, not a second output format for the rewrite.
			if edit.JSON {
				return emitJSON(cmd.OutOrStdout(), doc.Describe(path))
			}

			body, err := changelog.Release(doc, changelog.ReleaseOptions{
				Version: version,
				Date:    date,
				Yanked:  yanked,
				Slug:    changelogSlug(doc),
			})
			if err != nil {
				return err
			}
			return writeEdited(cmd, path, body, edit)
		},
	}
	cmd.Flags().StringVar(&version, "version", "", "version being released, e.g. 1.2.0 (required)")
	cmd.Flags().StringVar(&date, "date", "", "release date in YYYY-MM-DD form (default: today)")
	cmd.Flags().BoolVar(&yanked, "yanked", false, "mark the release withdrawn: `## [X.Y.Z] - YYYY-MM-DD [YANKED]`")
	edit.Bind(cmd, "emit the parsed document as JSON instead of rewriting the file")
	_ = cmd.MarkFlagRequired("version")
	return cmd
}

// changelogSlug resolves OWNER/REPO for a document that has no link block
// to read conventions from. A detection failure is not an error here: the
// rewrite reports the missing block itself, with a message about the
// document rather than about git.
func changelogSlug(doc *changelog.Document) string {
	if doc.LinksPresent {
		return ""
	}
	r, err := meta.DetectRepo()
	if err != nil {
		return ""
	}
	return r.Slug()
}

// writeEdited routes the result of an in-place edit. It is deliberately not
// render.writeBody: there is no --out to honour, no --force guard to apply,
// and the default sink is the source file rather than a computed path.
func writeEdited(cmd *cobra.Command, path string, body []byte, edit cliflags.Edit) error {
	stdout := cmd.OutOrStdout()
	quiet, _ := cmd.Flags().GetBool("quiet")

	if edit.DryRun {
		if !quiet {
			fmt.Fprintf(stdout, "# dry-run: would write %d bytes to %s\n", len(body), path)
		}
		_, err := stdout.Write(body)
		return err
	}
	if edit.Stdout {
		_, err := stdout.Write(body)
		return err
	}

	// The file already exists — it is the file that was just read — so the
	// mode here only applies if it was removed underneath us. 0o644 matches
	// what the generators write.
	if err := os.WriteFile(path, body, 0o644); err != nil { //nolint:gosec // G306: public documentation file, same mode as the generators
		return err
	}
	if !quiet {
		fmt.Fprintf(stdout, "wrote %s\n", path)
	}
	return nil
}

func emitJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}
