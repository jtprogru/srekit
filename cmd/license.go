package cmd

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/config"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// License texts are inlined as constants in v0.14.0 (previously shipped
// as license_*.tmpl files in the embed FS). Rationale: license bodies are
// static (only Year + Author placeholders vary), users almost never
// customize them, and inlining removes one file class from the templates
// directory + embed pattern. Users who want a custom license body can
// still pass --template FILE.

const mitLicense = `MIT License

Copyright (c) {{ .Year }} {{ .Author.Name }} <{{ .Author.Email }}>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

const apache2License = `                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   Copyright {{ .Year }} {{ .Author.Name }} <{{ .Author.Email }}>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
`

const wtfplLicense = `            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
                    Version 2, December 2004

 Copyright (C) {{ .Year }} {{ .Author.Name }} <{{ .Author.Email }}>

 Everyone is permitted to copy and distribute verbatim or modified
 copies of this license document, and changing it is allowed as long
 as the name is changed.

            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
   TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION

  0. You just DO WHAT THE FUCK YOU WANT TO.
`

var licenseBodies = map[string]string{
	"wtfpl":   wtfplLicense,
	"mit":     mitLicense,
	"apache2": apache2License,
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
		Long:    "Generate a LICENSE for your project. Default: WTFPL (matches the gch lic command). Supports mit and apache2 as alternatives. License texts are inlined in the binary as of v0.14.0; use --template FILE for a custom body.",
		Example: `  # MIT LICENSE to stdout (author/email from git config)
  srekit license --type mit

  # Apache-2.0 written to ./LICENSE
  srekit license --type apache2 --out LICENSE`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, ok := licenseBodies[licType]
			if !ok {
				return fmt.Errorf("unknown license type %q (supported: wtfpl, mit, apache2)", licType)
			}
			author, err := meta.Resolve(config.Global(), licAuthor, licEmail)
			if err != nil {
				return err
			}
			year := licYear
			if year == 0 {
				year = clock.Now().Year()
			}
			data := struct {
				Author meta.Author `json:"author"`
				Year   int         `json:"year"`
			}{author, year}

			// Render via inline template; --template FILE still routes
			// through render.Render's TemplatePath path.
			opts := out.RenderOptions(cmd, "")
			if out.Out == "" && !out.Stdout {
				opts.Stdout = true
			}
			if opts.TemplatePath != "" {
				return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "license", data, opts)
			}
			rendered, err := executeLicenseTemplate(licType, body, data)
			if err != nil {
				return err
			}
			return render.WriteRaw(cmd.OutOrStdout(), rendered, opts)
		},
	}
	cmd.Flags().StringVar(&licType, "type", "wtfpl", "license type: wtfpl | mit | apache2")
	cmd.Flags().StringVar(&licAuthor, "author", "", "author full name")
	cmd.Flags().StringVar(&licEmail, "email", "", "author email")
	cmd.Flags().IntVar(&licYear, "year", 0, "copyright year (default: current year)")
	out.Bind(cmd, "write to file (default: stdout)")
	// license is the only command that honors --template (the v1 artifact
	// loader doesn't run for license, so the flag genuinely routes to a
	// custom body file via render.Render's TemplatePath branch).
	out.BindTemplateFlag(cmd)
	return cmd
}

// executeLicenseTemplate runs the inlined license body through the shared
// FuncMap so any Go-template helpers (default, now, etc.) work consistently
// with the other generators.
func executeLicenseTemplate(name, body string, data any) ([]byte, error) {
	t, err := template.New("license_" + name).Funcs(tmpl.Funcs).Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse license %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render license %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
