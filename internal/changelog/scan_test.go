package changelog

import (
	"strings"
	"testing"
)

func scan(t *testing.T, src string) *Document {
	t.Helper()
	return Scan([]byte(src), Vocabularies())
}

// The shapes below are the ones a hand-maintained changelog actually takes.
// Each is a document the scanner must survive without either erroring or
// silently reclassifying content.

const handEdited = `# Changelog

All notable changes to this project will be documented in this file.

<!-- editors: keep entries newest-first -->

Maintained by the platform team. Ping #platform before releasing.

## [Unreleased]

### Added

- Retry budget for idempotent calls.

### Fixed

-

## [1.10.0] - 2026-05-02

Some prose the author wrote between the heading and the first change type.

### Changed

- Bumped the connection pool default.

## [1.9.0] - 2026-04-01

- An entry with no change-type subsection at all.

## [0.0.5] - 2014-12-13 [YANKED]

### Removed

- Published by mistake.

[Unreleased]: https://github.com/acme/api/compare/v1.10.0...HEAD
[1.10.0]: https://github.com/acme/api/compare/v1.9.0...v1.10.0
[1.9.0]: https://github.com/acme/api/compare/v0.0.5...v1.9.0
[0.0.5]: https://github.com/acme/api/releases/tag/v0.0.5
`

func TestScanHandEditedDocument(t *testing.T) {
	t.Parallel()
	doc := scan(t, handEdited)

	if doc.Unreleased == nil {
		t.Fatal("Unreleased section not found")
	}
	if got := len(doc.Versions); got != 3 {
		t.Fatalf("want 3 released versions, got %d", got)
	}

	// The preamble covers the H1, the prose, and the HTML comment; nothing
	// of it may leak into the first section.
	pre := doc.Preamble.Text(doc.Src)
	for _, want := range []string{"# Changelog", "<!-- editors:", "Ping #platform"} {
		if !strings.Contains(pre, want) {
			t.Errorf("preamble missing %q:\n%s", want, pre)
		}
	}
	if strings.Contains(pre, "## [Unreleased]") {
		t.Error("preamble swallowed the Unreleased heading")
	}

	v110 := doc.FindVersion("1.10.0")
	if v110 == nil {
		t.Fatal("1.10.0 not found")
	}
	if !v110.WellFormed || v110.Date != "2026-05-02" || v110.Yanked {
		t.Errorf("1.10.0 parsed wrong: %+v", v110)
	}
	// Prose between the version heading and its first `###` is the lead.
	if lead := v110.Lead.Text(doc.Src); !strings.Contains(lead, "Some prose the author wrote") {
		t.Errorf("lead prose not captured: %q", lead)
	}

	// A version whose entries sit under no subsection at all keeps them in
	// the lead rather than losing them.
	v19 := doc.FindVersion("1.9.0")
	if v19 == nil {
		t.Fatal("1.9.0 not found")
	}
	if len(v19.Changes) != 0 {
		t.Errorf("1.9.0 should have no change subsections, got %d", len(v19.Changes))
	}
	if !strings.Contains(v19.Lead.Text(doc.Src), "no change-type subsection") {
		t.Error("1.9.0 entries lost")
	}

	yanked := doc.FindVersion("0.0.5")
	if yanked == nil || !yanked.Yanked || !yanked.WellFormed {
		t.Errorf("existing [YANKED] entry parsed wrong: %+v", yanked)
	}

	if !doc.LinksPresent || len(doc.Links) != 4 {
		t.Fatalf("link block wrong: present=%v n=%d", doc.LinksPresent, len(doc.Links))
	}
	if doc.Links[0].Label != "Unreleased" {
		t.Errorf("first link def should be Unreleased, got %q", doc.Links[0].Label)
	}
	if doc.Lang != "en" {
		t.Errorf("vocabulary should be en, got %q", doc.Lang)
	}
}

func TestScanPlaceholderDetection(t *testing.T) {
	t.Parallel()
	doc := scan(t, handEdited)
	got := map[string]bool{}
	for _, c := range doc.Unreleased.Changes {
		got[c.Heading] = c.HasEntries(doc.Src)
	}
	if !got["Added"] {
		t.Error("Added has a real bullet and must count as entries")
	}
	if got["Fixed"] {
		t.Error("Fixed holds only the bare `-` placeholder and must not count")
	}
}

func TestScanNoLinkBlock(t *testing.T) {
	t.Parallel()
	doc := scan(t, `# Changelog

## [Unreleased]

### Added

- Something.
`)
	if doc.LinksPresent {
		t.Fatal("document has no link block")
	}
	if doc.LinksRegion.Start != len(doc.Src) {
		t.Errorf("append point should be EOF, got %d of %d", doc.LinksRegion.Start, len(doc.Src))
	}
	if _, ok := doc.DeriveConvention(); ok {
		t.Error("no block means no convention to derive")
	}
}

func TestScanNonGitHubHostAndNoTagPrefix(t *testing.T) {
	t.Parallel()
	doc := scan(t, `# Changelog

## [Unreleased]

### Added

- Something.

## [1.1.0] - 2026-01-01

### Fixed

- Something else.

[Unreleased]: https://git.example.com/group/proj/-/compare/1.1.0...HEAD
[1.1.0]: https://git.example.com/group/proj/-/compare/1.0.0...1.1.0
`)
	conv, ok := doc.DeriveConvention()
	if !ok {
		t.Fatal("convention should be derivable")
	}
	if conv.CompareBase != "https://git.example.com/group/proj/-" {
		t.Errorf("compare base wrong: %q", conv.CompareBase)
	}
	if conv.TagPrefix != "" {
		t.Errorf("this project's tags carry no v prefix, got %q", conv.TagPrefix)
	}
	if got := conv.CompareURL("1.2.0", "HEAD"); got != "https://git.example.com/group/proj/-/compare/1.2.0...HEAD" {
		t.Errorf("compare URL wrong: %q", got)
	}
}

func TestScanTagPrefixFromGitHubBlock(t *testing.T) {
	t.Parallel()
	doc := scan(t, handEdited)
	conv, ok := doc.DeriveConvention()
	if !ok {
		t.Fatal("convention should be derivable")
	}
	if conv.TagPrefix != "v" {
		t.Errorf("tag prefix should be v, got %q", conv.TagPrefix)
	}
	if got := conv.TagURL("2.0.0"); got != "https://github.com/acme/api/releases/tag/v2.0.0" {
		t.Errorf("tag URL wrong: %q", got)
	}
}

// A document whose only link definition points at a release tag still has
// to yield a usable convention — that is the first-release shape.
func TestScanConventionFromTagOnlyBlock(t *testing.T) {
	t.Parallel()
	doc := scan(t, `# Changelog

## [Unreleased]

### Added

- Something.

## [0.1.0] - 2026-01-01

### Added

- Initial release.

[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`)
	conv, ok := doc.DeriveConvention()
	if !ok {
		t.Fatal("convention should be derivable from a tag definition alone")
	}
	if conv.CompareBase != "https://github.com/acme/api" || conv.TagPrefix != "v" {
		t.Errorf("convention wrong: %+v", conv)
	}
}

func TestScanUnparseableHeadingIsKept(t *testing.T) {
	t.Parallel()
	doc := scan(t, `# Changelog

## [Unreleased]

## [1.2.0] - 04/03/2026

### Added

- Regional date above.

## Release notes for the beta

- Not a version heading at all.
`)
	if len(doc.Versions) != 2 {
		t.Fatalf("both headings must be kept, got %d", len(doc.Versions))
	}
	if doc.Versions[0].WellFormed {
		t.Error("a regional date is not a well-formed heading")
	}
	if doc.Versions[0].Date != "04/03/2026" {
		t.Errorf("the date must be kept as written, got %q", doc.Versions[0].Date)
	}
	if doc.Versions[1].Label != "" || doc.Versions[1].WellFormed {
		t.Errorf("a heading with no bracketed label is not well-formed: %+v", doc.Versions[1])
	}
}

// A link definition inside a version body must not be mistaken for the
// footer; the footer is the last run.
func TestScanLinkBlockIsTheLastRun(t *testing.T) {
	t.Parallel()
	doc := scan(t, `# Changelog

## [Unreleased]

### Added

- See [the RFC][rfc].

[rfc]: https://example.com/rfc

## [1.0.0] - 2026-01-01

### Added

- Initial release.

[Unreleased]: https://github.com/acme/api/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/acme/api/releases/tag/v1.0.0
`)
	if len(doc.Links) != 2 {
		t.Fatalf("footer should hold 2 definitions, got %d: %+v", len(doc.Links), doc.Links)
	}
	if len(doc.Versions) != 1 {
		t.Fatalf("the mid-document link must not break section scanning, got %d versions", len(doc.Versions))
	}
	if !strings.Contains(doc.Unreleased.Region.Text(doc.Src), "[rfc]: https://example.com/rfc") {
		t.Error("the mid-document link belongs to the Unreleased body")
	}
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b string
		want int
		ok   bool
	}{
		{"1.10.0", "1.9.0", 1, true},
		{"1.9.0", "1.10.0", -1, true},
		{"1.2.0", "1.2.0", 0, true},
		{"1.2", "1.2.0", 0, true},
		{"2.0.0", "1.99.99", 1, true},
		{"1.2.0-rc.1", "1.2.0", 0, true},
		{"next", "1.0.0", 0, false},
		{"", "1.0.0", 0, false},
	}
	for _, c := range cases {
		got, ok := CompareVersions(c.a, c.b)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("CompareVersions(%q, %q) = %d, %v; want %d, %v", c.a, c.b, got, ok, c.want, c.ok)
		}
	}
}
