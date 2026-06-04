package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/clock"
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

// TestTemplatesDirOverride verifies that --templates-dir picks up custom
// templates from the given directory. The loader is now scoped per command
// tree (cmd.Context), so this is parallel-safe.
func TestTemplatesDirOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// As of v0.13.0 the postmortem template sees a {Meta, Sections} struct
	// rather than a flat one — custom templates reference .Meta.Title.
	custom := []byte("# CUSTOM POSTMORTEM: {{ .Meta.Title }}\n")
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
// --templates-dir transparently fall back to the embedded versions. Uses
// postmortem because it has no author/email dependency that CI runners lack.
func TestTemplatesDirPartialFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // empty — nothing to override
	out, err := runCLI(t, "--templates-dir", dir, "postmortem", "--title", "X", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Постмортем") {
		t.Fatalf("expected embedded postmortem fallback, got: %s", out)
	}
}

// TestPerCommandTemplateOverride uses --template (single-file override) on
// a single command. render reads the file directly via opts.TemplatePath,
// bypassing the loader entirely.
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

	// At least the SRE-doc templates we ship must be there, plus the
	// sidecar manifest that the postmortem command now consumes.
	for _, name := range []string{
		"task.md.tmpl", "incident.md.tmpl", "postmortem.md.tmpl",
		"runbook.md.tmpl", "rfc.md.tmpl", "slo.md.tmpl",
		"ebp.md.tmpl", "capacity.md.tmpl",
		"oncall.md.tmpl", "retro.md.tmpl", "changelog.md.tmpl",
		"postmortem.sections.yaml",
		"TEMPLATES.md",
	} {
		if _, err := os.Stat(filepath.Join(target, name)); err != nil {
			t.Errorf("expected %s to be written: %v", name, err)
		}
	}
	// Snapshot for the new sidecar must be seeded too so future 3-way
	// merges work without falling back to additive mode.
	if _, err := os.Stat(filepath.Join(target, ".srekit-embedded", "postmortem.sections.yaml")); err != nil {
		t.Errorf("expected snapshot for postmortem.sections.yaml: %v", err)
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

// TestTemplatesPullRejectsNonGitDir verifies the friendly error message when
// the configured templates directory exists but isn't a git repo.
func TestTemplatesPullRejectsNonGitDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // exists, but no .git/
	_, err := runCLI(t, "--templates-dir", dir, "templates", "pull")
	if err == nil {
		t.Fatal("expected error when dir is not a git repo")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("error should mention 'not a git repository', got: %v", err)
	}
}

// TestTemplatesPullSyncsFromRemote spins up a local bare git remote, clones
// it into a fake user-templates dir, pushes a change from a separate source
// repo, then runs 'srekit templates pull' to verify the change reaches the
// user dir. Skips if git is unavailable.
func TestTemplatesPullSyncsFromRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()

	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	src := filepath.Join(base, "src")
	user := filepath.Join(base, "user-templates")

	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=t@test",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=t@test",
	)
	runGit := func(args ...string) {
		t.Helper()
		c := exec.CommandContext(t.Context(), "git", args...)
		c.Env = gitEnv
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("init", "--bare", "--initial-branch=main", remote)
	runGit("init", "--initial-branch=main", src)
	if err := os.WriteFile(filepath.Join(src, "task.md.tmpl"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("-C", src, "add", ".")
	runGit("-C", src, "commit", "-m", "init")
	runGit("-C", src, "remote", "add", "origin", remote)
	runGit("-C", src, "push", "-u", "origin", "main")
	runGit("clone", remote, user)

	// Sanity-check the initial clone.
	if b, _ := os.ReadFile(filepath.Join(user, "task.md.tmpl")); string(b) != "v1\n" {
		t.Fatalf("clone produced %q, expected v1", string(b))
	}

	// Source pushes an update.
	if err := os.WriteFile(filepath.Join(src, "task.md.tmpl"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("-C", src, "commit", "-am", "v2")
	runGit("-C", src, "push", "origin", "main")

	out, err := runCLI(t, "--templates-dir", user, "templates", "pull")
	if err != nil {
		t.Fatalf("pull failed: %v (output: %s)", err, out)
	}
	if b, _ := os.ReadFile(filepath.Join(user, "task.md.tmpl")); string(b) != "v2\n" {
		t.Fatalf("pull did not sync to v2: %q", string(b))
	}
}

// TestTemplatesValidateAllPass scaffolds a fresh dir and confirms every
// embedded template parses + renders cleanly against tmpl.Samples — this
// also exercises whether the Samples registry stays in sync with the
// struct shapes each cmd/*.go file uses.
func TestTemplatesValidateAllPass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err != nil {
		t.Fatalf("validate failed: %v (output: %s)", err, out)
	}
	if strings.Contains(out, "FAIL") {
		t.Fatalf("expected all OK, got: %s", out)
	}
	// Spot-check a couple of templates appear in the per-file output.
	for _, name := range []string{"task.md.tmpl", "postmortem.md.tmpl", "license_mit.tmpl"} {
		if !strings.Contains(out, "OK    "+name) {
			t.Errorf("expected OK line for %s, got: %s", name, out)
		}
	}
}

// TestTemplatesValidateCatchesTypo writes a template that references a
// field name not in the canonical struct shape — validation must fail
// and the error must point at the bad field.
func TestTemplatesValidateCatchesTypo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// .Servce typo (should be .Service)
	if err := os.WriteFile(filepath.Join(dir, "runbook.md.tmpl"),
		[]byte("# bad\n{{ .Servce }}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err == nil {
		t.Fatalf("expected non-zero exit on typo, output: %s", out)
	}
	if !strings.Contains(out, "FAIL  runbook.md.tmpl") {
		t.Errorf("expected FAIL line for runbook, got: %s", out)
	}
	if !strings.Contains(out, "Servce") {
		t.Errorf("error should reference the bad field name, got: %s", out)
	}
}

// TestTemplatesValidateCatchesSyntaxError covers the parse-time path.
func TestTemplatesValidateCatchesSyntaxError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"),
		[]byte("{{ .Title \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err == nil {
		t.Fatalf("expected non-zero exit on syntax error, output: %s", out)
	}
	if !strings.Contains(out, "FAIL  task.md.tmpl") {
		t.Errorf("expected FAIL for task.md.tmpl, got: %s", out)
	}
	if !strings.Contains(out, "parse") {
		t.Errorf("error should mention parse failure, got: %s", out)
	}
}

// TestTemplatesValidateUserOnlyTemplate verifies that a template whose
// filename isn't a built-in (e.g. user's bespoke template used via
// --template) gets parse-only validation instead of being rejected.
func TestTemplatesValidateUserOnlyTemplate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "my-custom.md.tmpl"),
		[]byte("hello {{ .Whatever }}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err != nil {
		t.Fatalf("validate failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "my-custom.md.tmpl (parse-only") {
		t.Errorf("expected parse-only note for user-only template, got: %s", out)
	}
}

// TestTemplatesValidateAcceptsManifest verifies that .sections.yaml files
// are recognized by `templates validate` and parsed as section manifests.
func TestTemplatesValidateAcceptsManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err != nil {
		t.Fatalf("validate failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "OK    postmortem.sections.yaml") {
		t.Errorf("expected OK line for sidecar manifest, got: %s", out)
	}
}

// TestTemplatesValidateRejectsBrokenManifest verifies that a malformed
// .sections.yaml fails validation with a descriptive message.
func TestTemplatesValidateRejectsBrokenManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Unknown section type — manifest validation must surface it.
	bad := []byte("version: 1\nsections:\n  - id: x\n    title: X\n    type: image\n")
	if err := os.WriteFile(filepath.Join(dir, "postmortem.sections.yaml"), bad, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "validate", dir)
	if err == nil {
		t.Fatalf("expected non-zero exit, output: %s", out)
	}
	if !strings.Contains(out, "FAIL  postmortem.sections.yaml") {
		t.Errorf("expected FAIL line for manifest, got: %s", out)
	}
	if !strings.Contains(out, "unknown type") {
		t.Errorf("error should mention unknown type, got: %s", out)
	}
}

func TestTemplatesDiffAllMatch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	out, err := runCLI(t, "templates", "diff", dir)
	if err != nil {
		t.Fatalf("diff failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "all templates match") {
		t.Errorf("expected clean-match summary, got: %s", out)
	}
}

func TestTemplatesDiffShowsModification(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Append a sentinel line so the file differs from embedded.
	target := filepath.Join(dir, "runbook.md.tmpl")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, append(b, []byte("\nSENTINEL_LINE\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "diff", dir, "--no-color")
	if err != nil {
		t.Fatalf("diff failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "embedded/runbook.md.tmpl") || !strings.Contains(out, "user/runbook.md.tmpl") {
		t.Errorf("expected diff header for runbook, got: %s", out)
	}
	if !strings.Contains(out, "SENTINEL_LINE") {
		t.Errorf("expected modification to appear in diff body, got: %s", out)
	}
}

func TestTemplatesDiffNameOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	target := filepath.Join(dir, "slo.md.tmpl")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, append(b, []byte("\n# tweak\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "diff", dir, "--name-only")
	if err != nil {
		t.Fatalf("diff failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "differs  slo.md.tmpl") {
		t.Errorf("expected name-only line for slo, got: %s", out)
	}
	// Make sure we did NOT emit a full diff body.
	if strings.Contains(out, "diff --git") {
		t.Errorf("--name-only should suppress diff bodies, got: %s", out)
	}
}

func TestTemplatesDiffUserOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "my-custom.md.tmpl"),
		[]byte("not in the binary\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "diff", dir)
	if err != nil {
		t.Fatalf("diff failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "user-only  my-custom.md.tmpl") {
		t.Errorf("expected user-only line, got: %s", out)
	}
}

// TestConfigInitWritesYAML covers the happy non-interactive path: with
// --yes plus explicit flags, no prompts are needed and the file is written
// with the expected keys.
func TestConfigInitWritesYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, ".srekit.yaml")

	out, err := runCLI(t,
		"--config", target,
		"config", "init",
		"--yes",
		"--author", "Test Person",
		"--email", "t@example.com",
		"--templates-dir", "~/.srekit/templates",
	)
	if err != nil {
		t.Fatalf("config init failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "Wrote ") {
		t.Errorf("expected friendly confirmation, got: %s", out)
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	got := string(body)
	for _, want := range []string{
		"author: Test Person",
		"email: t@example.com",
		"templates_dir: ~/.srekit/templates",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// TestConfigInitOmitsTemplatesDir verifies that when templates_dir is empty
// we emit the key as a commented-out hint instead of an empty value.
func TestConfigInitOmitsTemplatesDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, ".srekit.yaml")

	if _, err := runCLI(t,
		"--config", target,
		"config", "init",
		"--yes",
		"--author", "A",
		"--email", "a@a",
	); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if strings.Contains(got, "templates_dir:") && !strings.Contains(got, "# templates_dir:") {
		t.Errorf("expected templates_dir to be commented out when empty, got:\n%s", got)
	}
}

// TestConfigInitRefusesOverwrite locks the no-clobber default.
func TestConfigInitRefusesOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, ".srekit.yaml")
	if err := os.WriteFile(target, []byte("preexisting: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := runCLI(t,
		"--config", target,
		"config", "init",
		"--yes",
		"--author", "A",
		"--email", "a@a",
	)
	if err == nil {
		t.Fatal("expected error when config file exists without --force")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error should mention --force, got: %v", err)
	}
	// Verify the file wasn't clobbered.
	b, _ := os.ReadFile(target)
	if string(b) != "preexisting: true\n" {
		t.Fatalf("existing file was overwritten: %q", string(b))
	}
}

// TestConfigInitForce overwrites on top of an existing file.
func TestConfigInitForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, ".srekit.yaml")
	if err := os.WriteFile(target, []byte("preexisting: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t,
		"--config", target,
		"config", "init",
		"--yes", "--force",
		"--author", "B", "--email", "b@b",
	); err != nil {
		t.Fatalf("config init --force failed: %v", err)
	}
	b, _ := os.ReadFile(target)
	if !strings.Contains(string(b), "author: B") {
		t.Fatalf("--force should overwrite; got: %s", string(b))
	}
}

// TestConfigInitMissingAuthorFails covers the non-interactive case where
// --author isn't passed and git config has nothing useful — the command
// must error rather than write an invalid file.
func TestConfigInitMissingAuthorFails(t *testing.T) {
	// Run from a temp dir with no git config so gitConfigValue returns "".
	// Also clear HOME so a developer's global ~/.gitconfig doesn't leak in.
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(dir, "empty-gitconfig"))

	target := filepath.Join(dir, ".srekit.yaml")
	_, err := runCLI(t,
		"--config", target,
		"config", "init",
		"--yes",
	)
	if err == nil {
		t.Fatal("expected error when --yes has no author/email source")
	}
	if !strings.Contains(err.Error(), "author") && !strings.Contains(err.Error(), "email") {
		t.Fatalf("error should mention author or email, got: %v", err)
	}
}

// TestTaskJSON verifies --json emits the bootstrap envelope shape (meta +
// one synthetic "body" section) for commands that have not migrated to a
// sections manifest. camelCase keys remain the public contract.
func TestTaskJSON(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "task", "--title", "Tail latency", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	meta, ok := got["meta"].(map[string]any)
	if !ok {
		t.Fatalf("missing meta object in %v", got)
	}
	if meta["title"] != "Tail latency" {
		t.Errorf("meta.title mismatch: %v", meta["title"])
	}
	if _, ok := meta["id"].(string); !ok {
		t.Errorf("expected string meta.id, got %T: %v", meta["id"], meta["id"])
	}
	secs, ok := got["sections"].([]any)
	if !ok || len(secs) != 1 {
		t.Fatalf("expected one bootstrap section, got %v", got["sections"])
	}
	s0 := secs[0].(map[string]any)
	if s0["id"] != "body" || s0["type"] != "text" {
		t.Errorf("bootstrap section shape wrong: %v", s0)
	}
	if !strings.Contains(s0["body"].(string), "Tail latency") {
		t.Errorf("bootstrap body should include rendered markdown: %v", s0["body"])
	}
}

// TestPostmortemJSON verifies the structured --json shape: postmortem owns
// a sidecar manifest, so the payload is {meta, sections:[...]} with one
// element per section in manifest order.
func TestPostmortemJSON(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "postmortem",
		"--title", "API outage",
		"--severity", "SEV-1",
		"--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	meta, ok := got["meta"].(map[string]any)
	if !ok {
		t.Fatalf("missing meta object in %v", got)
	}
	if meta["title"] != "API outage" || meta["severity"] != "SEV-1" {
		t.Errorf("meta mismatch: %+v", meta)
	}
	secs, ok := got["sections"].([]any)
	if !ok || len(secs) < 5 {
		t.Fatalf("expected multiple sections, got %v", got["sections"])
	}
	first := secs[0].(map[string]any)
	if first["id"] != "summary" {
		t.Errorf("first section should be summary, got %v", first["id"])
	}
	// Every section has the canonical shape.
	for i, s := range secs {
		m := s.(map[string]any)
		for _, key := range []string{"id", "title", "type", "required", "body"} {
			if _, ok := m[key]; !ok {
				t.Errorf("section[%d] missing %q: %v", i, key, m)
			}
		}
	}
}

// TestPostmortemFromInput verifies the --from round-trip: an input.json
// with overrides for a subset of section bodies renders Markdown where
// those overrides survive and the remaining sections fall back to the
// manifest defaults.
func TestPostmortemFromInput(t *testing.T) {
	t.Parallel()
	input := `{
	  "sections": {
	    "summary": "Outage lasted 27 minutes during peak EU traffic.",
	    "root_cause": "Cache eviction storm after capacity flag flip."
	  }
	}`
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "postmortem",
		"--title", "Outage",
		"--from", inputPath,
		"--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Outage lasted 27 minutes") {
		t.Errorf("summary override not applied:\n%s", out)
	}
	if !strings.Contains(out, "Cache eviction storm") {
		t.Errorf("root_cause override not applied:\n%s", out)
	}
	if !strings.Contains(out, "Влияние на пользователей:") {
		t.Errorf("expected impact section to come from defaults:\n%s", out)
	}
}

// TestPostmortemFromUnknownSectionID verifies the typo-guard: an input
// with an unknown section ID fails with a message that lists both the
// offending IDs and the manifest's known set.
func TestPostmortemFromUnknownSectionID(t *testing.T) {
	t.Parallel()
	input := `{"sections": {"summery": "typo"}}`
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runCLI(t, "postmortem", "--title", "X", "--from", inputPath, "--stdout")
	if err == nil {
		t.Fatal("expected error for unknown section id")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unknown section IDs") {
		t.Errorf("error should mention 'unknown section IDs', got: %v", err)
	}
	if !strings.Contains(msg, "summery") {
		t.Errorf("error should mention the bad id, got: %v", err)
	}
	if !strings.Contains(msg, "summary") {
		t.Errorf("error should list known ids (e.g. summary), got: %v", err)
	}
}

// TestIncidentJSONBootstrap is a representative test for the bootstrap
// envelope shape used by every generator that hasn't migrated to a
// sections manifest. (Per-command spot-checks for runbook/rfc/etc. would
// be redundant — all go through the same RenderOptions path.)
func TestIncidentJSONBootstrap(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "incident", "--title", "Checkout 5xx", "--severity", "SEV-1", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	meta := got["meta"].(map[string]any)
	if meta["title"] != "Checkout 5xx" || meta["severity"] != "SEV-1" {
		t.Errorf("meta mismatch: %v", meta)
	}
	secs := got["sections"].([]any)
	if len(secs) != 1 {
		t.Fatalf("expected one bootstrap section, got %d", len(secs))
	}
	s0 := secs[0].(map[string]any)
	if s0["id"] != "body" {
		t.Errorf("bootstrap section id should be 'body', got %v", s0["id"])
	}
}

// TestTemplatesUpgradeAddsMissing scaffolds a partial dir (one template
// missing) and verifies upgrade copies the missing one in and reports it.
func TestTemplatesUpgradeAddsMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Remove one template to simulate "new in this binary".
	missing := filepath.Join(dir, "runbook.md.tmpl")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	if _, err := os.Stat(missing); err != nil {
		t.Fatalf("upgrade did not restore runbook.md.tmpl: %v", err)
	}
	if !strings.Contains(out, "+ added     runbook.md.tmpl") {
		t.Errorf("expected '+ added' line for runbook, got: %s", out)
	}
}

// TestTemplatesUpgradeSkipsCustomized verifies that when the user has
// customized a file but upstream hasn't changed (snapshot == embedded),
// upgrade silently leaves the file alone and counts it as skipped.
func TestTemplatesUpgradeSkipsCustomized(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	target := filepath.Join(dir, "task.md.tmpl")
	customized := []byte("# MY CUSTOM TASK: {{ .Title }}\n")
	if err := os.WriteFile(target, customized, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	b, _ := os.ReadFile(target)
	if string(b) != string(customized) {
		t.Fatalf("upgrade clobbered customized file: %s", string(b))
	}
	// Summary still accounts for it; no per-file line is expected because
	// there is nothing for 3-way to do here.
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("expected summary to count 1 skipped, got: %s", out)
	}
}

// TestTemplatesUpgradeForce verifies --force overwrites the user's customizations.
func TestTemplatesUpgradeForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	target := filepath.Join(dir, "task.md.tmpl")
	if err := os.WriteFile(target, []byte("CUSTOM\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "upgrade", dir, "--force")
	if err != nil {
		t.Fatalf("upgrade --force failed: %v (output: %s)", err, out)
	}
	b, _ := os.ReadFile(target)
	if !strings.Contains(string(b), "Расследование") {
		t.Fatalf("--force should overwrite with embedded content; got: %s", string(b))
	}
	if !strings.Contains(out, "~ updated   task.md.tmpl") {
		t.Errorf("expected '~ updated' line, got: %s", out)
	}
}

// TestTemplatesUpgradeDryRun verifies --dry-run reports diffs but writes nothing.
func TestTemplatesUpgradeDryRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	missing := filepath.Join(dir, "runbook.md.tmpl")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir, "--dry-run")
	if err != nil {
		t.Fatalf("upgrade --dry-run failed: %v (output: %s)", err, out)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("--dry-run must not write files; runbook.md.tmpl reappeared")
	}
	if !strings.Contains(out, "+ added     runbook.md.tmpl") {
		t.Errorf("expected '+ added' line in dry-run report, got: %s", out)
	}
	if !strings.Contains(out, "dry-run:") {
		t.Errorf("expected 'dry-run:' label in summary, got: %s", out)
	}
}

// TestTemplatesInitRespectsConfiguredDir locks in the fix where 'templates
// init' without a positional arg used to ignore templates_dir from viper
// and unconditionally scaffold ~/.srekit/templates. It must now resolve
// the configured directory the same way every other subcommand does.
func TestTemplatesInitRespectsConfiguredDir(t *testing.T) {
	// Uses withViper (global viper) so it is not parallel-safe.
	// Pre-stage an existing empty dir so configureTemplates (the root
	// PersistentPreRunE) doesn't warn about a missing path.
	dir := filepath.Join(t.TempDir(), "configured-templates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	withViper(t, map[string]string{"templates_dir": dir})

	if _, err := runCLI(t, "templates", "init", "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Files should land in the configured dir, not in defaultTemplatesDir.
	for _, name := range []string{"task.md.tmpl", "TEMPLATES.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s in configured dir, got: %v", name, err)
		}
	}
}

// TestTemplatesInitSeedsSnapshot verifies init writes the .srekit-embedded
// sidecar so the next upgrade has a merge base.
func TestTemplatesInitSeedsSnapshot(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	for _, name := range []string{"task.md.tmpl", "postmortem.md.tmpl", "runbook.md.tmpl"} {
		snap := filepath.Join(dir, ".srekit-embedded", name)
		body, err := os.ReadFile(snap)
		if err != nil {
			t.Errorf("snapshot %s missing: %v", name, err)
			continue
		}
		userBody, _ := os.ReadFile(filepath.Join(dir, name))
		if string(body) != string(userBody) {
			t.Errorf("snapshot %s should match the file just written by init", name)
		}
	}
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to be created: %v", err)
	}
	if !strings.Contains(string(gitignore), ".srekit-embedded/") {
		t.Errorf(".gitignore should contain '.srekit-embedded/', got: %s", string(gitignore))
	}
}

// TestTemplatesUpgrade3WayCleanMerge constructs a textbook 3-way scenario:
// user edits the top of a file, "upstream" edits the bottom, both off the
// same base. With non-overlapping edits, git merge-file produces a clean
// merged file with both changes.
func TestTemplatesUpgrade3WayCleanMerge(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	// Strategy: keep upstream's diff and user's diff in completely
	// non-overlapping regions, separated by the full embedded body of
	// context lines.
	//
	//   base     = "EXTRA_BOTTOM\n" appended to embedded body
	//   user     = "USER_TOP\n" prepended to base  (user edits the TOP region)
	//   upstream = embedded body, no EXTRA_BOTTOM  (upstream drops the BOTTOM)
	//
	// merge-file sees base→user as "add line at top", base→upstream as
	// "remove line at bottom" — disjoint hunks, clean merge.
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	name := "task.md.tmpl"
	embeddedBody, _ := os.ReadFile(filepath.Join(dir, name)) // identical to embedded
	snapBody := append(append([]byte{}, embeddedBody...), []byte("EXTRA_BOTTOM\n")...)
	userBody := append([]byte("USER_TOP\n"), snapBody...)

	if err := os.WriteFile(filepath.Join(dir, ".srekit-embedded", name), snapBody, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), userBody, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	merged, _ := os.ReadFile(filepath.Join(dir, name))
	if !bytes.Contains(merged, []byte("USER_TOP")) {
		t.Errorf("clean merge should preserve USER_TOP, got:\n%s", merged)
	}
	if bytes.Contains(merged, []byte("EXTRA_BOTTOM")) {
		t.Errorf("clean merge should drop EXTRA_BOTTOM (upstream removed it), got:\n%s", merged)
	}
	if bytes.Contains(merged, []byte("<<<<<<<")) {
		t.Errorf("clean merge should have no conflict markers, got:\n%s", merged)
	}
	if !strings.Contains(out, "~ merged    "+name) {
		t.Errorf("expected '~ merged' line, got: %s", out)
	}
	if !strings.Contains(out, "1 merged") {
		t.Errorf("expected '1 merged' in summary, got: %s", out)
	}

	// Snapshot moved to current embedded.
	snap, _ := os.ReadFile(filepath.Join(dir, ".srekit-embedded", name))
	if !bytes.Equal(snap, embeddedBody) {
		t.Errorf("snapshot should advance to current embedded after merge")
	}
}

// TestTemplatesUpgrade3WayConflict constructs overlapping edits on the same
// region — merge-file must surface conflict markers and the command must
// exit non-zero.
func TestTemplatesUpgrade3WayConflict(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	name := "task.md.tmpl"
	embeddedBody, _ := os.ReadFile(filepath.Join(dir, name))

	// Snapshot: embedded + leading TARGET\n. User edits to USER-EDIT,
	// upstream (embedded) drops TARGET — overlap.
	snapBody := append([]byte("TARGET\n"), embeddedBody...)
	userBody := bytes.Replace(snapBody, []byte("TARGET\n"), []byte("USER-EDIT\n"), 1)

	if err := os.WriteFile(filepath.Join(dir, ".srekit-embedded", name), snapBody, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), userBody, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err == nil {
		t.Fatalf("expected non-zero exit on conflict, output: %s", out)
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("error should mention conflict, got: %v", err)
	}
	merged, _ := os.ReadFile(filepath.Join(dir, name))
	if !bytes.Contains(merged, []byte("<<<<<<<")) || !bytes.Contains(merged, []byte(">>>>>>>")) {
		t.Errorf("conflict markers missing from merged file:\n%s", merged)
	}
	if !strings.Contains(out, "X conflict  "+name) {
		t.Errorf("expected 'X conflict' line, got: %s", out)
	}
}

// TestTemplatesUpgradeFastForward verifies that when the user hasn't edited
// a file but upstream did, the file is fast-forwarded without --force.
func TestTemplatesUpgradeFastForward(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	name := "task.md.tmpl"
	// Simulate "upstream changed since snapshot": rewrite the snapshot to
	// an older shape (the user file still matches that older shape).
	oldShape := []byte("OLD EMBEDDED VERSION\n")
	if err := os.WriteFile(filepath.Join(dir, name), oldShape, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".srekit-embedded", name), oldShape, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	merged, _ := os.ReadFile(filepath.Join(dir, name))
	if bytes.Equal(merged, oldShape) {
		t.Errorf("expected fast-forward to current embedded, file still at old shape")
	}
	if !strings.Contains(out, "~ updated   "+name+" (upstream change, no local edits)") {
		t.Errorf("expected fast-forward line, got: %s", out)
	}
}

// TestTemplatesUpgradeNoSnapshotFallback verifies behavior when the user
// dir has no .srekit-embedded sidecar (e.g. scaffolded before this
// feature): customized files are skipped but the snapshot is seeded so
// the *next* upgrade can do 3-way.
func TestTemplatesUpgradeNoSnapshotFallback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Wipe the sidecar to simulate "pre-3-way user dir."
	if err := os.RemoveAll(filepath.Join(dir, ".srekit-embedded")); err != nil {
		t.Fatal(err)
	}
	name := "task.md.tmpl"
	customized := []byte("CUSTOMIZED\n")
	if err := os.WriteFile(filepath.Join(dir, name), customized, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "! skipped   "+name+" (customized; no merge base") {
		t.Errorf("expected no-snapshot skip line, got: %s", out)
	}
	// Snapshot must now exist for next upgrade.
	if _, err := os.Stat(filepath.Join(dir, ".srekit-embedded", name)); err != nil {
		t.Errorf("expected snapshot to be seeded after no-base skip, got: %v", err)
	}
	// File preserved.
	b, _ := os.ReadFile(filepath.Join(dir, name))
	if string(b) != string(customized) {
		t.Errorf("customized file should be preserved, got: %s", string(b))
	}
}

// TestTemplatesUpgradeAllUnchanged verifies the quiet path: fresh init then
// upgrade should report 0 added / 0 updated / N unchanged.
func TestTemplatesUpgradeAllUnchanged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	if strings.Contains(out, "+ added") || strings.Contains(out, "~ updated") || strings.Contains(out, "! skipped") {
		t.Errorf("expected quiet upgrade after fresh init, got: %s", out)
	}
	if !strings.Contains(out, "0 added, 0 updated") {
		t.Errorf("expected summary to show 0 changes, got: %s", out)
	}
}

// TestTemplatesListClassifies verifies all four statuses round-trip: after
// init, customize one, drop another, add a bespoke file.
func TestTemplatesListClassifies(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Customize one (must remain a *.tmpl).
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), []byte("CUSTOM\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Remove one so it becomes embedded-only.
	if err := os.Remove(filepath.Join(dir, "runbook.md.tmpl")); err != nil {
		t.Fatal(err)
	}
	// Add a bespoke one with no embedded counterpart.
	if err := os.WriteFile(filepath.Join(dir, "my-custom.md.tmpl"), []byte("MINE\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "list", dir)
	if err != nil {
		t.Fatalf("list failed: %v (output: %s)", err, out)
	}
	for _, want := range []string{
		"task.md.tmpl",
		"customized",
		"runbook.md.tmpl",
		"embedded-only",
		"my-custom.md.tmpl",
		"user-only",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
	// At least one identical (postmortem hasn't been touched).
	if !strings.Contains(out, "postmortem.md.tmpl") || !strings.Contains(out, "identical") {
		t.Errorf("expected an identical entry, got:\n%s", out)
	}
}

// TestTemplatesListJSON verifies --json round-trips through encoding/json.
func TestTemplatesListJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), []byte("CUSTOM\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "list", dir, "--json")
	if err != nil {
		t.Fatalf("list --json failed: %v (output: %s)", err, out)
	}
	var got []map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	found := false
	for _, e := range got {
		if e["name"] == "task.md.tmpl" {
			if e["status"] != "customized" {
				t.Errorf("task.md.tmpl status: got %q, want customized", e["status"])
			}
			found = true
		}
	}
	if !found {
		t.Errorf("task.md.tmpl entry missing in JSON output: %s", out)
	}
}

// TestTemplatesListFilter verifies --filter narrows the output.
func TestTemplatesListFilter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), []byte("CUSTOM\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "list", dir, "--filter", "customized")
	if err != nil {
		t.Fatalf("list --filter failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "task.md.tmpl") {
		t.Errorf("expected task.md.tmpl in customized filter, got: %s", out)
	}
	// No 'identical' entries should slip through.
	if strings.Contains(out, "identical") {
		t.Errorf("--filter=customized leaked identical entries: %s", out)
	}
}

// TestTemplatesListInvalidFilter verifies bad --filter values are rejected.
func TestTemplatesListInvalidFilter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatal(err)
	}
	_, err := runCLI(t, "templates", "list", dir, "--filter", "bogus")
	if err == nil {
		t.Fatal("expected error on bogus --filter value")
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
