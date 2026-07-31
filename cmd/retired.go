package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// Commands retired in v0.30.0. They stay registered as hidden stubs so a
// user whose CI still runs `srekit capacity --service x` gets an answer
// instead of cobra's generic `unknown command "capacity" for "srekit"`.
//
// The stubs are a courtesy with an expiry date: they are dropped at 1.0,
// when the retired names become ordinary unknown commands. Do not grow
// this list into a general deprecation mechanism — a name that has been
// gone for a whole major line does not need a tombstone.
//
// RELEASE CHECKLIST: retiredIn is the version users are told to pin
// *below*, and it is hard-coded here and in ~20 doc/spec references. If
// the release carrying this removal is cut as anything other than the
// value below, update it here and grep the tree for the old string —
// otherwise every error message names a release that does not contain
// the removal.
const retiredIn = "v0.30.0"

const retiredDocsURL = "https://jtprogru.github.io/srekit/migration/removed-commands/"

type retiredCmd struct {
	name    string
	aliases []string
	// reason is the second sentence of the error: why the command left,
	// and what to reach for instead.
	reason string
}

// message is the one thing a retired name has to say, used for both the
// RunE error and the help text. `srekit help capacity` reaches cobra's
// help command rather than RunE, and cobra exits 0 there even for an
// entirely unknown topic — so without Short/Long the help path would
// print a bare usage block that reads like a working command.
func (r retiredCmd) message() string {
	return fmt.Sprintf("the %q command was removed in %s. %s See %s",
		r.name, retiredIn, r.reason, retiredDocsURL)
}

// retiredCmds builds the hidden stub commands for every retired name.
//
// DisableFlagParsing is deliberate: `srekit retro` with no --team must
// report the removal, not `--team is required`, and `srekit license
// --type mit` must not fail on an unknown flag before it gets the chance
// to explain itself. The stub never looks at its arguments — including
// `-h`, which is why `capacity --help` reports the removal too.
func retiredCmds() []*cobra.Command {
	retired := []retiredCmd{
		{
			name:   "capacity",
			reason: "Capacity planning is spreadsheet work, not a text artifact srekit is good at.",
		},
		{
			name:   "retro",
			reason: "Sprint retrospectives are an agile ceremony, not an SRE artifact.",
		},
		{
			name:    "license",
			aliases: []string{"lic"},
			reason:  "Use your code host's license picker, or copy the text once from choosealicense.com.",
		},
	}
	cmds := make([]*cobra.Command, 0, len(retired))
	for _, r := range retired {
		cmds = append(cmds, &cobra.Command{
			Use:                r.name,
			Aliases:            r.aliases,
			Short:              r.message(),
			Long:               r.message(),
			Hidden:             true,
			DisableFlagParsing: true,
			Args:               cobra.ArbitraryArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				return errors.New(r.message())
			},
		})
	}
	return cmds
}
