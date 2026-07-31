package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// The output-routing tests below exercise writeBody (--out / --stdout /
// --force / --dry-run / permissions), not artifact composition, so they
// use the smallest possible v1 artifact rather than a shipped one. The
// loader is built from a per-test fixture dir, decoupling the suite from
// embed contents.

const fixtureArtifact = `version: 1
title: "Fixture: {{ .Meta.Title }}"
sections:
  - id: body
    title: Body
    type: text
    body: "{{ .Meta.Body }}"
`

// newFixtureLoader returns a Loader that resolves the `fixture` artifact
// from a temp dir, falling back to embed for anything else.
func newFixtureLoader(t *testing.T) *tmpl.Loader {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fixture.yaml"), []byte(fixtureArtifact), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return &tmpl.Loader{Sources: []tmpl.Source{tmpl.DirSource{Dir: dir}, tmpl.EmbedSource{}}}
}

// fixturePayload is the data the fixture artifact expects. Its fields are
// exported so the bootstrap-JSON envelope, which marshals the data value
// itself as `meta`, sees them.
type fixturePayload struct {
	Title string
	Body  string
}

func (f fixturePayload) ArtifactPayload() ([]sections.RenderedSection, any) {
	return []sections.RenderedSection{
			{ID: "body", Title: "Body", Type: "text", Body: f.Body, Required: true},
		},
		struct{ Meta fixturePayload }{Meta: f}
}

func newFixtureData(title, body string) fixturePayload {
	return fixturePayload{Title: title, Body: body}
}

func TestRenderStdout(t *testing.T) {
	var out bytes.Buffer
	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("Hello", "world"), Options{Stdout: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "# Fixture: Hello") {
		t.Fatalf("missing title in output: %s", out.String())
	}
}

func TestRenderToFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "out.md")
	var out bytes.Buffer

	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("Foo", "x"), Options{Out: target})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Foo") {
		t.Fatalf("file missing content")
	}
}

func TestRenderRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("x", "y"), Options{Out: target})
	if err == nil {
		t.Fatal("expected error on existing file without --force")
	}
}

func TestRenderForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	_ = os.WriteFile(target, []byte("old"), 0o644)
	var out bytes.Buffer
	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("Force", "y"), Options{Out: target, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(target)
	if !strings.Contains(string(b), "Force") {
		t.Fatalf("file not overwritten")
	}
}

// TestRenderFilePermissions locks in the #9 fix: generated docs land at
// 0o644, not 0o600. These are public artifacts (LICENSE, CHANGELOG, etc.),
// not secrets.
func TestRenderFilePermissions(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	var out bytes.Buffer
	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("Perms", "y"), Options{Out: target})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("expected mode 0o644, got %o", perm)
	}
}

// TestRenderTemplatePathOverride was removed in v0.30.0 along with
// opts.TemplatePath: no render path reads a template from an arbitrary
// file any more.

// TestRenderJSONStdout verifies --json short-circuits the template and emits
// MarshalIndent'd JSON to stdout when --out/--stdout are unset.
func TestRenderJSONStdout(t *testing.T) {
	var out bytes.Buffer
	type payload struct {
		ID    string
		Title string
	}
	in := payload{ID: "abc", Title: "Hello"}
	err := Render(&out, tmpl.NewDefaultLoader(), "fixture", in, Options{JSON: true, Default: "ignored.md"})
	if err != nil {
		t.Fatal(err)
	}
	var got payload
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if got != in {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, in)
	}
	if !strings.Contains(out.String(), "\n  \"ID\"") {
		t.Fatalf("expected MarshalIndent (2-space) output, got: %s", out.String())
	}
}

// TestRenderJSONToFile verifies --json honors --out.
func TestRenderJSONToFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")
	var out bytes.Buffer
	err := Render(&out, tmpl.NewDefaultLoader(), "fixture",
		struct{ Title string }{Title: "Foo"},
		Options{Out: target, JSON: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("file is not valid JSON: %v\n%s", err, string(b))
	}
	if got["Title"] != "Foo" {
		t.Fatalf("Title mismatch: %v", got["Title"])
	}
}

// TestRenderJSONIgnoresDefaultPath locks in the rule that --json without
// --out/--stdout sinks to stdout rather than the markdown Default path,
// so a JSON payload never ends up in a .md file by accident.
func TestRenderJSONIgnoresDefaultPath(t *testing.T) {
	dir := t.TempDir()
	defaultPath := filepath.Join(dir, "should-not-be-written.md")
	var out bytes.Buffer
	err := Render(&out, tmpl.NewDefaultLoader(), "fixture",
		struct{ Title string }{Title: "X"},
		Options{JSON: true, Default: defaultPath},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		t.Fatalf("--json must not write to the markdown default path: %s", defaultPath)
	}
	if !strings.Contains(out.String(), "\"Title\"") {
		t.Fatalf("expected JSON on stdout, got: %s", out.String())
	}
}

// TestRenderBootstrapJSONEnvelope was removed in v0.30.0 with the
// envelope it covered: nothing had set BootstrapJSON since v0.20.0, when
// the last generator moved to the artifact path.

func TestRenderStructuredJSONPassThrough(t *testing.T) {
	// --json marshals the caller's payload directly, with no wrapping.
	var out bytes.Buffer
	in := map[string]any{
		"meta":     map[string]any{"title": "X"},
		"sections": []any{map[string]any{"id": "summary", "body": "hi"}},
	}
	err := Render(&out, tmpl.NewDefaultLoader(), "fixture", in,
		Options{JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	meta := got["meta"].(map[string]any)
	if meta["title"] != "X" {
		t.Errorf("structured pass-through dropped data: %v", got)
	}
	// Sections must NOT be re-wrapped as bootstrap envelope.
	secs := got["sections"].([]any)
	if len(secs) != 1 {
		t.Fatalf("expected one section, got %d", len(secs))
	}
	s0 := secs[0].(map[string]any)
	if s0["id"] != "summary" {
		t.Errorf("structured path should pass sections through: %v", s0)
	}
}

// TestRenderArtifactPath_Postmortem verifies the artifact render path:
// the loader resolves the embedded postmortem.yaml, parses it, and
// composes the markdown via sections.RenderArtifact using the data's
// ArtifactPayload() method.
func TestRenderArtifactPath_Postmortem(t *testing.T) {
	type meta struct {
		ID, Title, Severity, Owner, Start, End, Now string
	}
	m := meta{ID: "abc", Title: "Cache stampede", Severity: "SEV-1", Owner: "@oncall"}
	rendered, err := sections.Merge(
		mustLoadEmbeddedPostmortemManifest(t),
		nil,
		struct{ Meta meta }{Meta: m},
	)
	if err != nil {
		t.Fatal(err)
	}
	data := &artifactDataStub{meta: m, sections: rendered}

	var out bytes.Buffer
	if err := Render(&out, tmpl.NewDefaultLoader(), "postmortem.md.tmpl", data,
		Options{Stdout: true}); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"---\n",     // frontmatter open
		`id: "abc"`, // frontmatter eval
		"# Постмортем (Postmortem) — Cache stampede", // H1 eval
		"**Тяжесть (Severity):** SEV-1",              // meta_bullet eval
		"## Краткое описание (Summary)",              // first section
		"## Хронология (Timeline)",                   // table section
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in artifact-rendered output:\n%s", want, got)
		}
	}
}

func TestRenderArtifactPath_DataMissingInterface(t *testing.T) {
	var out bytes.Buffer
	err := Render(&out, tmpl.NewDefaultLoader(), "postmortem.md.tmpl",
		struct{ Foo string }{Foo: "bar"},
		Options{Stdout: true})
	if err == nil {
		t.Fatal("expected error when data type lacks ArtifactPayload")
	}
	if !strings.Contains(err.Error(), "ArtifactPayload") {
		t.Errorf("error should mention ArtifactPayload, got: %v", err)
	}
}

// mustLoadEmbeddedPostmortemManifest reads the embedded postmortem.yaml
// and returns a Manifest matching its sections list. Helper for tests.
func mustLoadEmbeddedPostmortemManifest(t *testing.T) *sections.Manifest {
	t.Helper()
	body, err := tmpl.NewDefaultLoader().LoadArtifactBytes("postmortem")
	if err != nil {
		t.Fatalf("load artifact: %v", err)
	}
	a, err := sections.ParseArtifact(body)
	if err != nil {
		t.Fatalf("parse artifact: %v", err)
	}
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}
}

// artifactDataStub implements ArtifactPayload for tests; minimal shape.
type artifactDataStub struct {
	meta     any
	sections []sections.RenderedSection
}

func (a *artifactDataStub) ArtifactPayload() ([]sections.RenderedSection, any) {
	return a.sections, struct{ Meta any }{Meta: a.meta}
}

// TestExtractH1 was removed in v0.30.0 with extractH1, whose only caller
// was the bootstrap-JSON envelope.

func TestRenderDryRunNoFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	var out bytes.Buffer
	err := Render(&out, newFixtureLoader(t), "fixture", newFixtureData("Dry", "x"), Options{Out: target, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create file")
	}
	if !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("dry-run header missing")
	}
}
