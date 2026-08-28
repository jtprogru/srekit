package sections

import (
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestRenderArtifact_FullDocument(t *testing.T) {
	t.Parallel()
	a, err := ParseArtifact([]byte(validArtifact))
	if err != nil {
		t.Fatal(err)
	}
	var ctx renderCtx
	ctx.Meta.Start = "2026-06-04T10:00:00Z"
	ctx.Meta.End = "2026-06-04T11:00:00Z"

	// Use type pun with the existing renderCtx + simulate Meta.X via separate ctx:
	type meta struct {
		ID       string
		Title    string
		Severity string
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, struct{ Meta meta }{
		Meta: meta{ID: "abc", Title: "X", Severity: "SEV-1"},
	})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	body, err := RenderArtifact(a, rendered, struct{ Meta meta }{Meta: meta{
		ID: "abc", Title: "X", Severity: "SEV-1",
	}})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	got := string(body)

	for _, want := range []string{
		"---\n",               // frontmatter open
		`id: "abc"`,           // frontmatter eval (quoted because source was quoted)
		"type: postmortem",    // literal
		"# Post — X",          // title eval
		"**Severity:** SEV-1", // meta_bullet eval
		"## Summary",          // first section heading
		"_placeholder_",       // first section body
		"## Timeline",         // second section heading
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}

	// Frontmatter keys appear in source order (not alphabetical).
	idIdx := strings.Index(got, "id:")
	typeIdx := strings.Index(got, "type:")
	tagsIdx := strings.Index(got, "tags:")
	if idIdx >= typeIdx || typeIdx >= tagsIdx {
		t.Errorf("frontmatter key order not preserved: id@%d type@%d tags@%d", idIdx, typeIdx, tagsIdx)
	}

	// No double blank lines piling up.
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("triple newlines in output suggest spacing bug:\n%s", got)
	}
}

func TestRenderArtifact_SkipsEmptyFields(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version: 1,
		// No frontmatter, no title, no bullets, no header_body.
		Sections: []Section{
			{ID: "x", Title: "X", Type: TypeText, DefaultBody: "body"},
		},
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err := RenderArtifact(a, rendered, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if strings.Contains(got, "---") {
		t.Errorf("empty frontmatter should not produce --- block:\n%s", got)
	}
	if strings.HasPrefix(got, "# ") {
		t.Errorf("empty title should not produce H1:\n%s", got)
	}
	if !strings.HasPrefix(got, "\n## X") && !strings.HasPrefix(got, "## X") {
		t.Errorf("first content should be the section header:\n%s", got)
	}
}

func TestRenderArtifact_DoesNotMutateInput(t *testing.T) {
	t.Parallel()
	a, err := ParseArtifact([]byte(validArtifact))
	if err != nil {
		t.Fatal(err)
	}
	type meta struct {
		ID       string
		Title    string
		Severity string
	}
	rendered, _ := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)

	// Capture frontmatter "id" value before render.
	before := frontmatterScalar(a, "id")

	if _, err := RenderArtifact(a, rendered, struct{ Meta meta }{Meta: meta{
		ID: "abc", Title: "X", Severity: "SEV-1",
	}}); err != nil {
		t.Fatal(err)
	}
	after := frontmatterScalar(a, "id")
	if before != after {
		t.Errorf("Frontmatter mutated by RenderArtifact: before=%q after=%q", before, after)
	}
}

// frontmatterScalar fetches a top-level frontmatter value by key for tests.
func frontmatterScalar(a *Artifact, key string) string {
	for i := 0; i < len(a.Frontmatter.Content); i += 2 {
		if a.Frontmatter.Content[i].Value == key {
			return a.Frontmatter.Content[i+1].Value
		}
	}
	return ""
}

// renderOne is the boilerplate every composition test repeats: merge the
// artifact's own sections with no overrides, then render against ctx.
func renderOne(t *testing.T, a *Artifact, ctx any) string {
	t.Helper()
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, ctx)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	body, err := RenderArtifact(a, rendered, ctx)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return string(body)
}

func TestRenderArtifact_FooterBody(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:    1,
		Title:      "Changelog",
		Sections:   []Section{{ID: "x", Title: "[Unreleased]", Type: TypeText, DefaultBody: "entry"}},
		FooterBody: "[Unreleased]: https://example.test/compare/v1.0.0...HEAD\n",
	}
	got := renderOne(t, a, nil)
	want := "# Changelog\n\n## [Unreleased]\n\nentry\n\n[Unreleased]: https://example.test/compare/v1.0.0...HEAD\n"
	if got != want {
		t.Errorf("footer composition:\ngot  %q\nwant %q", got, want)
	}
}

func TestRenderArtifact_FooterBodyAbsentIsUnchanged(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:  1,
		Title:    "T",
		Sections: []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
	}
	got := renderOne(t, a, nil)
	want := "# T\n\n## X\n\nb\n"
	if got != want {
		t.Errorf("artifact without a footer must render as it did before the key existed:\ngot  %q\nwant %q", got, want)
	}
}

func TestRenderArtifact_FooterBodyEvaluated(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:    1,
		Sections:   []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
		FooterBody: "[link]: https://example.test/{{ .Meta.Repo }}",
	}
	type meta struct{ Repo string }
	got := renderOne(t, a, struct{ Meta meta }{Meta: meta{Repo: "acme/api"}})
	if !strings.Contains(got, "[link]: https://example.test/acme/api") {
		t.Errorf("footer_body not expanded:\n%s", got)
	}
}

func TestRenderArtifact_FooterBodyWhitespaceOnly(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:    1,
		Sections:   []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
		FooterBody: "   \n\n  ",
	}
	got := renderOne(t, a, nil)
	if got != "## X\n\nb\n" {
		t.Errorf("whitespace-only footer should produce no trailing stanza: %q", got)
	}
}

// TestRenderArtifact_BlockSeparation pins the invariant that every pair of
// adjacent blocks is separated by exactly one blank line, across the
// combinations of present and absent elements. The title+header_body case
// is the one that used to emit two.
func TestRenderArtifact_BlockSeparation(t *testing.T) {
	t.Parallel()
	fm := mappingNode(t, "id: abc\n")
	section := []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}}

	cases := []struct {
		name string
		a    *Artifact
		want string
	}{
		{
			name: "title and header body",
			a:    &Artifact{Version: 1, Title: "T", HeaderBody: "intro", Sections: section},
			want: "# T\n\nintro\n\n## X\n\nb\n",
		},
		{
			name: "frontmatter and header body without a title",
			a:    &Artifact{Version: 1, Frontmatter: fm, HeaderBody: "intro", Sections: section},
			want: "---\nid: abc\n---\n\nintro\n\n## X\n\nb\n",
		},
		{
			name: "meta bullets and header body",
			a:    &Artifact{Version: 1, MetaBullets: []string{"a", "c"}, HeaderBody: "intro", Sections: section},
			want: "- a\n- c\n\nintro\n\n## X\n\nb\n",
		},
		{
			name: "sections and footer",
			a:    &Artifact{Version: 1, Sections: section, FooterBody: "[l]: u"},
			want: "## X\n\nb\n\n[l]: u\n",
		},
		{
			name: "every element present",
			a: &Artifact{
				Version: 1, Frontmatter: fm, Title: "T", MetaBullets: []string{"a"},
				HeaderBody: "intro", Sections: section, FooterBody: "[l]: u",
			},
			want: "---\nid: abc\n---\n\n# T\n\n- a\n\nintro\n\n## X\n\nb\n\n[l]: u\n",
		},
		{
			name: "sections only",
			a:    &Artifact{Version: 1, Sections: section},
			want: "## X\n\nb\n",
		},
		{
			name: "title only",
			a:    &Artifact{Version: 1, Title: "T", Sections: section},
			want: "# T\n\n## X\n\nb\n",
		},
		{
			name: "two sections",
			a: &Artifact{Version: 1, Sections: []Section{
				{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"},
				{ID: "y", Title: "Y", Type: TypeText, DefaultBody: "c"},
			}},
			want: "## X\n\nb\n\n## Y\n\nc\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := renderOne(t, tc.a, nil)
			if got != tc.want {
				t.Errorf("composition:\ngot  %q\nwant %q", got, tc.want)
			}
			if strings.Contains(got, "\n\n\n") {
				t.Errorf("two consecutive blank lines: %q", got)
			}
			if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
				t.Errorf("document must end with exactly one newline: %q", got)
			}
		})
	}
}

// mappingNode parses a YAML fragment into the mapping node shape the
// Artifact's Frontmatter field expects.
func mappingNode(t *testing.T, src string) yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatalf("parse frontmatter fixture: %v", err)
	}
	return *doc.Content[0]
}

func TestRenderArtifact_HeaderBodyEvaluated(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:    1,
		Title:      "T",
		HeaderBody: "> note for {{ .Meta.Title }}",
		Sections:   []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	type meta struct{ Title string }
	body, err := RenderArtifact(a, rendered, struct{ Meta meta }{Meta: meta{Title: "Ouch"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "> note for Ouch") {
		t.Errorf("header_body not expanded:\n%s", body)
	}
}

// TestRenderArtifact_TypedFrontmatterValues covers the explicitly tagged
// frontmatter scalar: a templated value is a Go string, so without a tag
// it can only come out quoted. `!!int` and `!!seq` say "read what this
// rendered into", which is what the tools that parse frontmatter care
// about — `duration: 30` is a number, `duration: "30"` is not.
func TestRenderArtifact_TypedFrontmatterValues(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version: 1,
		Frontmatter: mappingNode(t, `
topic: "{{ .Meta.Topic }}"
level: !!seq '[{{ .Meta.Level | join ", " }}]'
duration: !!int "{{ .Meta.Duration }}"
`),
		Sections: []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	type meta struct {
		Topic    string
		Level    []string
		Duration int
	}
	body, err := RenderArtifact(a, rendered, struct{ Meta meta }{
		Meta: meta{Topic: "go", Level: []string{"middle", "senior"}, Duration: 30},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{
		"level: [middle, senior]",
		"duration: 30",
		`topic: "go"`, // untagged: still a quoted string, as before
	} {
		if !strings.Contains(got, want) {
			t.Errorf("frontmatter missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "!!") {
		t.Errorf("the tag is an authoring instruction, it must not reach the document:\n%s", got)
	}
}

// TestRenderArtifact_TypedFrontmatterMismatch pins the diagnostic: a
// declared type the rendered text does not have fails at render time,
// with the frontmatter key named — not silently emitted as a string.
func TestRenderArtifact_TypedFrontmatterMismatch(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:     1,
		Frontmatter: mappingNode(t, `duration: !!int "{{ .Meta.Duration }}"`),
		Sections:    []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	type meta struct{ Duration string }
	_, err = RenderArtifact(a, rendered, struct{ Meta meta }{Meta: meta{Duration: "half an hour"}})
	if err == nil {
		t.Fatal("expected a type error for a non-numeric !!int value")
	}
	for _, want := range []string{"duration", "!!int"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}
}

// TestRenderArtifact_UnknownTagIsLeftAlone guards the conservative half
// of the rule: srekit retypes the tags YAML itself defines, and leaves an
// application tag's payload to the application that put it there.
func TestRenderArtifact_UnknownTagIsLeftAlone(t *testing.T) {
	t.Parallel()
	a := &Artifact{
		Version:     1,
		Frontmatter: mappingNode(t, `ref: !Ref "{{ .Meta.Name }}"`),
		Sections:    []Section{{ID: "x", Title: "X", Type: TypeText, DefaultBody: "b"}},
	}
	rendered, err := Merge(&Manifest{Version: 1, Sections: a.Sections}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	type meta struct{ Name string }
	body, err := RenderArtifact(a, rendered, struct{ Meta meta }{Meta: meta{Name: "bucket"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `!Ref "bucket"`) {
		t.Errorf("custom tag and its rendered payload should survive:\n%s", body)
	}
}
