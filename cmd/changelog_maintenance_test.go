package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
)

// releasedScaffold is what `srekit changelog` emits with two real entries
// added under [Unreleased] and the rest left as scaffold placeholders. It is
// the shape a user actually cuts a release from.
const releasedScaffold = `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

-

### Changed

-

### Deprecated

-

### Removed

-

### Fixed

- Race in the connection pool.
- Retry storm on 503.

### Security

-

## [0.1.0] - 2026-01-01

### Added

- Initial release.

[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`

// writeChangelog drops src into a fresh temp dir, chdirs there, and returns
// a reader for the file's current contents.
func writeChangelog(t *testing.T, src string) (read func() string) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	path := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	return func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
}

// mustIndex is strings.Index with the not-found case turned into a test
// failure, so a slice built from it can never silently start at -1.
func mustIndex(t *testing.T, s, sub string) int {
	t.Helper()
	i := strings.Index(s, sub)
	if i < 0 {
		t.Fatalf("%q not found in:\n%s", sub, s)
	}
	return i
}

func TestChangelogReleaseMovesUnreleased(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)

	out, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "wrote CHANGELOG.md") {
		t.Errorf("expected a wrote line, got: %s", out)
	}

	got := read()
	if !strings.Contains(got, "## [1.0.0] - 2026-03-04\n\n### Fixed\n\n- Race in the connection pool.\n- Retry storm on 503.\n") {
		t.Errorf("entries did not move under the new heading:\n%s", got)
	}
	if !strings.Contains(got, "## [Unreleased]\n\n## [1.0.0]") {
		t.Errorf("[Unreleased] should remain, empty:\n%s", got)
	}
	if mustIndex(t, got, "## [1.0.0]") > mustIndex(t, got, "## [0.1.0]") {
		t.Error("the new version must come before previously released ones")
	}
}

// Only the change types with real entries ship; the scaffold's five
// placeholder subsections must not appear in the released version.
func TestChangelogReleaseDropsPlaceholders(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)

	if _, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04"); err != nil {
		t.Fatal(err)
	}
	got := read()
	released := got[mustIndex(t, got, "## [1.0.0]"):mustIndex(t, got, "## [0.1.0]")]
	if !strings.Contains(released, "### Fixed") {
		t.Errorf("Fixed had entries and must ship:\n%s", released)
	}
	for _, absent := range []string{"### Added", "### Changed", "### Deprecated", "### Removed", "### Security"} {
		if strings.Contains(released, absent) {
			t.Errorf("placeholder-only %s shipped:\n%s", absent, released)
		}
	}
}

// With no --date the release is dated today, off the injectable clock.
func TestChangelogReleaseDefaultsToToday(t *testing.T) {
	orig := clock.Now
	t.Cleanup(func() { clock.Now = orig })
	clock.Now = func() time.Time { return time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC) }

	read := writeChangelog(t, releasedScaffold)
	if _, err := runCLI(t, "changelog", "release", "--version", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(), "## [1.0.0] - 2026-07-09") {
		t.Errorf("release should be dated today:\n%s", read())
	}
}

func TestChangelogReleaseYanked(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)

	if _, err := runCLI(t, "changelog", "release", "--version", "0.0.5", "--date", "2014-12-13", "--yanked"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(), "## [0.0.5] - 2014-12-13 [YANKED]\n") {
		t.Errorf("yanked heading wrong:\n%s", read())
	}
}

// An explicit positional target is the file rewritten, and CHANGELOG.md in
// the working directory is left alone.
func TestChangelogReleaseExplicitPath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	root := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(root, []byte(releasedScaffold), 0o600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "docs")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(nested, "CHANGELOG.md")
	if err := os.WriteFile(target, []byte(releasedScaffold), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04", "docs/CHANGELOG.md"); err != nil {
		t.Fatal(err)
	}

	rewritten, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rewritten), "## [1.0.0] - 2026-03-04") {
		t.Errorf("the positional target was not rewritten:\n%s", rewritten)
	}
	untouched, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	if string(untouched) != releasedScaffold {
		t.Error("CHANGELOG.md in the working directory must not be touched")
	}
}

// Every refusal path leaves the file byte-identical. That is the property
// that makes the command safe to run against a document someone else wrote.
func TestChangelogReleaseRefusalsLeaveFileIntact(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		args    []string
		wantErr string
	}{
		{
			name:    "non-ISO date",
			src:     releasedScaffold,
			args:    []string{"--version", "1.0.0", "--date", "04/03/2026"},
			wantErr: "YYYY-MM-DD",
		},
		{
			name:    "version already released",
			src:     releasedScaffold,
			args:    []string{"--version", "0.1.0"},
			wantErr: "0.1.0",
		},
		{
			name:    "nothing to release",
			src:     strings.Replace(releasedScaffold, "- Race in the connection pool.\n- Retry storm on 503.\n", "-\n", 1),
			args:    []string{"--version", "1.0.0"},
			wantErr: "no entries",
		},
		{
			name:    "no Unreleased heading",
			src:     strings.Replace(releasedScaffold, "## [Unreleased]", "## [0.2.0] - 2026-02-02", 1),
			args:    []string{"--version", "1.0.0"},
			wantErr: "[Unreleased]",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			read := writeChangelog(t, c.src)
			args := append([]string{"changelog", "release"}, c.args...)
			_, err := runCLI(t, args...)
			if err == nil {
				t.Fatal("expected a non-zero exit")
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error should mention %q, got: %v", c.wantErr, err)
			}
			if read() != c.src {
				t.Error("the file must be byte-identical after a refusal")
			}
		})
	}
}

func TestChangelogReleaseMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	_, err := runCLI(t, "changelog", "release", "--version", "1.0.0")
	if err == nil {
		t.Fatal("expected a non-zero exit")
	}
	if !strings.Contains(err.Error(), "CHANGELOG.md") || !strings.Contains(err.Error(), "srekit changelog") {
		t.Errorf("error should name the path and point at the generator, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "CHANGELOG.md")); statErr == nil {
		t.Error("no file may be created")
	}
}

func TestChangelogReleaseDryRunAndStdoutWriteNothing(t *testing.T) {
	for _, flag := range []string{"--dry-run", "--stdout"} {
		t.Run(flag, func(t *testing.T) {
			read := writeChangelog(t, releasedScaffold)
			out, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04", flag)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, "## [1.0.0] - 2026-03-04") {
				t.Errorf("the result should be printed:\n%s", out)
			}
			if read() != releasedScaffold {
				t.Error("the file must be byte-identical")
			}
			if flag == "--dry-run" && !strings.Contains(out, "CHANGELOG.md") {
				t.Errorf("dry-run should name the file it would have written:\n%s", out)
			}
		})
	}
}

func TestChangelogReleaseJSONDescribesDocument(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)

	out, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if read() != releasedScaffold {
		t.Error("--json must not modify the file")
	}

	var got struct {
		Path       string `json:"path"`
		Vocabulary string `json:"vocabulary"`
		Unreleased struct {
			Changes []struct {
				Type       string `json:"type"`
				HasEntries bool   `json:"hasEntries"`
			} `json:"changes"`
		} `json:"unreleased"`
		Versions []struct {
			Version string `json:"version"`
			Date    string `json:"date"`
			Yanked  bool   `json:"yanked"`
		} `json:"versions"`
		Links []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"links"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if got.Path != "CHANGELOG.md" || got.Vocabulary != "en" {
		t.Errorf("path/vocabulary wrong: %+v", got)
	}
	if len(got.Versions) != 1 || got.Versions[0].Version != "0.1.0" || got.Versions[0].Date != "2026-01-01" {
		t.Errorf("versions wrong: %+v", got.Versions)
	}
	if len(got.Links) != 2 {
		t.Errorf("want 2 link definitions, got %d", len(got.Links))
	}
	var fixed bool
	for _, c := range got.Unreleased.Changes {
		if c.Type == "Fixed" && c.HasEntries {
			fixed = true
		}
	}
	if !fixed {
		t.Errorf("Fixed should be reported as having entries: %+v", got.Unreleased.Changes)
	}
}

// An editing command is not a generator: a second destination has no
// meaning and an overwrite guard would guard against its own purpose.
func TestChangelogReleaseFlagSetIsNarrow(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "changelog", "release", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, absent := range []string{"--out ", "--force"} {
		if strings.Contains(out, absent) {
			t.Errorf("%s must not be offered on an editing command:\n%s", absent, out)
		}
	}
	for _, present := range []string{"--dry-run", "--stdout", "--json", "--yanked", "--date", "--version"} {
		if !strings.Contains(out, present) {
			t.Errorf("%s should be listed:\n%s", present, out)
		}
	}
}

func TestChangelogReleaseRejectsGeneratorFlags(t *testing.T) {
	for _, flag := range []string{"--out", "--force"} {
		t.Run(flag, func(t *testing.T) {
			read := writeChangelog(t, releasedScaffold)
			args := []string{"changelog", "release", "--version", "1.0.0", flag}
			if flag == "--out" {
				args = append(args, "other.md")
			}
			_, err := runCLI(t, args...)
			if err == nil {
				t.Fatalf("%s should be rejected as an unknown flag", flag)
			}
			if !strings.Contains(err.Error(), "unknown flag") {
				t.Errorf("want an unknown-flag error, got: %v", err)
			}
			if read() != releasedScaffold {
				t.Error("the file must be untouched")
			}
		})
	}
}

func TestChangelogReleaseRewritesLinkBlock(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)

	if _, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04"); err != nil {
		t.Fatal(err)
	}
	got := read()
	block := got[mustIndex(t, got, "[Unreleased]: "):]
	want := "[Unreleased]: https://github.com/acme/api/compare/v1.0.0...HEAD\n" +
		"[1.0.0]: https://github.com/acme/api/compare/v0.1.0...v1.0.0\n" +
		"[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0\n"
	if block != want {
		t.Errorf("link block wrong.\nwant:\n%s\ngot:\n%s", want, block)
	}
}

func TestChangelogReleaseLinkBlockPreservesConvention(t *testing.T) {
	src := `# Changelog

## [Unreleased]

### Added

- A thing.

## [1.1.0] - 2026-01-01

### Fixed

- Another thing.

[Unreleased]: https://git.example.com/group/proj/-/compare/1.1.0...HEAD
[1.1.0]: https://git.example.com/group/proj/-/compare/1.0.0...1.1.0
`
	read := writeChangelog(t, src)
	if _, err := runCLI(t, "changelog", "release", "--version", "1.2.0", "--date", "2026-03-04"); err != nil {
		t.Fatal(err)
	}
	got := read()
	if !strings.Contains(got, "[Unreleased]: https://git.example.com/group/proj/-/compare/1.2.0...HEAD\n") {
		t.Errorf("the document's own host must be preserved:\n%s", got)
	}
	if strings.Contains(got, "compare/v1.2.0") {
		t.Errorf("a v prefix was invented:\n%s", got)
	}
}

func TestChangelogReleaseFirstVersionPointsAtTag(t *testing.T) {
	src := `# Changelog

## [Unreleased]

### Added

- Initial work.

[Unreleased]: https://github.com/acme/api/compare/v0.0.0...HEAD
`
	read := writeChangelog(t, src)
	if _, err := runCLI(t, "changelog", "release", "--version", "0.1.0", "--date", "2026-03-04"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(), "[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0\n") {
		t.Errorf("the first release must point at its tag:\n%s", read())
	}
}

func TestChangelogValidatePassesOnReleasedScaffold(t *testing.T) {
	writeChangelogAndRelease(t)

	out, err := runCLI(t, "changelog", "validate")
	if err != nil {
		t.Fatalf("a released scaffold should validate cleanly: %v\n%s", err, out)
	}
	for _, check := range []string{"heading-shape", "unreleased-section", "version-order", "no-duplicate-versions", "change-types", "link-definitions"} {
		if !strings.Contains(out, "OK    "+check) {
			t.Errorf("check %s should be reported OK:\n%s", check, out)
		}
	}
}

func writeChangelogAndRelease(t *testing.T) {
	t.Helper()
	writeChangelog(t, releasedScaffold)
	if _, err := runCLI(t, "changelog", "release", "--version", "1.0.0", "--date", "2026-03-04", "--quiet"); err != nil {
		t.Fatal(err)
	}
}

func TestChangelogValidateFailures(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		check string
		names string
	}{
		{
			name:  "regional date",
			src:   strings.Replace(releasedScaffold, "## [0.1.0] - 2026-01-01", "## [0.1.0] - 04/03/2026", 1),
			check: "heading-shape",
			names: "04/03/2026",
		},
		{
			name:  "invented change type",
			src:   strings.Replace(releasedScaffold, "### Fixed", "### Improvements", 1),
			check: "change-types",
			names: "Improvements",
		},
		{
			name:  "missing link definition",
			src:   strings.Replace(releasedScaffold, "[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0\n", "", 1),
			check: "link-definitions",
			names: "0.1.0",
		},
		{
			name:  "versions out of order",
			src:   strings.Replace(releasedScaffold, "## [0.1.0] - 2026-01-01\n\n### Added\n\n- Initial release.\n", "## [0.1.0] - 2026-01-01\n\n### Added\n\n- Initial release.\n\n## [0.2.0] - 2026-02-01\n\n### Added\n\n- Later.\n", 1),
			check: "version-order",
			names: "0.2.0",
		},
		{
			name:  "duplicate version",
			src:   strings.Replace(releasedScaffold, "## [0.1.0] - 2026-01-01\n\n### Added\n\n- Initial release.\n", "## [0.1.0] - 2026-01-01\n\n### Added\n\n- Initial release.\n\n## [0.1.0] - 2025-01-01\n\n### Added\n\n- Again.\n", 1),
			check: "no-duplicate-versions",
			names: "0.1.0",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writeChangelog(t, c.src)
			out, err := runCLI(t, "changelog", "validate")
			if err == nil {
				t.Fatalf("expected a non-zero exit:\n%s", out)
			}
			if !strings.Contains(out, "FAIL  "+c.check) {
				t.Errorf("check %s should be reported FAIL:\n%s", c.check, out)
			}
			if !strings.Contains(out, c.names) {
				t.Errorf("the detail should name %q:\n%s", c.names, out)
			}
			// Every check is reported, not just the failing one.
			if got := strings.Count(out, "OK    ") + strings.Count(out, "FAIL  "); got != 7 {
				t.Errorf("want all 7 checks reported, got %d:\n%s", got, out)
			}
		})
	}
}

func TestChangelogValidateMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if _, err := runCLI(t, "changelog", "validate"); err == nil {
		t.Fatal("expected a non-zero exit")
	}
}

// The bare invocation is the generator it always was, and the maintenance
// subcommands are discoverable from its help.
func TestChangelogBareInvocationUnchanged(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "changelog", "--stdout", "--repo", "acme/api", "--version", "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# Changelog",
		"## [Unreleased]",
		"### Security",
		"[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD",
		"[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("scaffold missing %q:\n%s", want, out)
		}
	}
}

func TestChangelogHelpListsSubcommands(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "changelog", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"release", "validate"} {
		if !strings.Contains(out, want) {
			t.Errorf("%s should be listed as a subcommand:\n%s", want, out)
		}
	}
	// The generator keeps its own flags on the parent.
	for _, want := range []string{"--repo", "--from", "--out", "--force"} {
		if !strings.Contains(out, want) {
			t.Errorf("the generator flag %s should still be listed:\n%s", want, out)
		}
	}
}

// A maintenance subcommand is not a catalog entry: the root help lists only
// the parent.
func TestRootHelpListsOnlyChangelogParent(t *testing.T) {
	t.Parallel()
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, absent := range []string{"changelog release", "changelog validate"} {
		if strings.Contains(out, absent) {
			t.Errorf("%q must not appear in the root catalog:\n%s", absent, out)
		}
	}
}
