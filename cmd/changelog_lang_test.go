package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
)

// englishChangelog is what `srekit changelog --repo acme/api --stdout`
// emitted before `--lang` existed. The default output is a compatibility
// surface — everything that greps `### Added` depends on it — so it is
// pinned in full rather than probed for substrings.
const englishChangelog = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

-

### Security

-

## [0.1.0] - 2026-01-01

### Added

- Initial release.

[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`

// pinToday fixes the wall clock so the generated version heading is stable.
func pinToday(t *testing.T) {
	t.Helper()
	orig := clock.Now
	t.Cleanup(func() { clock.Now = orig })
	clock.Now = func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }
}

func TestChangelogDefaultIsUnchangedEnglish(t *testing.T) {
	withConfig(t, nil)
	pinToday(t)
	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if out != englishChangelog {
		t.Errorf("the unqualified invocation must be byte-identical to the pre--lang output.\ngot:\n%s\nwant:\n%s", out, englishChangelog)
	}
}

func TestChangelogLangRU(t *testing.T) {
	withConfig(t, nil)
	pinToday(t)
	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--lang", "ru", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"### Добавлено", "### Изменено", "### Устарело",
		"### Удалено", "### Исправлено", "### Безопасность",
		"https://keepachangelog.com/ru/1.1.0/",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Russian variant missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "### Added") {
		t.Errorf("English change types should be gone:\n%s", out)
	}
}

// The translation stops where the document stops being prose: the heading
// and the reference label that points at it are two halves of one link, and
// the label is the part that points outward, at tags and compare URLs.
func TestChangelogLangRUKeepsIdentifiersEnglish(t *testing.T) {
	withConfig(t, nil)
	pinToday(t)
	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--lang", "ru", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## [Unreleased]",
		"## [0.1.0] - 2026-01-01",
		"[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD",
		"[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("identifier %q must stay English:\n%s", want, out)
		}
	}
}

func TestChangelogLangFromConfig(t *testing.T) {
	withConfig(t, map[string]string{"changelog_lang": "ru"})
	pinToday(t)
	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "### Добавлено") {
		t.Errorf("changelog_lang: ru should select the variant:\n%s", out)
	}
}

func TestChangelogLangFlagBeatsConfig(t *testing.T) {
	withConfig(t, map[string]string{"changelog_lang": "ru"})
	pinToday(t)
	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--lang", "en", "--stdout")
	if err != nil {
		t.Fatal(err)
	}
	if out != englishChangelog {
		t.Errorf("--lang en must override changelog_lang: ru:\n%s", out)
	}
}

func TestChangelogUnknownLangRejected(t *testing.T) {
	withConfig(t, nil)
	dir := t.TempDir()
	t.Chdir(dir)

	out, err := runCLI(t, "changelog", "--repo", "acme/api", "--lang", "de")
	if err == nil {
		t.Fatalf("expected a non-zero exit:\n%s", out)
	}
	for _, want := range []string{"en", "ru"} {
		if !strings.Contains(out, want) {
			t.Errorf("the error should name %q as accepted:\n%s", want, out)
		}
	}
	// The rejection happens before anything is written.
	if _, statErr := os.Stat(filepath.Join(dir, "CHANGELOG.md")); statErr == nil {
		t.Error("no file should be created when the language is rejected")
	}
}

func TestChangelogUnknownLangInConfigRejected(t *testing.T) {
	withConfig(t, map[string]string{"changelog_lang": "de"})
	dir := t.TempDir()
	t.Chdir(dir)

	out, err := runCLI(t, "changelog", "--repo", "acme/api")
	if err == nil {
		t.Fatalf("expected a non-zero exit:\n%s", out)
	}
	if !strings.Contains(out, "changelog_lang") {
		t.Errorf("the error should name the setting it came from:\n%s", out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "CHANGELOG.md")); statErr == nil {
		t.Error("no file should be created when the language is rejected")
	}
}

// russianScaffold is a Russian changelog with one real entry, the shape a
// release is actually cut from.
const russianScaffold = `# Changelog

Все заметные изменения в этом проекте документируются в этом файле.

## [Unreleased]

### Добавлено

-

### Исправлено

- Гонка в пуле соединений.

## [0.1.0] - 2026-01-01

### Добавлено

- Первый релиз.

[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`

func TestChangelogReleasePreservesRussianVocabulary(t *testing.T) {
	read := writeChangelog(t, russianScaffold)
	if _, err := runCLI(t, "changelog", "release", "--version", "0.2.0", "--date", "2026-03-04"); err != nil {
		t.Fatal(err)
	}
	got := read()
	if !strings.Contains(got, "## [0.2.0] - 2026-03-04\n\n### Исправлено\n\n- Гонка в пуле соединений.") {
		t.Errorf("the released version should carry the Russian change type:\n%s", got)
	}
	// The empty placeholder is pruned in either vocabulary.
	if strings.Contains(got, "## [0.2.0] - 2026-03-04\n\n### Добавлено") {
		t.Errorf("an empty Russian placeholder should be dropped:\n%s", got)
	}
}

// Generation and parsing run in opposite directions: a team that switched
// to Russian still has an English CHANGELOG.md from before the switch, and
// release must not translate it out from under them.
func TestChangelogReleaseIgnoresLangWhenParsing(t *testing.T) {
	read := writeChangelog(t, releasedScaffold)
	if _, err := runCLI(t, "changelog", "release", "--version", "0.2.0", "--date", "2026-03-04", "--lang", "ru"); err != nil {
		t.Fatal(err)
	}
	got := read()
	if !strings.Contains(got, "### Fixed") {
		t.Errorf("--lang ru must not touch an English document's change types:\n%s", got)
	}
	if strings.Contains(got, "Исправлено") {
		t.Errorf("no Russian heading should appear in an English document:\n%s", got)
	}
}

// mixedScaffold is a half-translated changelog: `### Added` and
// `### Исправлено` in one document, where neither a reader nor a tool can
// group the entries.
const mixedScaffold = `# Changelog

## [Unreleased]

### Added

- A new endpoint.

### Исправлено

- Гонка в пуле соединений.

## [0.1.0] - 2026-01-01

### Added

- Initial release.

[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`

func TestChangelogValidateRejectsMixedVocabularies(t *testing.T) {
	writeChangelog(t, mixedScaffold)
	out, err := runCLI(t, "changelog", "validate")
	if err == nil {
		t.Fatalf("expected a non-zero exit:\n%s", out)
	}
	if !strings.Contains(out, "FAIL  change-type-language") {
		t.Errorf("the mixed document should fail the vocabulary check:\n%s", out)
	}
	for _, want := range []string{`"Added"`, `"Исправлено"`} {
		if !strings.Contains(out, want) {
			t.Errorf("the detail should name %s:\n%s", want, out)
		}
	}
}

func TestChangelogReleaseRefusesMixedVocabularies(t *testing.T) {
	read := writeChangelog(t, mixedScaffold)
	out, err := runCLI(t, "changelog", "release", "--version", "0.2.0")
	if err == nil {
		t.Fatalf("expected a non-zero exit:\n%s", out)
	}
	if !strings.Contains(out, "mixes change-type vocabularies") {
		t.Errorf("the error should explain the mix:\n%s", out)
	}
	if got := read(); got != mixedScaffold {
		t.Errorf("a refused release must leave the file byte-identical:\n%s", got)
	}
}

func TestChangelogValidateReportsDetectedVocabulary(t *testing.T) {
	writeChangelog(t, russianScaffold)
	out, err := runCLI(t, "changelog", "validate")
	if err != nil {
		t.Fatalf("unexpected failure:\n%s", out)
	}
	if !strings.Contains(out, "OK    change-type-language: Russian (ru)") {
		t.Errorf("validate should name the vocabulary it used:\n%s", out)
	}
}

func TestChangelogValidateNoChangeTypesAtAll(t *testing.T) {
	writeChangelog(t, `# Changelog

## [Unreleased]

## [0.1.0] - 2026-01-01

- Initial release.

[Unreleased]: https://github.com/acme/api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/acme/api/releases/tag/v0.1.0
`)
	out, err := runCLI(t, "changelog", "validate")
	if err == nil {
		t.Fatalf("expected a non-zero exit:\n%s", out)
	}
	if !strings.Contains(out, "FAIL  change-type-language: no recognized change types") {
		t.Errorf("a document with no change types is neither language:\n%s", out)
	}
}

// The variant is a file like any other in the embedded set: it is listed,
// scaffolded, snapshotted and upgraded on its own, never conflated with the
// base artifact it falls back to.
func TestTemplatesListsChangelogVariantSeparately(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "templates", "list", "--templates-dir", dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"changelog.yaml", "changelog.ru.yaml"} {
		if !strings.Contains(out, name) {
			t.Errorf("%s should be listed as its own entry:\n%s", name, out)
		}
		if _, statErr := os.Stat(filepath.Join(dir, ".srekit-embedded", name)); statErr != nil {
			t.Errorf("%s should have its own snapshot: %v", name, statErr)
		}
	}
}

func TestTemplatesUpgradeVariantIndependently(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Parallel()
	dir := t.TempDir()
	if _, err := runCLI(t, "templates", "init", dir, "--no-git"); err != nil {
		t.Fatal(err)
	}
	const name = "changelog.ru.yaml"
	variant := filepath.Join(dir, name)
	base := filepath.Join(dir, "changelog.yaml")

	baseBefore, err := os.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	embeddedVariant, err := os.ReadFile(variant)
	if err != nil {
		t.Fatal(err)
	}

	// The same disjoint-hunks setup as TestTemplatesUpgrade3WayCleanMerge,
	// applied to the variant: its snapshot carries a line upstream has since
	// dropped, and the user's edit sits at the other end of the file. The
	// merge base is the variant's own snapshot — nothing here involves the
	// English artifact.
	snapBody := append(append([]byte{}, embeddedVariant...), []byte("EXTRA_BOTTOM\n")...)
	userBody := append([]byte("# локальная правка\n"), snapBody...)
	if err := os.WriteFile(filepath.Join(dir, ".srekit-embedded", name), snapBody, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(variant, userBody, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCLI(t, "templates", "upgrade", dir)
	if err != nil {
		t.Fatalf("upgrade failed: %v (output: %s)", err, out)
	}
	if !strings.Contains(out, "~ merged    "+name) {
		t.Errorf("the variant should be merged against its own snapshot, got: %s", out)
	}
	merged, _ := os.ReadFile(variant)
	if !bytes.Contains(merged, []byte("# локальная правка")) {
		t.Errorf("the merge should preserve the user's edit:\n%s", merged)
	}
	if bytes.Contains(merged, []byte("EXTRA_BOTTOM")) {
		t.Errorf("the merge should drop what upstream removed:\n%s", merged)
	}
	if !bytes.Contains(merged, []byte("### Добавлено")) {
		t.Errorf("the merged variant should still be the Russian artifact:\n%s", merged)
	}
	if got, _ := os.ReadFile(base); !bytes.Equal(got, baseBefore) {
		t.Errorf("the untouched base artifact should be unaffected:\n%s", got)
	}
	snap, _ := os.ReadFile(filepath.Join(dir, ".srekit-embedded", name))
	if !bytes.Equal(snap, embeddedVariant) {
		t.Error("the variant's snapshot should advance to the current embedded variant")
	}
}
