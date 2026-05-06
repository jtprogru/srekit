package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	licType    string
	licAuthor  string
	licEmail   string
	licYear    int
	licOut     string
	licStdout  bool
	licForce   bool
	licDry     bool
)

var licenseTemplates = map[string]string{
	"wtfpl":   "license_wtfpl.tmpl",
	"mit":     "license_mit.tmpl",
	"apache2": "license_apache2.tmpl",
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Generate a LICENSE file (default: WTFPL)",
	Long:  "Generate a LICENSE for your project. Default: WTFPL (matches the gch lic command). Supports mit and apache2 as alternatives.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		tmplName, ok := licenseTemplates[licType]
		if !ok {
			return fmt.Errorf("unknown license type %q (supported: wtfpl, mit, apache2)", licType)
		}
		author, err := meta.Resolve(licAuthor, licEmail)
		if err != nil {
			return err
		}
		year := licYear
		if year == 0 {
			year = time.Now().Year()
		}
		stdout := licStdout || licOut == ""
		return render.Render(cmd.OutOrStdout(), tmplName, struct {
			Author meta.Author
			Year   int
		}{author, year}, render.Options{
			Out:    licOut,
			Stdout: stdout,
			Force:  licForce,
			DryRun: licDry,
		})
	},
}

var licAlias = &cobra.Command{
	Use:    "lic",
	Short:  "Alias of `license` (kept for gch users)",
	Hidden: true,
	RunE:   licenseCmd.RunE,
}

func init() {
	for _, c := range []*cobra.Command{licenseCmd, licAlias} {
		c.Flags().StringVar(&licType, "type", "wtfpl", "license type: wtfpl | mit | apache2")
		c.Flags().StringVar(&licAuthor, "author", "", "author full name")
		c.Flags().StringVar(&licEmail, "email", "", "author email")
		c.Flags().IntVar(&licYear, "year", 0, "copyright year (default: current year)")
		c.Flags().StringVar(&licOut, "out", "", "write to file (default: stdout)")
		c.Flags().BoolVar(&licStdout, "stdout", false, "force stdout output")
		c.Flags().BoolVar(&licForce, "force", false, "overwrite existing file")
		c.Flags().BoolVar(&licDry, "dry-run", false, "print result, do not write a file")
	}
	rootCmd.AddCommand(licenseCmd)
	rootCmd.AddCommand(licAlias)
}
