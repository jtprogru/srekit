// Package cliflags wires the standard srekit output flags
// (--out / --stdout / --force / --dry-run) onto a cobra command.
package cliflags

import (
	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/render"
)

// Output collects the four output-related flags shared by every srekit command.
// Keep it on the stack inside each command's constructor; the helper exists
// only to avoid duplicating the four StringVar/BoolVar calls in every file.
type Output struct {
	Out    string
	Stdout bool
	Force  bool
	DryRun bool
}

// Bind registers the four flags on cmd. outDesc overrides the --out help text
// so each command can document its own default (e.g. "default: CHANGELOG.md").
func (o *Output) Bind(cmd *cobra.Command, outDesc string) {
	if outDesc == "" {
		outDesc = "write to file"
	}
	cmd.Flags().StringVar(&o.Out, "out", "", outDesc)
	cmd.Flags().BoolVar(&o.Stdout, "stdout", false, "print to stdout")
	cmd.Flags().BoolVar(&o.Force, "force", false, "overwrite existing file")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "print result, do not write a file")
}

// RenderOptions converts the flag state into render.Options. defaultPath is
// the file path used when the user passed neither --out nor --stdout.
func (o *Output) RenderOptions(defaultPath string) render.Options {
	return render.Options{
		Out:     o.Out,
		Stdout:  o.Stdout,
		Force:   o.Force,
		DryRun:  o.DryRun,
		Default: defaultPath,
	}
}
