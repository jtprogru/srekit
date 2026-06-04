package sections

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONSchema_Shape(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	sch := m.JSONSchema("test schema", []string{"title", "owner"})

	b, err := json.MarshalIndent(sch, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	// Top-level invariants.
	for _, want := range []string{
		`"$schema": "https://json-schema.org/draft/2020-12/schema"`,
		`"title": "test schema"`,
		`"type": "object"`,
		`"additionalProperties": false`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("schema missing %q in:\n%s", want, got)
		}
	}

	// Meta fields appear.
	for _, f := range []string{"title", "owner"} {
		if !strings.Contains(got, `"`+f+`": {`) {
			t.Errorf("meta property %q missing", f)
		}
	}

	// Each section ID appears as a property and required ones listed.
	for _, s := range m.Sections {
		if !strings.Contains(got, `"`+s.ID+`"`) {
			t.Errorf("section %q missing from schema", s.ID)
		}
	}
	// summary is required in sampleManifest.
	if !strings.Contains(got, `"required"`) || !strings.Contains(got, `"summary"`) {
		t.Errorf("required list should include summary:\n%s", got)
	}

	// Type info appears in description.
	if !strings.Contains(got, "type: text") || !strings.Contains(got, "type: list") || !strings.Contains(got, "type: table") {
		t.Errorf("section descriptions should mention type:\n%s", got)
	}
}

func TestJSONSchema_NoRequiredSections(t *testing.T) {
	t.Parallel()
	m := &Manifest{
		Version: 1,
		Sections: []Section{
			{ID: "a", Title: "A", Type: TypeText, Required: false},
		},
	}
	sch := m.JSONSchema("x", nil)
	sections := sch["properties"].(map[string]any)["sections"].(map[string]any)
	if _, has := sections["required"]; has {
		t.Errorf("required key should be omitted when no required sections")
	}
}
