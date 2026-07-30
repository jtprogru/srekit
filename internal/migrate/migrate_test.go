package migrate

import (
	"strings"
	"testing"
)

// minimalTmpl is the smallest end-to-end input: frontmatter + H1 +
// meta_bullets + two sections (one text, one table).
const minimalTmpl = `---
id: {{ .ID }}
type: investigation
title: {{ .Title }}
---

# Investigation — {{ .Title }}

- **Severity:** {{ .Severity }}
- **Owner:** {{ .Owner | default "<owner>" }}

## Context

_What raised this investigation?_

## Action items

| # | Action | Owner |
|---|--------|-------|
| 1 |        |       |
`

func TestConvert_HappyPath(t *testing.T) {
	t.Parallel()
	got, err := Convert([]byte(minimalTmpl), nil)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	out := string(got)

	for _, want := range []string{
		"version: 1",
		"frontmatter:",
		`id: "{{ .ID }}"`, // pre-quoted so yaml.v3 doesn't mis-parse as flow mapping
		"type: investigation",
		`title: "Investigation — {{ .Title }}"`,
		`- "**Severity:** {{ .Severity }}"`,
		"- id: context",
		"type: text",
		"_What raised this investigation?_",
		"- id: action_items",
		"type: table",
		"columns:",
		"- '#'", // yaml.v3 single-quotes pure-digit/symbol strings
		"- Action",
		"- Owner",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in converted YAML:\n%s", want, out)
		}
	}
}

func TestConvert_SectionsManifestOverride(t *testing.T) {
	t.Parallel()
	manifest := []byte(`version: 1
sections:
  - id: only
    title: Only
    type: text
    required: true
    default_body: from-manifest
`)
	got, err := Convert([]byte(minimalTmpl), manifest)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if !strings.Contains(out, "- id: only") {
		t.Errorf("manifest override not applied; expected single section 'only':\n%s", out)
	}
	if strings.Contains(out, "- id: context") {
		t.Errorf(".tmpl sections should be ignored when manifest provided:\n%s", out)
	}
}

// TestConvert_SidecarSuppressesHeaderBody covers a v0.13.x postmortem.md.tmpl
// shape: header + meta_bullets + `{{ range .Sections }}{{ .Body }}{{ end }}`
// (no `## ` blocks in the .tmpl itself — sections live in the sidecar). The
// migrator must not capture that range loop as header_body; doing so would
// emit `{{ range .Sections }}` into the v1 yaml and break the renderer.
// Regression for the v0.26.0 smoke-audit finding.
func TestConvert_SidecarSuppressesHeaderBody(t *testing.T) {
	t.Parallel()
	tmpl := `---
id: {{ .Meta.ID }}
type: postmortem
---

# Postmortem — {{ .Meta.Title }}

- **Severity:** {{ .Meta.Severity }}

{{ range .Sections }}
## {{ .Title }}

{{ .Body }}
{{ end }}
`
	manifest := []byte(`version: 1
sections:
  - id: summary
    title: Summary
    type: text
    required: true
    default_body: one-paragraph summary
`)
	got, err := Convert([]byte(tmpl), manifest)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if strings.Contains(out, "header_body:") {
		t.Errorf("header_body must be omitted when sidecar supplies sections; got:\n%s", out)
	}
	if strings.Contains(out, "{{ range .Sections") {
		t.Errorf("template `{{ range .Sections }}` from sidecar-driven .tmpl leaked into v1 yaml:\n%s", out)
	}
}

func TestConvert_BilingualHeadingsGetEnglishIDs(t *testing.T) {
	t.Parallel()
	tmpl := `# T

## Контекст (Context)

body
`
	got, err := Convert([]byte(tmpl), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "- id: context") {
		t.Errorf("bilingual heading should yield English-only id:\n%s", got)
	}
}

func TestConvert_PreservesFrontmatterKeyOrder(t *testing.T) {
	t.Parallel()
	tmpl := `---
id: x
creation_date: x
type: x
title: x
---

# T

## Context
body
`
	got, err := Convert([]byte(tmpl), nil)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	// Key order in frontmatter should match source: id, creation_date, type, title.
	idIdx := strings.Index(out, "id: x")
	cdIdx := strings.Index(out, "creation_date: x")
	typeIdx := strings.Index(out, "type: x")
	titleIdx := strings.Index(out, "title: x")
	if idIdx >= cdIdx || cdIdx >= typeIdx || typeIdx >= titleIdx {
		t.Errorf("frontmatter key order not preserved:\n%s", out)
	}
}

func TestConvert_ControlFlowWrappedInDiffMarkers(t *testing.T) {
	t.Parallel()
	tmpl := `# T

## Tricky

{{ if .Foo }}
yes
{{ else }}
no
{{ end }}
`
	got, err := Convert([]byte(tmpl), nil)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if !strings.Contains(out, "<<<<<<< srekit migrate") {
		t.Errorf("control flow should be wrapped in diff markers:\n%s", out)
	}
	if !strings.Contains(out, ">>>>>>>") {
		t.Errorf("closing marker missing:\n%s", out)
	}
}

func TestConvert_NoSectionsError(t *testing.T) {
	t.Parallel()
	tmpl := `# Just a title
- and a bullet
`
	_, err := Convert([]byte(tmpl), nil)
	if err == nil {
		t.Fatal("expected error for tmpl with no sections")
	}
	if !strings.Contains(err.Error(), "no `## ` sections") {
		t.Errorf("error should explain: got %v", err)
	}
}

func TestSlugifyHeading(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Context":            "context",
		"Action items":       "action_items",
		"Контекст (Context)": "context",
		"Краткое описание (Summary)": "summary",
		"What went well":     "what_went_well",
		"Where we got lucky": "where_we_got_lucky",
		"":                   "section",
		"---":                "section",
	}
	for in, want := range cases {
		if got := slugifyHeading(in); got != want {
			t.Errorf("slugifyHeading(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHasGFMTable(t *testing.T) {
	t.Parallel()
	yes := []string{
		"| a | b |\n|---|---|\n| 1 | 2 |\n",
		"intro\n\n| a | b |\n|---|---|",
	}
	no := []string{
		"just text",
		"- a list",
		"| pipe | but | no separator |",
	}
	for _, body := range yes {
		if !hasGFMTable(body) {
			t.Errorf("hasGFMTable should detect table in:\n%s", body)
		}
	}
	for _, body := range no {
		if hasGFMTable(body) {
			t.Errorf("hasGFMTable false positive on:\n%s", body)
		}
	}
}

func TestParseGFMTable(t *testing.T) {
	t.Parallel()
	body := `| # | Action | Owner |
|---|--------|-------|
| 1 |        |       |
| 2 | thing  | alice |
`
	cols, rows := parseGFMTable(body)
	if len(cols) != 3 || cols[0] != "#" || cols[1] != "Action" || cols[2] != "Owner" {
		t.Errorf("columns wrong: %v", cols)
	}
	if len(rows) != 2 {
		t.Fatalf("rows count: %d", len(rows))
	}
	if rows[0][0] != "1" || rows[1][2] != "alice" {
		t.Errorf("row content wrong: %v", rows)
	}
}
