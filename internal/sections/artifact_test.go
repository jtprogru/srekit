package sections

import (
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

const validArtifact = `
version: 1
frontmatter:
  id: "{{ .Meta.ID }}"
  type: postmortem
  tags: [postmortem, incident]
title: "Post — {{ .Meta.Title }}"
meta_bullets:
  - "**Severity:** {{ .Meta.Severity }}"
header_body: ""
sections:
  - id: summary
    title: Summary
    type: text
    required: true
    default_body: _placeholder_
  - id: timeline
    title: Timeline
    type: table
    columns: [t, event]
`

func TestParseArtifact_Happy(t *testing.T) {
	t.Parallel()
	a, err := ParseArtifact([]byte(validArtifact))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.Version != 1 {
		t.Errorf("version=%d", a.Version)
	}
	if a.Title != "Post — {{ .Meta.Title }}" {
		t.Errorf("title=%q", a.Title)
	}
	if len(a.MetaBullets) != 1 {
		t.Errorf("meta_bullets len=%d", len(a.MetaBullets))
	}
	if len(a.Sections) != 2 || a.Sections[0].ID != "summary" {
		t.Errorf("sections wrong: %+v", a.Sections)
	}
	if a.Frontmatter.Kind != yaml.MappingNode {
		t.Errorf("frontmatter should be MappingNode, got kind %d", a.Frontmatter.Kind)
	}
}

func TestParseArtifact_PreservesFrontmatterKeyOrder(t *testing.T) {
	t.Parallel()
	a, err := ParseArtifact([]byte(validArtifact))
	if err != nil {
		t.Fatal(err)
	}
	// Frontmatter mapping nodes alternate Key, Value, Key, Value, ...
	want := []string{"id", "type", "tags"}
	got := make([]string, 0, len(a.Frontmatter.Content)/2)
	for i := 0; i < len(a.Frontmatter.Content); i += 2 {
		got = append(got, a.Frontmatter.Content[i].Value)
	}
	if len(got) != len(want) {
		t.Fatalf("keys len=%d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("key[%d]=%q, want %q (full order: %v)", i, got[i], want[i], got)
		}
	}
}

func TestParseArtifact_Rejects(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		yaml    string
		wantSub string
	}{
		{
			name:    "wrong version",
			yaml:    "version: 2\nsections:\n  - id: x\n    title: X\n    type: text\n",
			wantSub: "unsupported artifact version",
		},
		{
			name:    "no sections",
			yaml:    "version: 1\nsections: []\n",
			wantSub: "no sections",
		},
		{
			name:    "frontmatter is a scalar",
			yaml:    "version: 1\nfrontmatter: not-a-map\nsections:\n  - id: x\n    title: X\n    type: text\n",
			wantSub: "frontmatter must be a YAML mapping",
		},
		{
			name:    "section bad type",
			yaml:    "version: 1\nsections:\n  - id: x\n    title: X\n    type: image\n",
			wantSub: "unknown type",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseArtifact([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("want error containing %q", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("want %q in error, got: %v", tc.wantSub, err)
			}
		})
	}
}

func TestArtifact_SectionIDs(t *testing.T) {
	t.Parallel()
	a, err := ParseArtifact([]byte(validArtifact))
	if err != nil {
		t.Fatal(err)
	}
	got := a.SectionIDs()
	if len(got) != 2 || got[0] != "summary" || got[1] != "timeline" {
		t.Errorf("SectionIDs=%v", got)
	}
}
