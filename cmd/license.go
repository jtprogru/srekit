package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var licenseTemplates = map[string]string{
	"wtfpl":   "license_wtfpl.tmpl",
	"mit":     "license_mit.tmpl",
	"apache2": "license_apache2.tmpl",
}

func newLicenseCmd() *cobra.Command {
	var (
		licType   string
		licAuthor string
		licEmail  string
		licYear   int
		out       cliflags.Output
	)
	cmd := &cobra.Command{
		Use:     "license",
		Aliases: []string{"lic"},
		Short:   "Generate a LICENSE file (default: WTFPL)",
		Long:    "Generate a LICENSE for your project. Default: WTFPL (matches the gch lic command). Supports mit and apache2 as alternatives.",
		Example: `  # MIT LICENSE to stdout (author/email from git config)
  srekit license --type mit

  # Apache-2.0 written to ./LICENSE
  srekit license --type apache2 --out LICENSE`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tmplName, ok := licenseTemplates[licType]
			if !ok {
				return fmt.Errorf("unknown license type %q (supported: wtfpl, mit, apache2)", licType)
			}
			author, err := meta.Resolve(viper.GetViper(), licAuthor, licEmail)
			if err != nil {
				return err
			}
			year := licYear
			if year == 0 {
				year = clock.Now().Year()
			}
			opts := out.RenderOptions(cmd, "")
			if out.Out == "" && !out.Stdout {
				opts.Stdout = true
			}
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), tmplName, struct {
				Author meta.Author `json:"author"`
				Year   int         `json:"year"`
			}{author, year}, opts)
		},
	}
	cmd.Flags().StringVar(&licType, "type", "wtfpl", "license type: wtfpl | mit | apache2")
	cmd.Flags().StringVar(&licAuthor, "author", "", "author full name")
	cmd.Flags().StringVar(&licEmail, "email", "", "author email")
	cmd.Flags().IntVar(&licYear, "year", 0, "copyright year (default: current year)")
	out.Bind(cmd, "write to file (default: stdout)")
	return cmd
}
