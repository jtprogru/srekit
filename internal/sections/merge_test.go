package sections

import (
	"strings"
	"testing"
)

func sampleManifest() *Manifest {
	return &Manifest{
		Version: 1,
		Sections: []Section{
			{ID: "summary", Title: "Summary", Type: TypeText, Required: true, DefaultBody: "_default summary_"},
			{ID: "tasks", Title: "Tasks", Type: TypeList, Items: []string{"a", "b"}},
			{ID: "log", Title: "Log", Type: TypeTable, Columns: []string{"t", "event"}},
		},
	}
}

func TestMergeAllDefaults(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	got, err := Merge(m, nil, nil)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Body != "_default summary_" {
		t.Errorf("summary body: %q", got[0].Body)
	}
	if got[1].Body != "- a\n- b" {
		t.Errorf("tasks body: %q", got[1].Body)
	}
	if !strings.HasPrefix(got[2].Body, "| t | event |") {
		t.Errorf("log body: %q", got[2].Body)
	}
}

func TestMergePartialOverride(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	input := map[string]string{"summary": "custom summary"}
	got, err := Merge(m, input, nil)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if got[0].Body != "custom summary" {
		t.Errorf("summary not overridden: %q", got[0].Body)
	}
	if got[1].Body != "- a\n- b" {
		t.Errorf("tasks should remain default: %q", got[1].Body)
	}
}

func TestMergePreservesManifestOrder(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	input := map[string]string{
		"log":     "log override",
		"summary": "summary override",
		"tasks":   "tasks override",
	}
	got, err := Merge(m, input, nil)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	wantOrder := []string{"summary", "tasks", "log"}
	for i, w := range wantOrder {
		if got[i].ID != w {
			t.Errorf("order[%d] = %q, want %q", i, got[i].ID, w)
		}
	}
}

func TestMergeOverrideBypassesTemplateExpansion(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	input := map[string]string{"summary": "look: {{ this stays literal }}"}
	got, err := Merge(m, input, nil)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if got[0].Body != "look: {{ this stays literal }}" {
		t.Errorf("override should pass through verbatim: %q", got[0].Body)
	}
}

func TestMergeUnknownIDError(t *testing.T) {
	t.Parallel()
	m := sampleManifest()
	input := map[string]string{"summery": "typo", "log": "ok"}
	_, err := Merge(m, input, nil)
	if err == nil {
		t.Fatal("want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unknown section IDs") {
		t.Errorf("want 'unknown section IDs', got: %v", err)
	}
	if !strings.Contains(msg, "summery") {
		t.Errorf("error should mention bad id: %v", err)
	}
	if !strings.Contains(msg, "summary") || !strings.Contains(msg, "tasks") || !strings.Contains(msg, "log") {
		t.Errorf("error should list known ids: %v", err)
	}
}

func TestMergeNilManifest(t *testing.T) {
	t.Parallel()
	if _, err := Merge(nil, nil, nil); err == nil {
		t.Fatal("want error for nil manifest")
	}
}
