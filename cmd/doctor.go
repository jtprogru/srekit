package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Finding statuses. `error` means a generator will fail or produce wrong
// output in this environment; `warn` means it works but something is being
// ignored, outdated, or about to break; `ok` means there is nothing to do.
const (
	statusOK    = "ok"
	statusWarn  = "warn"
	statusError = "error"
)

// Finding categories. Every check belongs to exactly one, and findings are
// rendered grouped in this order.
const (
	categoryConfig       = "config"
	categoryTemplates    = "templates"
	categoryDependencies = "dependencies"
)

// annotationSelfReportsPaths marks a command that reports path-resolution
// problems itself, so configureTemplates suppresses its stderr fallback
// warning rather than double-reporting what a finding already carries.
const annotationSelfReportsPaths = "srekit.self-reports-paths"

// statusWidth is the display width of the longest status word ("error"),
// used to align the finding table and its remedy continuation lines.
const statusWidth = 5

// finding is one check's verdict. The JSON tags are the public shape of
// `srekit doctor --json`; ID is a stable identifier users grep for and gate
// CI on, so renaming one is a breaking change.
type finding struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Summary  string `json:"summary"`
	Remedy   string `json:"remedy,omitempty"`
}

// doctorReport is the top-level `--json` document: the worst status among
// the findings, then every finding in render order.
type doctorReport struct {
	Status string    `json:"status"`
	Checks []finding `json:"checks"`
}

// checkResult is what a check function returns. The identifier and category
// come from the descriptor, so a check cannot disagree with its own name.
type checkResult struct {
	status  string
	summary string
	remedy  string
}

// check is one diagnostic: a stable identifier, the category it renders
// under, and the function that inspects the environment. run returns a
// result rather than an error — a failure to inspect is itself a finding,
// which is what makes "one broken check never aborts the rest" structural.
type check struct {
	id       string
	category string
	run      func(*doctorEnv) checkResult
}

func newDoctorCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the srekit environment: config, templates, git",
		Long: `Report the state srekit resolves before it renders anything: which config
file is actually read (and whether a second one is being shadowed), where the
templates directory resolves and whether its artifacts still parse, whether an
author identity can be resolved at all, and whether git is on PATH.

doctor is read-only — it never creates, changes, or repairs anything, and it
makes no network request. Each check reports ok, warn, or error:

  ok     nothing to do
  warn   srekit works, but something is being ignored or is about to break
  error  a generator will fail or produce wrong output in this environment

Exit status is 1 when any check reports error, 0 otherwise — warnings never
fail the run, so doctor is safe to adopt in CI. --quiet prints only the
findings that need attention, so silence means healthy.`,
		Example: `  # Full report
  srekit doctor

  # Only what needs attention; prints nothing on a healthy machine
  srekit doctor --quiet

  # Gate CI on it
  srekit doctor --json | jq -e '.status != "error"'`,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{annotationSelfReportsPaths: "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit the findings as a JSON document")
	return cmd
}

func runDoctor(cmd *cobra.Command, jsonOut bool) error {
	findings := runChecks(newDoctorEnv(cmd), doctorChecks())
	out := cmd.OutOrStdout()

	if jsonOut {
		// --quiet does not filter here: JSON is a data document, and a
		// consumer that asked for the full structure and got a subset of it
		// has a bug on its hands rather than a preference.
		body, err := json.MarshalIndent(doctorReport{Status: worstStatus(findings), Checks: findings}, "", "  ")
		if err != nil {
			return fmt.Errorf("encode findings: %w", err)
		}
		fmt.Fprintln(out, string(body))
	} else {
		quiet, _ := cmd.Flags().GetBool("quiet")
		renderFindings(out, findings, quiet, useColor(out))
	}

	if n := countStatus(findings, statusError); n > 0 {
		return fmt.Errorf("%d check(s) reported an error", n)
	}
	return nil
}

// doctorChecks is the fixed check set, in render order: config and path
// resolution first, then template health, then external dependencies. The
// slice is built per invocation rather than registered into a package-level
// variable, so nothing is shared between command trees.
func doctorChecks() []check {
	return []check{
		{id: "config.file", category: categoryConfig, run: checkConfigFile},
		{id: "config.parse", category: categoryConfig, run: checkConfigParse},
		{id: "config.shadowed", category: categoryConfig, run: checkConfigShadowed},
		{id: "config.writable", category: categoryConfig, run: checkConfigWritable},
		{id: "config.env", category: categoryConfig, run: checkConfigEnv},
		{id: "config.templates-dir", category: categoryConfig, run: checkTemplatesDir},
		{id: "config.templates-shadowed", category: categoryConfig, run: checkTemplatesShadowed},
		{id: "config.identity", category: categoryConfig, run: checkIdentity},
		{id: "templates.parse", category: categoryTemplates, run: checkTemplatesParse},
		{id: "templates.legacy", category: categoryTemplates, run: checkTemplatesLegacy},
		{id: "templates.drift", category: categoryTemplates, run: checkTemplatesDrift},
		{id: "dependencies.git", category: categoryDependencies, run: checkGit},
	}
}

// runChecks walks every descriptor in order and collects its finding. There
// is deliberately no early exit and no error return: ordering is slice order
// (so two runs against an unchanged environment agree), and a check that
// cannot inspect its subject says so in its own finding.
func runChecks(env *doctorEnv, checks []check) []finding {
	findings := make([]finding, 0, len(checks))
	for _, c := range checks {
		res := c.run(env)
		findings = append(findings, finding{
			ID:       c.id,
			Category: c.category,
			Status:   res.status,
			Summary:  res.summary,
			Remedy:   res.remedy,
		})
	}
	return findings
}

// renderFindings prints the findings as an aligned table grouped by category,
// each remedy on its own continuation line, followed by the per-status counts.
// quiet drops the ok findings and the summary line, so a healthy environment
// prints nothing at all.
func renderFindings(w io.Writer, findings []finding, quiet, color bool) {
	shown := findings
	if quiet {
		shown = needingAttention(findings)
	}

	idWidth := 0
	for _, f := range shown {
		if len(f.ID) > idWidth {
			idWidth = len(f.ID)
		}
	}

	category := ""
	for _, f := range shown {
		if f.Category != category {
			if category != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintln(w, strings.ToUpper(f.Category))
			category = f.Category
		}
		// Pad before colorizing so the escape sequences don't count toward
		// the column width.
		status := fmt.Sprintf("%-*s", statusWidth, f.Status)
		if color {
			status = colorize(f.Status, status)
		}
		fmt.Fprintf(w, "  %s  %-*s  %s\n", status, idWidth, f.ID, f.Summary)
		if f.Remedy != "" {
			fmt.Fprintf(w, "  %s  %s  fix: %s\n",
				strings.Repeat(" ", statusWidth), strings.Repeat(" ", idWidth), f.Remedy)
		}
	}

	if quiet {
		return
	}
	if len(shown) > 0 {
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "%d checks: %d ok, %d warn, %d error\n",
		len(findings),
		countStatus(findings, statusOK),
		countStatus(findings, statusWarn),
		countStatus(findings, statusError))
}

// needingAttention keeps the warn and error findings in their original order.
func needingAttention(findings []finding) []finding {
	out := make([]finding, 0, len(findings))
	for _, f := range findings {
		if f.Status != statusOK {
			out = append(out, f)
		}
	}
	return out
}

func countStatus(findings []finding, status string) int {
	n := 0
	for _, f := range findings {
		if f.Status == status {
			n++
		}
	}
	return n
}

// worstStatus returns the most severe status among the findings, which is
// what the JSON document reports as its overall status.
func worstStatus(findings []finding) string {
	worst := statusOK
	for _, f := range findings {
		if statusRank(f.Status) > statusRank(worst) {
			worst = f.Status
		}
	}
	return worst
}

func statusRank(status string) int {
	switch status {
	case statusError:
		return 2
	case statusWarn:
		return 1
	default:
		return 0
	}
}

// useColor reports whether to colorize the status column: only when the
// destination is a terminal (so a redirected run stays free of escapes) and
// NO_COLOR does not ask us to stop, matching `templates diff`.
func useColor(w io.Writer) bool {
	if noColorEnv() {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func colorize(status, text string) string {
	code := "32"
	switch status {
	case statusError:
		code = "31"
	case statusWarn:
		code = "33"
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}
