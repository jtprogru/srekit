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
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func TestTaskStdout(t *testing.T) {
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
	_, err := runCLI(t, "task", "--title=", "--stdout")
	if err == nil {
		t.Fatal("expected error when --title is empty")
	}
}

func TestLicenseWTFPLDefault(t *testing.T) {
	viper.Reset()
	viper.Set("author", "Test Person")
	viper.Set("email", "t@example.com")
	t.Cleanup(viper.Reset)

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
	viper.Reset()
	viper.Set("author", "Test Person")
	viper.Set("email", "t@example.com")
	t.Cleanup(viper.Reset)

	out, err := runCLI(t, "license", "--type", "mit", "--stdout", "--year", "2026")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "MIT License") {
		t.Fatalf("MIT body missing: %s", out)
	}
}

func TestLicenseUnknownType(t *testing.T) {
	viper.Reset()
	viper.Set("author", "x")
	viper.Set("email", "x@x")
	t.Cleanup(viper.Reset)
	_, err := runCLI(t, "license", "--type", "gpl", "--stdout")
	if err == nil {
		t.Fatal("expected error for unknown license type")
	}
}

func TestPostmortem(t *testing.T) {
	out, err := runCLI(t, "postmortem", "--title", "API outage", "--severity", "SEV-1", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Postmortem — API outage") || !strings.Contains(out, "SEV-1") {
		t.Fatalf("postmortem body wrong: %s", out)
	}
}

func TestRFC(t *testing.T) {
	viper.Reset()
	viper.Set("author", "Test Person")
	viper.Set("email", "t@example.com")
	t.Cleanup(viper.Reset)
	out, err := runCLI(t, "rfc", "--title", "Migrate to gRPC", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Migrate to gRPC") || !strings.Contains(out, "proposed") {
		t.Fatalf("rfc body wrong: %s", out)
	}
}

func TestRunbook(t *testing.T) {
	out, err := runCLI(t, "runbook", "--title", "p99 latency spike", "--service", "api-gw", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Runbook — p99 latency spike") || !strings.Contains(out, "api-gw") {
		t.Fatalf("runbook body wrong: %s", out)
	}
}

func TestChangelog(t *testing.T) {
	viper.Reset()
	viper.Set("author", "Test Person")
	viper.Set("email", "t@example.com")
	t.Cleanup(viper.Reset)
	out, err := runCLI(t, "changelog", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Keep a Changelog") {
		t.Fatalf("changelog body wrong: %s", out)
	}
}

func TestSretaskAlias(t *testing.T) {
	out, err := runCLI(t, "sretask", "--title", "alias works", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "alias works") {
		t.Fatalf("alias not working")
	}
}
