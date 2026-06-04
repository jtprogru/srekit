package sections

import "testing"

func TestRequiredCheck_AllProvided(t *testing.T) {
	t.Parallel()
	m := sampleManifest() // summary is the only required section
	input := map[string]string{"summary": "real summary"}
	results := m.RequiredCheck(input)
	if len(results) != 1 {
		t.Fatalf("want 1 result for the only required section, got %d", len(results))
	}
	if results[0].Status != "OK" || results[0].ID != "summary" {
		t.Errorf("unexpected result: %+v", results[0])
	}
}

func TestRequiredCheck_MissingFails(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	results := m.RequiredCheck(nil)
	if len(results) != 1 || results[0].Status != "FAIL" {
		t.Fatalf("missing required body should FAIL, got %+v", results)
	}
	if results[0].Reason == "" {
		t.Errorf("FAIL should carry a reason")
	}
}

func TestRequiredCheck_WhitespaceOnlyFails(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	results := m.RequiredCheck(map[string]string{"summary": "   \n  "})
	if results[0].Status != "FAIL" {
		t.Errorf("whitespace-only body should FAIL, got %+v", results[0])
	}
}

func TestRequiredCheck_SkipsOptionalSections(t *testing.T) {
	t.Parallel()
	m := &Manifest{
		Version: 1,
		Sections: []Section{
			{ID: "a", Title: "A", Type: TypeText, Required: true, DefaultBody: ""},
			{ID: "b", Title: "B", Type: TypeText, Required: false, DefaultBody: ""},
			{ID: "c", Title: "C", Type: TypeText, Required: true, DefaultBody: ""},
		},
	}
	results := m.RequiredCheck(map[string]string{"a": "x", "c": "y"})
	if len(results) != 2 {
		t.Fatalf("expected only the 2 required sections, got %d (%+v)", len(results), results)
	}
	for _, r := range results {
		if r.Status != "OK" {
			t.Errorf("unexpected fail: %+v", r)
		}
	}
}
