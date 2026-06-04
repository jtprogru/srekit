// Package cliflags wires the standard srekit output flags
// (--out / --stdout / --force / --dry-run) onto a cobra command.
package cliflags

import (
	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/render"
)

// Output collects the output-related flags plus --template (single-template
// override). The helper exists only to avoid duplicating the StringVar/BoolVar
// calls in every cmd file. The name is kept for back-compat; consider it the
// shared per-command flag bundle.
type Output struct {
	Out          string
	Stdout       bool
	Force        bool
	DryRun       bool
	TemplatePath string
	JSON         bool
}

// Bind registers the flags on cmd. outDesc overrides the --out help text
// so each command can document its own default (e.g. "default: CHANGELOG.md").
func (o *Output) Bind(cmd *cobra.Command, outDesc string) {
	if outDesc == "" {
		outDesc = "write to file"
	}
	cmd.Flags().StringVar(&o.Out, "out", "", outDesc)
	cmd.Flags().BoolVar(&o.Stdout, "stdout", false, "print to stdout")
	cmd.Flags().BoolVar(&o.Force, "force", false, "overwrite existing file")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "print result, do not write a file")
	cmd.Flags().StringVar(&o.TemplatePath, "template", "", "use this template file instead of the embedded/templates-dir template")
	cmd.Flags().BoolVar(&o.JSON, "json", false, "emit the template data as JSON instead of rendering the template (default sink: stdout)")
}

// RenderOptions converts the flag state into render.Options. defaultPath is
// the file path used when the user passed neither --out nor --stdout. The cmd
// is used to read the inherited persistent --quiet flag.
//
// This is the default for legacy commands that have not migrated to a
// sections manifest: BootstrapJSON is true, so --json wraps the rendered
// markdown into the uniform `{meta, sections:[{id:"body", ...}]}` envelope.
// Commands that own their {meta, sections} payload (postmortem) call
// RenderOptionsStructured instead.
func (o *Output) RenderOptions(cmd *cobra.Command, defaultPath string) render.Options {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return render.Options{
		Out:           o.Out,
		Stdout:        o.Stdout,
		Force:         o.Force,
		DryRun:        o.DryRun,
		Default:       defaultPath,
		TemplatePath:  o.TemplatePath,
		JSON:          o.JSON,
		Quiet:         quiet,
		BootstrapJSON: true,
	}
}

// RenderOptionsStructured is for commands that build a {meta, sections}
// payload themselves (e.g. postmortem with its sidecar manifest). It
// returns the same Options as RenderOptions but with BootstrapJSON cleared
// so --json short-circuits straight to MarshalIndent without re-wrapping.
func (o *Output) RenderOptionsStructured(cmd *cobra.Command, defaultPath string) render.Options {
	opts := o.RenderOptions(cmd, defaultPath)
	opts.BootstrapJSON = false
	return opts
}
