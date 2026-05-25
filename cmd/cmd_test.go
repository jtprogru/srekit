package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
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
	if !strings.Contains(out, "# Tasker - Tail latency") {
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
	if !strings.Contains(out, "Postmortem — API outage") || !strings.Contains(out, "SEV-1") {
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
	if !strings.Contains(out, "Runbook — p99 latency spike") || !strings.Contains(out, "api-gw") {
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
	if !strings.Contains(out, "On-call report — platform") || !strings.Contains(out, "2026-05-04") {
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
	if !strings.Contains(out, "Retro — platform / 2026-W19") {
		t.Fatalf("retro body wrong: %s", out)
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
