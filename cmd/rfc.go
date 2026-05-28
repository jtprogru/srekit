package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var allowedRFCStatus = map[string]bool{
	"proposed":   true,
	"accepted":   true,
	"rejected":   true,
	"superseded": true,
	"deprecated": true,
}

func newRFCCmd() *cobra.Command {
	var (
		title  string
		status string
		author string
		email  string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "rfc",
		Short: "Generate an RFC / ADR markdown template",
		Long:  "Generate an RFC (a.k.a. ADR) document with Context / Decision / Alternatives / Consequences sections.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			if !allowedRFCStatus[status] {
				return fmt.Errorf("invalid --status %q (proposed, accepted, rejected, superseded, deprecated)", status)
			}
			a, err := meta.Resolve(viper.GetViper(), author, email)
			if err != nil {
				return err
			}
			data := struct {
				ID, Title, Status, Now string
				Author                 meta.Author
			}{
				ID:     ids.UUID(),
				Title:  title,
				Status: status,
				Now:    clock.Now().Format(time.RFC3339),
				Author: a,
			}
			def := fmt.Sprintf("rfc-%s.md", ids.Slug(title))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "rfc.md.tmpl", data, out.RenderOptions(def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "RFC title (required)")
	cmd.Flags().StringVar(&status, "status", "proposed", "RFC status: proposed | accepted | rejected | superseded | deprecated")
	cmd.Flags().StringVar(&author, "author", "", "author full name")
	cmd.Flags().StringVar(&email, "email", "", "author email")
	out.Bind(cmd, "write to file (default: rfc-<slug>.md)")
	return cmd
}
