package changelog

import (
	"errors"
	"strings"
	"testing"
)

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

func release(t *testing.T, src string, opts ReleaseOptions) string {
	t.Helper()
	out, err := Release(scan(t, src), opts)
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	return string(out)
}

func TestReleaseMovesUnreleased(t *testing.T) {
	t.Parallel()
	got := release(t, handEdited, ReleaseOptions{Version: "1.11.0", Date: "2026-06-01"})

	if !strings.Contains(got, "## [1.11.0] - 2026-06-01\n\n### Added\n\n- Retry budget for idempotent calls.\n") {
		t.Errorf("entries did not move under the new heading:\n%s", got)
	}
	// [Unreleased] stays, with nothing under it: the very next heading is
	// the version just cut.
	if !strings.Contains(got, "## [Unreleased]\n\n## [1.11.0]") {
		t.Errorf("[Unreleased] should remain and be empty:\n%s", got)
	}
	// The placeholder-only Fixed subsection must not ship.
	newVersion := got[mustIndex(t, got, "## [1.11.0]"):mustIndex(t, got, "## [1.10.0]")]
	if strings.Contains(newVersion, "### Fixed") {
		t.Errorf("a placeholder-only subsection shipped:\n%s", newVersion)
	}
	if mustIndex(t, got, "## [1.11.0]") > mustIndex(t, got, "## [1.10.0]") {
		t.Error("the new version must precede every previously released one")
	}
}

// The single most important property of the design: outside the regions it
// edits, the file comes out byte for byte as it went in.
func TestReleasePreservesEverythingOutsideEditedRegions(t *testing.T) {
	t.Parallel()
	got := release(t, handEdited, ReleaseOptions{Version: "1.11.0", Date: "2026-06-01"})

	// Preamble: from the start of the file to the [Unreleased] heading.
	before := handEdited[:mustIndex(t, handEdited, "## [Unreleased]")]
	if !strings.HasPrefix(got, before) {
		t.Errorf("preamble changed.\nwant prefix:\n%q\ngot:\n%q", before, got[:len(before)])
	}

	// Previously released versions: from the first old version heading to
	// the link block, byte for byte.
	oldTail := handEdited[mustIndex(t, handEdited, "## [1.10.0]"):mustIndex(t, handEdited, "[Unreleased]: ")]
	if !strings.Contains(got, oldTail) {
		t.Errorf("previously released versions were not preserved verbatim.\nwant:\n%q", oldTail)
	}
}

func TestReleaseYanked(t *testing.T) {
	t.Parallel()
	got := release(t, handEdited, ReleaseOptions{Version: "1.11.0", Date: "2014-12-13", Yanked: true})
	if !strings.Contains(got, "## [1.11.0] - 2014-12-13 [YANKED]\n") {
		t.Errorf("yanked heading wrong:\n%s", got)
	}
}

func TestReleaseRewritesLinkBlock(t *testing.T) {
	t.Parallel()
	got := release(t, handEdited, ReleaseOptions{Version: "1.11.0", Date: "2026-06-01"})
	block := got[mustIndex(t, got, "[Unreleased]: "):]

	want := "[Unreleased]: https://github.com/acme/api/compare/v1.11.0...HEAD\n" +
		"[1.11.0]: https://github.com/acme/api/compare/v1.10.0...v1.11.0\n" +
		"[1.10.0]: https://github.com/acme/api/compare/v1.9.0...v1.10.0\n" +
		"[1.9.0]: https://github.com/acme/api/compare/v0.0.5...v1.9.0\n" +
		"[0.0.5]: https://github.com/acme/api/releases/tag/v0.0.5\n"
	if block != want {
		t.Errorf("link block wrong.\nwant:\n%s\ngot:\n%s", want, block)
	}
}

// A project on a self-hosted host with unprefixed tags keeps its own
// convention, because the convention is read out of the document.
func TestReleasePreservesHostAndTagPrefix(t *testing.T) {
	t.Parallel()
	src := `# Changelog

## [Unreleased]

### Added

- Something.

## [1.1.0] - 2026-01-01

### Fixed

- Something else.

[Unreleased]: https://git.example.com/group/proj/-/compare/1.1.0...HEAD
[1.1.0]: https://git.example.com/group/proj/-/compare/1.0.0...1.1.0
`
	got := release(t, src, ReleaseOptions{Version: "1.2.0", Date: "2026-02-02"})
	if !strings.Contains(got, "[Unreleased]: https://git.example.com/group/proj/-/compare/1.2.0...HEAD\n") {
		t.Errorf("host or tag prefix not preserved:\n%s", got)
	}
	if strings.Contains(got, "/compare/v1.2.0") {
		t.Errorf("a v prefix was invented:\n%s", got)
	}
}

func TestReleaseFirstVersionPointsAtTag(t *testing.T) {
	t.Parallel()
	src := `# Changelog

## [Unreleased]

### Added

- Initial work.

[Unreleased]: https://github.com/acme/api/compare/v0.0.0...HEAD
`
	got := release(t, src, ReleaseOptions{Version: "0.1.0", Date: "2026-02-02"})
	if !strings.Contains(got, "[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0\n") {
		t.Errorf("the first release must point at its tag, not a comparison:\n%s", got)
	}
}

func TestReleaseCreatesLinkBlockFromSlug(t *testing.T) {
	t.Parallel()
	src := `# Changelog

## [Unreleased]

### Added

- Initial work.
`
	got := release(t, src, ReleaseOptions{Version: "0.1.0", Date: "2026-02-02", Slug: "acme/api"})
	if !strings.HasSuffix(got, "\n[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD\n[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0\n") {
		t.Errorf("a link block was not created:\n%s", got)
	}
}

func TestReleaseRefusals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		src  string
		opts ReleaseOptions
		want error
	}{
		{
			name: "no Unreleased heading",
			src:  "# Changelog\n\n## [1.0.0] - 2026-01-01\n\n### Added\n\n- Thing.\n",
			opts: ReleaseOptions{Version: "1.1.0", Date: "2026-02-02"},
			want: ErrNoUnreleased,
		},
		{
			name: "nothing to release",
			src:  "# Changelog\n\n## [Unreleased]\n\n### Added\n\n-\n\n### Fixed\n\n-\n",
			opts: ReleaseOptions{Version: "1.1.0", Date: "2026-02-02"},
			want: ErrNothingToRelease,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Release(scan(t, c.src), c.opts); !errors.Is(err, c.want) {
				t.Fatalf("want %v, got %v", c.want, err)
			}
		})
	}

	t.Run("version already released", func(t *testing.T) {
		t.Parallel()
		_, err := Release(scan(t, handEdited), ReleaseOptions{Version: "1.10.0", Date: "2026-06-01"})
		var already *AlreadyReleasedError
		if !errors.As(err, &already) {
			t.Fatalf("want AlreadyReleasedError, got %v", err)
		}
		if already.Version != "1.10.0" {
			t.Errorf("error must name the version, got %q", already.Version)
		}
	})
}

// Lead prose under [Unreleased] is content, so it moves with the release
// rather than being discarded as unclassifiable.
func TestReleaseCarriesLeadProse(t *testing.T) {
	t.Parallel()
	src := `# Changelog

## [Unreleased]

This release reworks the retry path.

### Added

- Retry budget.

[Unreleased]: https://github.com/acme/api/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/acme/api/releases/tag/v1.0.0
`
	got := release(t, src, ReleaseOptions{Version: "1.1.0", Date: "2026-02-02"})
	if !strings.Contains(got, "## [1.1.0] - 2026-02-02\n\nThis release reworks the retry path.\n\n### Added\n") {
		t.Errorf("lead prose was not carried into the release:\n%s", got)
	}
	if strings.Contains(got[:mustIndex(t, got, "## [1.1.0]")], "reworks the retry path") {
		t.Error("lead prose was left behind under [Unreleased]")
	}
}
