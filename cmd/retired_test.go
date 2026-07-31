package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRetiredCommandsExplainThemselves covers every name retired in
// v0.30.0: the command must fail with an explanation naming the release
// and the migration note, and must not write anything.
// Not parallel: t.Chdir isolates the working directory so a stray write
// is visible, and the two are mutually exclusive.
func TestRetiredCommandsExplainThemselves(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"capacity", []string{"capacity", "--service", "payments"}},
		{"retro", []string{"retro", "--team", "platform", "--sprint", "2026-W19"}},
		{"license", []string{"license", "--type", "mit", "--stdout"}},
		{"lic", []string{"lic"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run from a scratch dir so a stray write would be visible.
			dir := t.TempDir()
			t.Chdir(dir)

			out, err := runCLI(t, tc.args...)
			if err == nil {
				t.Fatalf("expected %q to fail, got output: %s", tc.name, out)
			}
			msg := err.Error()
			if !strings.Contains(msg, "was removed in v0.30.0") {
				t.Errorf("error should name the removal release, got: %v", err)
			}
			if !strings.Contains(msg, retiredDocsURL) {
				t.Errorf("error should point at the migration note, got: %v", err)
			}

			entries, readErr := os.ReadDir(dir)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(entries) != 0 {
				t.Errorf("retired command wrote %d entries into the working dir", len(entries))
			}
		})
	}
}

// TestRetiredCommandHelpExplainsRemoval covers the path that does not go
// through RunE: `srekit help capacity` is handled by cobra's help
// command, which exits 0 even for an unknown topic. Without Short/Long
// it printed a bare usage block that read like a working command — and
// advertised a `-h` that then failed.
func TestRetiredCommandHelpExplainsRemoval(t *testing.T) {
	t.Parallel()

	out, err := runCLI(t, "help", "capacity")
	if err != nil {
		t.Fatalf("help should not fail: %v", err)
	}
	if !strings.Contains(out, "was removed in v0.30.0") {
		t.Errorf("help output must state the removal, got:\n%s", out)
	}
	if !strings.Contains(out, retiredDocsURL) {
		t.Errorf("help output must point at the migration note, got:\n%s", out)
	}
}

// TestRetiredCommandIgnoresFlagValidation pins the reason the stubs
// disable flag parsing: a user debugging `--team is required` on a
// command that no longer exists cannot satisfy the error.
func TestRetiredCommandIgnoresFlagValidation(t *testing.T) {
	t.Parallel()

	_, err := runCLI(t, "retro")
	if err == nil {
		t.Fatal("expected retro to fail")
	}
	if !strings.Contains(err.Error(), "was removed") {
		t.Fatalf("expected the removal message, got: %v", err)
	}
	if strings.Contains(err.Error(), "--team") {
		t.Fatalf("stub must not validate flags, got: %v", err)
	}
}

// TestCatalogIsExactlyTheSurvivingGenerators locks the artifact catalog
// down: help lists the eight generators and neither advertises nor hides
// a surprise. Retired names must not appear.
func TestCatalogIsExactlyTheSurvivingGenerators(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	got := map[string]bool{}
	for _, c := range root.Commands() {
		if c.Hidden {
			continue
		}
		got[c.Name()] = true
	}

	want := []string{
		"task", "postmortem", "rfc", "runbook",
		"changelog", "oncall-report", "slo", "ebp",
		// management and diagnostic commands, not generators
		"templates", "config", "doctor",
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("expected %q in the visible command set", name)
		}
		delete(got, name)
	}
	// cobra injects help/completion; anything else is unexpected.
	for name := range got {
		switch name {
		case "help", "completion":
		default:
			t.Errorf("unexpected visible command %q", name)
		}
	}

	for _, retiredName := range []string{"capacity", "retro", "license", "lic"} {
		if strings.Contains(root.UsageString(), " "+retiredName+" ") {
			t.Errorf("retired name %q must not be advertised in help", retiredName)
		}
	}
}

// TestNoCommandOffersTemplateFlag pins the removal of --template FILE:
// it is now an unknown flag everywhere, not a silently ignored one.
// Not parallel: see TestRetiredCommandsExplainThemselves.
func TestNoCommandOffersTemplateFlag(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	for _, args := range [][]string{
		{"slo", "--service", "api", "--template", "./custom.tmpl"},
		{"rfc", "--title", "X", "--template", "./custom.tmpl"},
		{"postmortem", "--title", "X", "--template", "./custom.tmpl"},
	} {
		out, err := runCLI(t, args...)
		if err == nil {
			t.Fatalf("%v should have failed on --template, got: %s", args, out)
		}
		if !strings.Contains(err.Error(), "unknown flag: --template") {
			t.Errorf("%v: expected an unknown-flag error, got: %v", args, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("a rejected --template invocation wrote %d entries", len(entries))
	}
}

// TestLeftoverRetiredTemplateIsUserOnly covers the upgrade path for
// someone who ran `templates init` before v0.30.0: their capacity.yaml
// is no longer embedded, so it reclassifies from customized to
// user-only, and the surviving generators keep rendering.
func TestLeftoverRetiredTemplateIsUserOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Simulate the leftover from a pre-0.30.0 scaffold.
	leftover := filepath.Join(dir, "capacity.yaml")
	body := "version: 1\ntitle: Leftover\nsections:\n  - id: body\n    title: Body\n    type: text\n    body: x\n"
	if err := os.WriteFile(leftover, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "list", dir)
	if err != nil {
		t.Fatalf("list failed: %v (%s)", err, out)
	}
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "capacity.yaml") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("capacity.yaml missing from listing:\n%s", out)
	}
	if !strings.Contains(line, "user-only") {
		t.Errorf("expected capacity.yaml to be user-only, got: %s", line)
	}

	// A surviving generator still renders from that same directory.
	out, err = runCLI(t, "--templates-dir", dir, "slo", "--service", "api-gw", "--stdout")
	if err != nil {
		t.Fatalf("slo failed with a leftover retired template present: %v (%s)", err, out)
	}
	if !strings.Contains(out, "SLO — api-gw") {
		t.Errorf("slo body wrong: %s", out)
	}
}
