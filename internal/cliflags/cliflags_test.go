package cliflags

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBindRegistersFlags(t *testing.T) {
	t.Parallel()
	var o Output
	cmd := &cobra.Command{Use: "x"}
	o.Bind(cmd, "")
	for _, name := range []string{"out", "stdout", "force", "dry-run"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not registered", name)
		}
	}
}

func TestBindCustomOutDesc(t *testing.T) {
	t.Parallel()
	var o Output
	cmd := &cobra.Command{Use: "x"}
	o.Bind(cmd, "write to file (default: foo.md)")
	if got := cmd.Flags().Lookup("out").Usage; got != "write to file (default: foo.md)" {
		t.Errorf("custom --out usage not applied: %q", got)
	}
}

func TestRenderOptionsCopiesState(t *testing.T) {
	t.Parallel()
	o := &Output{Out: "x.md", Stdout: true, Force: true, DryRun: true}
	cmd := &cobra.Command{Use: "x"}
	cmd.Flags().Bool("quiet", true, "")
	opts := o.RenderOptions(cmd, "def.md")
	if opts.Out != "x.md" || !opts.Stdout || !opts.Force || !opts.DryRun || opts.Default != "def.md" {
		t.Errorf("RenderOptions mismatch: %+v", opts)
	}
	if !opts.Quiet {
		t.Errorf("RenderOptions did not pick up --quiet: %+v", opts)
	}
}
