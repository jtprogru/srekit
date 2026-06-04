package sections

import (
	"strings"
	"testing"
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
