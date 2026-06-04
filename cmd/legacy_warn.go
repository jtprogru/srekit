package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/tmpl"
)

// warnStaleLegacyFiles checks every DirSource in loader for pre-v1.0
// legacy artifact files (the `.tmpl` / `.sections.yaml` layouts retired
// across the v0.14–v0.20 YAML-first migration) that are superseded by a
// v1 `<name>.yaml` shipped in embed. Prints one stderr WARN line per
// stale file found, pointing the user at the v1 file they should
// migrate into. Suppressed by --quiet.
//
// Called by every generator command after a successful artifact load
// (loader.LoadArtifactBytes) — by that point we know the v1 file
// resolved; anything matching legacyNames in user dir is now dead weight.
func warnStaleLegacyFiles(cmd *cobra.Command, loader *tmpl.Loader, v1Name string, legacyNames ...string) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		return
	}
	for _, src := range loader.Sources {
		ds, ok := src.(tmpl.DirSource)
		if !ok {
			continue
		}
		for _, name := range legacyNames {
			if _, err := os.Stat(filepath.Join(ds.Dir, name)); err == nil {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"WARN: %s in %s is a legacy pre-v1.0 format and is being ignored. "+
						"Run 'srekit templates upgrade' to scaffold %s, "+
						"then move your customizations into it.\n",
					name, ds.Dir, v1Name)
			}
		}
	}
}
