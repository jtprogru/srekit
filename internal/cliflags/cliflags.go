// Package cliflags wires the standard srekit output flags
// (--out / --stdout / --force / --dry-run) onto a cobra command.
package cliflags

import (
	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/render"
)

// Output collects the output-related flags shared by every generator
// (--out / --stdout / --force / --dry-run / --json).
type Output struct {
	Out    string
	Stdout bool
	Force  bool
	DryRun bool
	JSON   bool
}

// Bind registers the flags on cmd. outDesc overrides the --out help text
// so each command can document its own default (e.g. "default: CHANGELOG.md").
//
// There is deliberately no --template FILE flag, here or anywhere else.
// Every generator resolves its artifact by name through the template
// source chain, so a template-file flag would be silently ignored — and a
// silently-ignored CLI flag is worse than a missing one. `license` was the
// last command that honored it; both went in v0.30.0. Per-artifact
// customization is a <name>.yaml in templates_dir.
func (o *Output) Bind(cmd *cobra.Command, outDesc string) {
	if outDesc == "" {
		outDesc = "write to file"
	}
	cmd.Flags().StringVar(&o.Out, "out", "", outDesc)
	cmd.Flags().BoolVar(&o.Stdout, "stdout", false, "print to stdout")
	cmd.Flags().BoolVar(&o.Force, "force", false, "overwrite existing file")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "print result, do not write a file")
	cmd.Flags().BoolVar(&o.JSON, "json", false, "emit the template data as JSON instead of rendering the template (default sink: stdout)")
}

// Edit is the narrower bundle for a command that rewrites a document the
// user already owns, rather than generating a new one.
//
// It deliberately omits --out and --force. The command's destination is the
// file it was pointed at, so a second destination has no meaning; and an
// overwrite guard would guard against the command's own purpose, since
// writing back to the source is what the command is for. --dry-run,
// --stdout and --json keep their usual meanings: show the result, don't
// write it.
type Edit struct {
	Stdout bool
	DryRun bool
	JSON   bool
}

// Bind registers the editing flags on cmd. jsonDesc documents what the
// command's --json emits, which differs from a generator's (an editing
// command has no template data to dump).
func (e *Edit) Bind(cmd *cobra.Command, jsonDesc string) {
	if jsonDesc == "" {
		jsonDesc = "emit the parsed document as JSON instead of writing"
	}
	cmd.Flags().BoolVar(&e.Stdout, "stdout", false, "print the result to stdout, do not write the file")
	cmd.Flags().BoolVar(&e.DryRun, "dry-run", false, "print the result, do not write the file")
	cmd.Flags().BoolVar(&e.JSON, "json", false, jsonDesc)
}

// RenderOptions converts the flag state into render.Options. defaultPath is
// the file path used when the user passed neither --out nor --stdout. The cmd
// is used to read the inherited persistent --quiet flag.
//
// There used to be a second constructor, RenderOptionsStructured, whose only
// job was to clear a BootstrapJSON flag that wrapped rendered markdown into a
// synthetic {meta, sections} envelope. Every generator has owned its own
// {meta, sections} payload since v0.20.0, so nothing set the flag; both it and
// the second constructor went in v0.30.0.
func (o *Output) RenderOptions(cmd *cobra.Command, defaultPath string) render.Options {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return render.Options{
		Out:     o.Out,
		Stdout:  o.Stdout,
		Force:   o.Force,
		DryRun:  o.DryRun,
		Default: defaultPath,
		JSON:    o.JSON,
		Quiet:   quiet,
	}
}
