package cmd

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/jtprogru/srekit/internal/config"
)

// runSplit runs the CLI with stdout and stderr captured separately. runCLI's
// shared buffer is fine for generators, but doctor returns an error on a
// failing run, and cobra prints that to stderr — which would interleave into
// the JSON document the --json assertions parse.
func runSplit(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

// isolateEnv points every location doctor inspects at a fresh temp tree so a
// developer's real ~/.srekit.yaml, templates directory and git identity can't
// leak into the assertions. It sets environment variables and resets the
// process-wide config, so tests using it are not parallel-safe — the same
// constraint withConfig carries.
func isolateEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(dir, "empty-gitconfig"))
	unsetEnv(t, "SREKIT_TEMPLATES_DIR")
	unsetEnv(t, "SREKIT_AUTHOR")
	unsetEnv(t, "SREKIT_EMAIL")
	unsetEnv(t, "SREKIT_FULL_NAME")
	config.Reset()
	t.Cleanup(config.Reset)
	return dir
}

// unsetEnv removes a variable for the duration of the test. t.Setenv registers
// the restore hook; os.Unsetenv then clears the value it just set, which is
// the only way to get an *absent* variable back after the test.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

// withIdentity supplies an author through the environment so identity-neutral
// tests aren't at the mercy of whether the machine has a git identity.
func withIdentity(t *testing.T) {
	t.Helper()
	t.Setenv("SREKIT_AUTHOR", "Test Person")
	t.Setenv("SREKIT_EMAIL", "test@example.com")
}

func parseDoctorJSON(t *testing.T, out string) doctorReport {
	t.Helper()
	var report doctorReport
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	return report
}

func findingByID(t *testing.T, report doctorReport, id string) finding {
	t.Helper()
	for _, f := range report.Checks {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("no finding with id %q in %v", id, report.Checks)
	return finding{}
}

// TestDoctorFreshInstall covers the documented default: no config file, no
// templates directory, git installed. Their absence is the default, not a
// defect, so every check must be ok and the run must exit 0.
func TestDoctorFreshInstall(t *testing.T) {
	isolateEnv(t)
	withIdentity(t)

	out, _, err := runSplit(t, "doctor")
	if err != nil {
		t.Fatalf("doctor on a fresh install should exit 0, got: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 warn, 0 error") {
		t.Fatalf("expected a clean summary line, got:\n%s", out)
	}
	for _, header := range []string{"CONFIG", "TEMPLATES", "DEPENDENCIES"} {
		if !strings.Contains(out, header) {
			t.Errorf("findings should be grouped under %s:\n%s", header, out)
		}
	}
	if strings.Contains(out, "warn ") || strings.Contains(out, "error ") {
		t.Errorf("no finding should need attention:\n%s", out)
	}
}

// TestDoctorQuietOnHealthyEnvIsSilent — silence means healthy, which is the
// property that makes doctor usable as a CI step.
func TestDoctorQuietOnHealthyEnvIsSilent(t *testing.T) {
	isolateEnv(t)
	withIdentity(t)

	out, _, err := runSplit(t, "doctor", "--quiet")
	if err != nil {
		t.Fatalf("expected exit 0, got %v", err)
	}
	if out != "" {
		t.Fatalf("--quiet on a healthy environment should print nothing, got:\n%s", out)
	}
}

// TestDoctorWarningsDoNotFailTheRun — a configured-but-missing templates
// directory is advisory: generation still works from the embedded set.
func TestDoctorWarningsDoNotFailTheRun(t *testing.T) {
	dir := isolateEnv(t)
	withIdentity(t)
	missing := filepath.Join(dir, "no-such-templates")
	t.Setenv("SREKIT_TEMPLATES_DIR", missing)

	out, errOut, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("warnings must not fail the run, got: %v", err)
	}
	report := parseDoctorJSON(t, out)
	if report.Status != statusWarn {
		t.Fatalf("overall status should be warn, got %q", report.Status)
	}
	f := findingByID(t, report, "config.templates-dir")
	if f.Status != statusWarn {
		t.Errorf("config.templates-dir should warn, got %q", f.Status)
	}
	if !strings.Contains(f.Summary, missing) || !strings.Contains(f.Summary, "SREKIT_TEMPLATES_DIR") {
		t.Errorf("finding should name the path and its source: %q", f.Summary)
	}
	if !strings.Contains(f.Summary, "embedded") {
		t.Errorf("finding should state the fallback to embedded templates: %q", f.Summary)
	}
	if f.Remedy == "" {
		t.Error("a warn finding must carry a remedy")
	}
	// configureTemplates' own stderr fallback line would report the same
	// problem a second time, with less information.
	if strings.Contains(errOut, "falling back to embedded") {
		t.Errorf("the stderr fallback warning should be suppressed for doctor, got:\n%s", errOut)
	}
}

// TestDoctorErrorFailsTheRun pairs with the warn case above: an unparseable
// user artifact means the generator behind it cannot render at all.
func TestDoctorErrorFailsTheRun(t *testing.T) {
	dir := isolateEnv(t)
	withIdentity(t)
	templates := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templates, 0o755); err != nil {
		t.Fatal(err)
	}
	broken := "version: 1\nsections:\n  - id: dup\n    title: A\n    type: text\n  - id: dup\n    title: B\n    type: text\n"
	if err := os.WriteFile(filepath.Join(templates, "slo.yaml"), []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "legacy.md.tmpl"), []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SREKIT_TEMPLATES_DIR", templates)

	out, _, err := runSplit(t, "doctor", "--json")
	if err == nil {
		t.Fatal("an error finding must fail the run")
	}
	report := parseDoctorJSON(t, out)
	if report.Status != statusError {
		t.Fatalf("overall status should be error, got %q", report.Status)
	}

	parse := findingByID(t, report, "templates.parse")
	if parse.Status != statusError {
		t.Errorf("templates.parse should be error, got %q", parse.Status)
	}
	if !strings.Contains(parse.Summary, "slo.yaml") || !strings.Contains(parse.Summary, "duplicate id") {
		t.Errorf("finding should name the file and its parse error: %q", parse.Summary)
	}

	legacy := findingByID(t, report, "templates.legacy")
	if legacy.Status != statusWarn {
		t.Errorf("templates.legacy should be warn, got %q", legacy.Status)
	}
	if !strings.Contains(legacy.Summary, "legacy.md.tmpl") {
		t.Errorf("finding should name the legacy file: %q", legacy.Summary)
	}
	if !strings.Contains(legacy.Remedy, "templates migrate") {
		t.Errorf("remedy should point at templates migrate: %q", legacy.Remedy)
	}

	drift := findingByID(t, report, "templates.drift")
	if drift.Status != statusWarn {
		t.Errorf("templates.drift should be warn, got %q", drift.Status)
	}
	if !strings.Contains(drift.Remedy, "templates upgrade") {
		t.Errorf("remedy should point at templates upgrade: %q", drift.Remedy)
	}

	// --quiet still has to show the problems.
	quiet, _, err := runSplit(t, "doctor", "--quiet")
	if err == nil {
		t.Fatal("--quiet must not change the exit status")
	}
	if !strings.Contains(quiet, "templates.parse") {
		t.Errorf("--quiet should still print error findings, got:\n%s", quiet)
	}
	if strings.Contains(quiet, "checks:") {
		t.Errorf("--quiet should drop the summary line, got:\n%s", quiet)
	}
}

// TestDoctorReportsShadowedConfig asserts both halves of the shadowing
// contract: doctor warns, and the file it names as the winner is the one a
// generator actually reads.
func TestDoctorReportsShadowedConfig(t *testing.T) {
	home := isolateEnv(t)
	legacy := filepath.Join(home, ".srekit.yaml")
	xdg := filepath.Join(home, ".config", "srekit", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(xdg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("author: Legacy Person\nemail: legacy@example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(xdg, []byte("author: XDG Person\nemail: xdg@example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// initConfig is what Execute runs before any command; without it the
	// process-wide config is empty and no command — doctor included — would
	// see either file.
	initConfig("")

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("shadowing is advisory and must not fail the run: %v", err)
	}
	report := parseDoctorJSON(t, out)
	f := findingByID(t, report, "config.shadowed")
	if f.Status != statusWarn {
		t.Fatalf("config.shadowed should warn, got %q", f.Status)
	}
	if !strings.Contains(f.Summary, legacy) || !strings.Contains(f.Summary, xdg) {
		t.Errorf("finding should name both paths: %q", f.Summary)
	}
	if !strings.Contains(f.Summary, "never read") {
		t.Errorf("finding should say which file is ignored: %q", f.Summary)
	}

	// The reported winner must be the file the CLI really loads: a generator
	// reads whatever initConfig left in the process-wide config.
	body, err := runCLI(t, "rfc", "--title", "Shadowing", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Legacy Person") {
		t.Fatalf("doctor names %s as the file in effect, but the generator used something else:\n%s", legacy, body)
	}
}

// TestDoctorNoShadowingWithOneConfigFile — exactly one location present is
// the normal case and must stay quiet.
func TestDoctorNoShadowingWithOneConfigFile(t *testing.T) {
	home := isolateEnv(t)
	withIdentity(t)
	legacy := filepath.Join(home, ".srekit.yaml")
	if err := os.WriteFile(legacy, []byte("author: Only One\nemail: one@example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if f := findingByID(t, parseDoctorJSON(t, out), "config.shadowed"); f.Status != statusOK {
		t.Fatalf("one config location should not warn: %q — %s", f.Status, f.Summary)
	}
}

// TestDoctorMalformedConfigIsAWarning — a malformed config never fails a
// command that needs nothing from it, so doctor reports it without failing
// either. It is also the only place the swallowed load error surfaces.
func TestDoctorMalformedConfigIsAWarning(t *testing.T) {
	home := isolateEnv(t)
	withIdentity(t)
	cfg := filepath.Join(home, ".srekit.yaml")
	if err := os.WriteFile(cfg, []byte("author: [unclosed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("a malformed config must not fail the run: %v", err)
	}
	f := findingByID(t, parseDoctorJSON(t, out), "config.parse")
	if f.Status != statusWarn {
		t.Fatalf("config.parse should warn, got %q", f.Status)
	}
	if !strings.Contains(f.Summary, cfg) {
		t.Errorf("finding should name the file: %q", f.Summary)
	}
}

// TestDoctorMissingIdentityIsAnError — every generator that stamps an author
// fails in this environment, which is what error means.
func TestDoctorMissingIdentityIsAnError(t *testing.T) {
	isolateEnv(t)

	out, _, err := runSplit(t, "doctor", "--json")
	if err == nil {
		t.Fatal("an unresolvable author must fail the run")
	}
	f := findingByID(t, parseDoctorJSON(t, out), "config.identity")
	if f.Status != statusError {
		t.Fatalf("config.identity should be error, got %q: %s", f.Status, f.Summary)
	}
	for _, want := range []string{"config init", "--author", "user.name"} {
		if !strings.Contains(f.Remedy, want) {
			t.Errorf("remedy should mention %q, got %q", want, f.Remedy)
		}
	}
}

// TestDoctorEnvOverridesAreVisible — an override that changes what every
// generator stamps should not be invisible.
func TestDoctorEnvOverridesAreVisible(t *testing.T) {
	isolateEnv(t)
	withIdentity(t)

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatal(err)
	}
	report := parseDoctorJSON(t, out)
	if f := findingByID(t, report, "config.env"); !strings.Contains(f.Summary, "SREKIT_AUTHOR") {
		t.Errorf("config.env should name SREKIT_AUTHOR: %q", f.Summary)
	}
	if f := findingByID(t, report, "config.identity"); !strings.Contains(f.Summary, "SREKIT_AUTHOR") {
		t.Errorf("config.identity should name the source of the value: %q", f.Summary)
	}
}

// TestDoctorGitAbsent points PATH at an empty directory rather than assuming
// anything about the machine's git.
func TestDoctorGitAbsent(t *testing.T) {
	dir := isolateEnv(t)
	withIdentity(t)
	empty := filepath.Join(dir, "empty-path")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", empty)

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("an absent git is advisory, not fatal: %v", err)
	}
	f := findingByID(t, parseDoctorJSON(t, out), "dependencies.git")
	if f.Status != statusWarn {
		t.Fatalf("dependencies.git should warn, got %q: %s", f.Status, f.Summary)
	}
	if !strings.Contains(f.Summary, "--author") {
		t.Errorf("finding should state what falls back to flags and config: %q", f.Summary)
	}
}

// TestDoctorGitPresent covers the other branch — path and version reported.
func TestDoctorGitPresent(t *testing.T) {
	isolateEnv(t)
	withIdentity(t)

	out, _, err := runSplit(t, "doctor", "--json")
	if err != nil {
		t.Fatal(err)
	}
	f := findingByID(t, parseDoctorJSON(t, out), "dependencies.git")
	if f.Status != statusOK {
		t.Skipf("git is not available on this machine: %s", f.Summary)
	}
	if !strings.Contains(f.Summary, "git version") {
		t.Errorf("finding should report the version git prints: %q", f.Summary)
	}
}

// TestDoctorJSONShape covers the machine-readable contract: one indented
// document on stdout, camelCase keys, overall status equal to the worst
// finding, and --quiet leaving the document complete.
func TestDoctorJSONShape(t *testing.T) {
	isolateEnv(t)

	out, _, err := runSplit(t, "doctor", "--json")
	if err == nil {
		t.Fatal("expected the missing identity to fail the run")
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("the JSON document should be newline-terminated")
	}
	if !strings.Contains(out, "\n  \"checks\"") {
		t.Errorf("the JSON document should be indented:\n%s", out)
	}
	if strings.Contains(out, "CONFIG") {
		t.Errorf("--json must not also print the table:\n%s", out)
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	assertCamelCaseKeys(t, raw)

	report := parseDoctorJSON(t, out)
	if report.Status != statusError {
		t.Errorf("overall status should be the worst finding, got %q", report.Status)
	}
	if len(report.Checks) != len(doctorChecks()) {
		t.Errorf("expected %d findings, got %d", len(doctorChecks()), len(report.Checks))
	}

	// --quiet is a text-output preference; the data document stays complete.
	quiet, _, _ := runSplit(t, "doctor", "--json", "--quiet")
	quietReport := parseDoctorJSON(t, quiet)
	if len(quietReport.Checks) != len(report.Checks) {
		t.Fatalf("--json --quiet dropped findings: %d vs %d", len(quietReport.Checks), len(report.Checks))
	}
	if countStatus(quietReport.Checks, statusOK) == 0 {
		t.Error("--json --quiet should still carry the ok findings")
	}
}

var camelCaseKey = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)

func assertCamelCaseKeys(t *testing.T, v any) {
	t.Helper()
	switch node := v.(type) {
	case map[string]any:
		for k, child := range node {
			if !camelCaseKey.MatchString(k) {
				t.Errorf("JSON key %q is not camelCase", k)
			}
			assertCamelCaseKeys(t, child)
		}
	case []any:
		for _, child := range node {
			assertCamelCaseKeys(t, child)
		}
	}
}

// TestDoctorFindingsAreDeterministic — CI diffs two runs, so the order must
// not depend on map iteration anywhere in the chain.
func TestDoctorFindingsAreDeterministic(t *testing.T) {
	isolateEnv(t)
	withIdentity(t)

	first, _, err := runSplit(t, "doctor")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := runSplit(t, "doctor")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("two runs against an unchanged environment disagree:\n%s\n---\n%s", first, second)
	}
}

// TestDoctorWritesNothing — a diagnostic that mutates is a diagnostic you
// stop trusting. Nothing may appear on disk, including the config file and
// templates directory whose absence doctor reports.
func TestDoctorWritesNothing(t *testing.T) {
	home := isolateEnv(t)
	withIdentity(t)

	before := treeSnapshot(t, home)
	if _, _, err := runSplit(t, "doctor"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runSplit(t, "doctor", "--json"); err != nil {
		t.Fatal(err)
	}
	after := treeSnapshot(t, home)
	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Fatalf("doctor changed the filesystem:\nbefore: %v\nafter:  %v", before, after)
	}
}

func treeSnapshot(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	return paths
}

// TestDoctorIsDiscoverable and the tests below touch no global state, so they
// run in parallel with the rest of the suite.
func TestDoctorIsDiscoverable(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "doctor") {
		t.Fatalf("doctor should be listed in the root help:\n%s", out)
	}
}

func TestDoctorRejectsPositionalArgs(t *testing.T) {
	t.Parallel()
	if _, err := runCLI(t, "doctor", "extra-arg"); err == nil {
		t.Fatal("doctor takes no positional arguments")
	}
}

// TestDoctorHasNoWriteFlags — doctor writes nothing, and a flag a command
// would ignore must not exist.
func TestDoctorHasNoWriteFlags(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "doctor", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"--out", "--stdout", "--force", "--dry-run"} {
		if strings.Contains(out, flag) {
			t.Errorf("%s must not exist on doctor:\n%s", flag, out)
		}
	}
	if !strings.Contains(out, "--json") {
		t.Errorf("--json should be listed:\n%s", out)
	}
}

// TestDoctorTextIsPlainWhenPiped — the status has to be legible as a word,
// because that is what a redirected run and a log both see.
func TestDoctorTextIsPlainWhenPiped(t *testing.T) {
	t.Parallel()
	findings := []finding{
		{ID: "config.file", Category: categoryConfig, Status: statusOK, Summary: "fine"},
		{ID: "templates.parse", Category: categoryTemplates, Status: statusError, Summary: "broken", Remedy: "fix it"},
	}

	var plain bytes.Buffer
	renderFindings(&plain, findings, false, false)
	if strings.Contains(plain.String(), "\x1b[") {
		t.Errorf("piped output must carry no ANSI escapes:\n%q", plain.String())
	}
	for _, want := range []string{"ok", "error", "fix: fix it", "2 checks: 1 ok, 0 warn, 1 error"} {
		if !strings.Contains(plain.String(), want) {
			t.Errorf("expected %q in:\n%s", want, plain.String())
		}
	}

	var colored bytes.Buffer
	renderFindings(&colored, findings, false, true)
	if !strings.Contains(colored.String(), "\x1b[") {
		t.Error("the colorized branch should emit escapes")
	}
}

// TestUseColorHonorsNoColor — NO_COLOR wins over any terminal detection,
// matching templates diff.
func TestUseColorHonorsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if useColor(os.Stdout) {
		t.Error("NO_COLOR=1 must suppress color")
	}
}
