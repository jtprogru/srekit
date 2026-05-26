package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/tmpl"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// resetTmplDefault snapshots tmpl.Default and restores it after the test.
// Use in any test that exercises --templates-dir, which mutates the
// package-level loader. Not safe to use with t.Parallel().
func resetTmplDefault(t *testing.T) {
	t.Helper()
	orig := tmpl.Default
	t.Cleanup(func() { tmpl.Default = orig })
}

func TestTaskStdout(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "task", "--title", "Tail latency", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "# Расследование (Investigation) — Tail latency") {
		t.Fatalf("missing rendered title: %s", out)
	}
	if !strings.Contains(out, "tags:") {
		t.Fatalf("front matter missing")
	}
}

func TestTaskRequiresTitle(t *testing.T) {
	t.Parallel()
	_, err := runCLI(t, "task", "--title=", "--stdout")
	if err == nil {
		t.Fatal("expected error when --title is empty")
	}
}

// withViper temporarily seeds the global viper for tests that go through
// meta.Resolve(viper.GetViper(), ...). It is not parallel-safe.
func withViper(t *testing.T, kv map[string]string) {
	t.Helper()
	viper.Reset()
	for k, v := range kv {
		viper.Set(k, v)
	}
	t.Cleanup(viper.Reset)
}

func TestLicenseWTFPLDefault(t *testing.T) {
	withViper(t, map[string]string{"author": "Test Person", "email": "t@example.com"})

	out, err := runCLI(t, "license", "--stdout", "--year", "2026")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "DO WHAT THE FUCK YOU WANT") {
		t.Fatalf("WTFPL body missing: %s", out)
	}
	if !strings.Contains(out, "2026 Test Person <t@example.com>") {
		t.Fatalf("author/year not interpolated: %s", out)
	}
}

func TestLicenseMIT(t *testing.T) {
	withViper(t, map[string]string{"author": "Test Person", "email": "t@example.com"})

	out, err := runCLI(t, "license", "--type", "mit", "--stdout", "--year", "2026")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "MIT License") {
		t.Fatalf("MIT body missing: %s", out)
	}
}

func TestLicenseUnknownType(t *testing.T) {
	withViper(t, map[string]string{"author": "x", "email": "x@x"})
	_, err := runCLI(t, "license", "--type", "gpl", "--stdout")
	if err == nil {
		t.Fatal("expected error for unknown license type")
	}
}

func TestPostmortem(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "postmortem", "--title", "API outage", "--severity", "SEV-1", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Постмортем (Postmortem) — API outage") || !strings.Contains(out, "SEV-1") {
		t.Fatalf("postmortem body wrong: %s", out)
	}
}

func TestRFC(t *testing.T) {
	withViper(t, map[string]string{"author": "Test Person", "email": "t@example.com"})
	out, err := runCLI(t, "rfc", "--title", "Migrate to gRPC", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Migrate to gRPC") || !strings.Contains(out, "proposed") {
		t.Fatalf("rfc body wrong: %s", out)
	}
}

func TestRunbook(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "runbook", "--title", "p99 latency spike", "--service", "api-gw", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Рунбук (Runbook) — p99 latency spike") || !strings.Contains(out, "api-gw") {
		t.Fatalf("runbook body wrong: %s", out)
	}
}

func TestChangelog(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "changelog", "--stdout", "--repo", "jtprogru/srekit", "--version", "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Keep a Changelog") {
		t.Fatalf("changelog body wrong: %s", out)
	}
	if !strings.Contains(out, "github.com/jtprogru/srekit/compare/v0.1.0...HEAD") {
		t.Fatalf("repo not interpolated: %s", out)
	}
}

func TestOncallReport(t *testing.T) {
	withViper(t, map[string]string{"author": "Test Person", "email": "t@example.com"})
	out, err := runCLI(t, "oncall-report", "--team", "platform", "--start", "2026-05-04", "--end", "2026-05-10", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Отчёт по дежурству (On-call report) — platform") || !strings.Contains(out, "2026-05-04") {
		t.Fatalf("oncall body wrong: %s", out)
	}
}

func TestSLO(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "slo", "--service", "api-gw", "--target", "99.95%", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "SLO — api-gw") || !strings.Contains(out, "99.95%") {
		t.Fatalf("slo body wrong: %s", out)
	}
}

func TestRetro(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "retro", "--team", "platform", "--sprint", "2026-W19", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Ретро (Retro) — platform / 2026-W19") {
		t.Fatalf("retro body wrong: %s", out)
	}
}

// TestOncallSundayWeekBoundary pins the wall clock to Sunday 2026-05-10 and
// verifies the week is computed as Mon 2026-05-04 → Sun 2026-05-10. Regression
// for the bug where Sunday's Weekday()==0 caused the formula to roll to the
// following Monday instead.
func TestOncallSundayWeekBoundary(t *testing.T) {
	withViper(t, map[string]string{"author": "X", "email": "x@x"})
	orig := clock.Now
	t.Cleanup(func() { clock.Now = orig })
	clock.Now = func() time.Time { return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC) }

	out, err := runCLI(t, "oncall-report", "--team", "platform", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2026-05-04") || !strings.Contains(out, "2026-05-10") {
		t.Fatalf("Sunday should yield Mon 2026-05-04 → Sun 2026-05-10, got: %s", out)
	}
}

// TestChangelogMissingRepoFails locks in the #11 fix: when git remote
// detection fails and --repo is omitted, we error instead of writing
// "OWNER/REPO" into the file. The git mock returns an empty remote.
func TestChangelogMissingRepoFails(t *testing.T) {
	withViper(t, nil)
	dir := t.TempDir()
	// chdir into a non-git directory so DetectRepo's git invocation fails.
	t.Chdir(dir)

	_, err := runCLI(t, "changelog", "--stdout")
	if err == nil {
		t.Fatal("expected error when --repo is missing and remote detection fails")
	}
	if !strings.Contains(err.Error(), "could not detect repo") {
		t.Fatalf("error should mention repo detection, got: %v", err)
	}
}

func TestIncidentReport(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "incident", "--title", "API down", "--severity", "SEV-1", "--lead", "alice", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Инцидент (Incident) — API down") || !strings.Contains(out, "SEV-1") || !strings.Contains(out, "alice") {
		t.Fatalf("incident body wrong: %s", out)
	}
}

func TestIncidentInvalidStatus(t *testing.T) {
	t.Parallel()
	_, err := runCLI(t, "incident", "--title", "X", "--status", "broken", "--stdout")
	if err == nil {
		t.Fatal("expected error on invalid status")
	}
}

func TestErrorBudgetPolicy(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "ebp", "--service", "api-gw", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Политика бюджета ошибок (Error budget policy) — api-gw") {
		t.Fatalf("ebp body wrong: %s", out)
	}
}

func TestCapacityPlan(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "capacity", "--service", "api-gw", "--horizon", "6m", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "План ёмкости (Capacity plan) — api-gw") || !strings.Contains(out, "6m") {
		t.Fatalf("capacity body wrong: %s", out)
	}
}

// TestTemplatesDirOverride mutates the package-level tmpl.Default, so it
// must not be parallel and must reset Default in cleanup. Verifies that
// --templates-dir picks up custom templates from the given directory.
func TestTemplatesDirOverride(t *testing.T) {
	resetTmplDefault(t)

	dir := t.TempDir()
	custom := []byte("# CUSTOM POSTMORTEM: {{ .Title }}\n")
	if err := os.WriteFile(filepath.Join(dir, "postmortem.md.tmpl"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "--templates-dir", dir, "postmortem", "--title", "X", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "# CUSTOM POSTMORTEM: X") {
		t.Fatalf("expected custom template body, got: %s", out)
	}
}

// TestTemplatesDirPartialFallback verifies that templates not present in
// --templates-dir transparently fall back to the embedded versions.
func TestTemplatesDirPartialFallback(t *testing.T) {
	resetTmplDefault(t)

	dir := t.TempDir() // empty — nothing to override
	out, err := runCLI(t, "--templates-dir", dir, "rfc", "--title", "X", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "RFC-") {
		t.Fatalf("expected embedded rfc fallback, got: %s", out)
	}
}

// TestPerCommandTemplateOverride uses --template (single-file override) on
// a single command. Does not touch tmpl.Default — render reads the file
// directly via opts.TemplatePath — so this test is parallel-safe.
func TestPerCommandTemplateOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	custom := filepath.Join(dir, "my-runbook.tmpl")
	if err := os.WriteFile(custom, []byte("# ONESHOT RUNBOOK: {{ .Title }}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "runbook", "--title", "p99 spike", "--template", custom, "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "# ONESHOT RUNBOOK: p99 spike") {
		t.Fatalf("expected single-template override, got: %s", out)
	}
}

// TestTemplatesInitScaffolds verifies that 'srekit templates init <dir> --no-git'
// copies every embedded template and writes TEMPLATES.md.
func TestTemplatesInitScaffolds(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "templates")
	out, err := runCLI(t, "templates", "init", target, "--no-git")
	if err != nil {
		t.Fatalf("init failed: %v (output: %s)", err, out)
	}

	// At least the SRE-doc templates we ship must be there.
	for _, name := range []string{
		"task.md.tmpl", "incident.md.tmpl", "postmortem.md.tmpl",
		"runbook.md.tmpl", "rfc.md.tmpl", "slo.md.tmpl",
		"ebp.md.tmpl", "capacity.md.tmpl",
		"oncall.md.tmpl", "retro.md.tmpl", "changelog.md.tmpl",
		"TEMPLATES.md",
	} {
		if _, err := os.Stat(filepath.Join(target, name)); err != nil {
			t.Errorf("expected %s to be written: %v", name, err)
		}
	}
	if !strings.Contains(out, "Templates scaffolded in") {
		t.Errorf("expected friendly summary, got: %s", out)
	}
}

// TestTemplatesInitRefusesOverwrite locks the no-clobber default.
func TestTemplatesInitRefusesOverwrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), []byte("MINE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runCLI(t, "templates", "init", dir, "--no-git")
	if err == nil {
		t.Fatal("expected error when target contains existing template")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error should mention --force, got: %v", err)
	}
	// Verify the file wasn't clobbered.
	b, _ := os.ReadFile(filepath.Join(dir, "task.md.tmpl"))
	if string(b) != "MINE\n" {
		t.Fatalf("existing file was overwritten: %q", string(b))
	}
}

// TestTemplatesInitForce overrides --force on top of an existing file.
func TestTemplatesInitForce(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), []byte("MINE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "templates", "init", dir, "--no-git", "--force"); err != nil {
		t.Fatalf("init --force failed: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "task.md.tmpl"))
	if !strings.Contains(string(b), "Расследование") {
		t.Fatalf("--force should overwrite with embedded content; got: %s", string(b))
	}
}

func TestSretaskAlias(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "sretask", "--title", "alias works", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "alias works") {
		t.Fatalf("alias not working")
	}
}
