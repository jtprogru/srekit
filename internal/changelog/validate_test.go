package changelog

import (
	"strings"
	"testing"
)

func checks(t *testing.T, src string) map[string]CheckResult {
	t.Helper()
	out := map[string]CheckResult{}
	for _, r := range Validate(scan(t, src), Vocabularies()) {
		out[r.Name] = r
	}
	return out
}

const cleanDoc = `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [1.2.0] - 2026-03-04

### Added

- A thing.

## [1.1.0] - 2026-02-01

### Fixed

- Another thing.

[Unreleased]: https://github.com/acme/api/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/acme/api/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/acme/api/releases/tag/v1.1.0
`

func TestValidateCleanDocumentPasses(t *testing.T) {
	t.Parallel()
	for name, r := range checks(t, cleanDoc) {
		if !r.OK {
			t.Errorf("%s should pass: %s", name, r.Detail)
		}
	}
}

func TestValidateFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		src        string
		check      string
		wantDetail string
	}{
		{
			name:       "regional date",
			src:        strings.Replace(cleanDoc, "## [1.2.0] - 2026-03-04", "## [1.2.0] - 04/03/2026", 1),
			check:      CheckHeadingShape,
			wantDetail: "04/03/2026",
		},
		{
			name:       "invented change type",
			src:        strings.Replace(cleanDoc, "### Added", "### Improvements", 1),
			check:      CheckChangeTypes,
			wantDetail: "Improvements",
		},
		{
			name:       "missing link definition",
			src:        strings.Replace(cleanDoc, "[1.2.0]: https://github.com/acme/api/compare/v1.1.0...v1.2.0\n", "", 1),
			check:      CheckLinks,
			wantDetail: "1.2.0",
		},
		{
			name:       "versions out of order",
			src:        strings.Replace(cleanDoc, "## [1.2.0] - 2026-03-04", "## [1.0.0] - 2026-03-04", 1),
			check:      CheckOrder,
			wantDetail: "1.1.0",
		},
		{
			name:       "duplicate version",
			src:        strings.Replace(cleanDoc, "## [1.1.0] - 2026-02-01", "## [1.2.0] - 2026-02-01", 1),
			check:      CheckDuplicates,
			wantDetail: "1.2.0",
		},
		{
			name:       "no Unreleased section",
			src:        strings.Replace(cleanDoc, "## [Unreleased]\n\n", "", 1),
			check:      CheckUnreleased,
			wantDetail: "Unreleased",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := checks(t, c.src)
			r, ok := got[c.check]
			if !ok {
				t.Fatalf("check %q was not reported", c.check)
			}
			if r.OK {
				t.Fatalf("check %q should have failed", c.check)
			}
			if !strings.Contains(r.Detail, c.wantDetail) {
				t.Errorf("detail should name %q, got %q", c.wantDetail, r.Detail)
			}
		})
	}
}

// The change-type failure has to list the accepted set, or the person
// reading it has to go find the specification.
func TestValidateChangeTypeFailureListsAllowedSet(t *testing.T) {
	t.Parallel()
	r := checks(t, strings.Replace(cleanDoc, "### Added", "### Improvements", 1))[CheckChangeTypes]
	for _, want := range English.Types {
		if !strings.Contains(r.Detail, want) {
			t.Errorf("detail should list %q: %s", want, r.Detail)
		}
	}
}

// Every check is reported, so one pass shows the whole list rather than
// stopping at the first failure.
func TestValidateReportsEveryCheck(t *testing.T) {
	t.Parallel()
	results := Validate(scan(t, "# Changelog\n"), Vocabularies())
	if len(results) != 6 {
		t.Fatalf("want 6 checks, got %d", len(results))
	}
	seen := map[string]bool{}
	for _, r := range results {
		seen[r.Name] = true
	}
	for _, name := range []string{CheckHeadingShape, CheckUnreleased, CheckOrder, CheckDuplicates, CheckChangeTypes, CheckLinks} {
		if !seen[name] {
			t.Errorf("check %q missing", name)
		}
	}
}
