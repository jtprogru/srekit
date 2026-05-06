package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	rfcTitle  string
	rfcStatus string
	rfcAuthor string
	rfcEmail  string
	rfcOut    string
	rfcStdout bool
	rfcForce  bool
	rfcDry    bool
)

var allowedRFCStatus = map[string]bool{
	"proposed":   true,
	"accepted":   true,
	"rejected":   true,
	"superseded": true,
	"deprecated": true,
}

var rfcCmd = &cobra.Command{
	Use:   "rfc",
	Short: "Generate an RFC / ADR markdown template",
	Long:  "Generate an RFC (a.k.a. ADR) document with Context / Decision / Alternatives / Consequences sections.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if rfcTitle == "" {
			return errors.New("--title is required")
		}
		if !allowedRFCStatus[rfcStatus] {
			return fmt.Errorf("invalid --status %q (proposed, accepted, rejected, superseded, deprecated)", rfcStatus)
		}
		author, err := meta.Resolve(rfcAuthor, rfcEmail)
		if err != nil {
			return err
		}
		id := ids.UUID()
		data := struct {
			ID, ShortID, Title, Status, Now string
			Author                          meta.Author
		}{
			ID:      id,
			ShortID: id[:8],
			Title:   rfcTitle,
			Status:  rfcStatus,
			Now:     time.Now().Format(time.RFC3339),
			Author:  author,
		}
		def := fmt.Sprintf("rfc-%s.md", ids.Slug(rfcTitle))
		return render.Render(cmd.OutOrStdout(), "rfc.md.tmpl", data, render.Options{
			Out:     rfcOut,
			Stdout:  rfcStdout,
			Force:   rfcForce,
			DryRun:  rfcDry,
			Default: def,
		})
	},
}

func init() {
	rfcCmd.Flags().StringVarP(&rfcTitle, "title", "T", "", "RFC title (required)")
	rfcCmd.Flags().StringVar(&rfcStatus, "status", "proposed", "RFC status: proposed | accepted | rejected | superseded | deprecated")
	rfcCmd.Flags().StringVar(&rfcAuthor, "author", "", "author full name")
	rfcCmd.Flags().StringVar(&rfcEmail, "email", "", "author email")
	rfcCmd.Flags().StringVar(&rfcOut, "out", "", "write to file (default: rfc-<slug>.md)")
	rfcCmd.Flags().BoolVar(&rfcStdout, "stdout", false, "print to stdout")
	rfcCmd.Flags().BoolVar(&rfcForce, "force", false, "overwrite existing file")
	rfcCmd.Flags().BoolVar(&rfcDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(rfcCmd)
}
