package sections

import (
	"strings"
	"testing"
)

func TestParseManifestValid(t *testing.T) {
	t.Parallel()
	data := []byte(`
version: 1
sections:
  - id: summary
    title: Summary
    type: text
    required: true
    default_body: "_one paragraph_"
  - id: tasks
    title: Tasks
    type: list
    items: ["do A", "do B"]
  - id: log
    title: Log
    type: table
    columns: ["t", "event"]
    rows:
      - ["t0", "start"]
      - ["t1", "end"]
`)
	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Version != 1 || len(m.Sections) != 3 {
		t.Fatalf("unexpected: %+v", m)
	}
	if m.Sections[0].Type != TypeText || m.Sections[1].Type != TypeList || m.Sections[2].Type != TypeTable {
		t.Errorf("types wrong: %+v", m.Sections)
	}
	if !m.Sections[0].Required {
		t.Errorf("summary should be required")
	}
}

func TestParseManifestRejects(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		yaml    string
		wantSub string
	}{
		{
			name:    "unknown type",
			yaml:    "version: 1\nsections:\n  - id: x\n    title: X\n    type: image\n",
			wantSub: "unknown type",
		},
		{
			name:    "missing id",
			yaml:    "version: 1\nsections:\n  - title: X\n    type: text\n",
			wantSub: "id is required",
		},
		{
			name:    "missing title",
			yaml:    "version: 1\nsections:\n  - id: x\n    type: text\n",
			wantSub: "title is required",
		},
		{
			name:    "missing type",
			yaml:    "version: 1\nsections:\n  - id: x\n    title: X\n",
			wantSub: "type is required",
		},
		{
			name:    "duplicate id",
			yaml:    "version: 1\nsections:\n  - id: x\n    title: X\n    type: text\n  - id: x\n    title: Y\n    type: text\n",
			wantSub: "duplicate id",
		},
		{
			name:    "unsupported version",
			yaml:    "version: 2\nsections:\n  - id: x\n    title: X\n    type: text\n",
			wantSub: "unsupported manifest version",
		},
		{
			name:    "no sections",
			yaml:    "version: 1\nsections: []\n",
			wantSub: "no sections",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseManifest([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("want error containing %q, got %v", tc.wantSub, err)
			}
		})
	}
}

func TestManifestIDs(t *testing.T) {
	t.Parallel()
	m := &Manifest{Sections: []Section{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}}
	got := m.IDs()
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("IDs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
