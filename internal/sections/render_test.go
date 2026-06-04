package sections

import (
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
)

type renderCtx struct {
	Meta struct {
		Start string
		End   string
	}
}

func TestRenderDefaultText(t *testing.T) {
	t.Parallel()
	s := Section{ID: "summary", Type: TypeText, DefaultBody: "_one paragraph_"}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "_one paragraph_" {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultTextTemplateExpansion(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: TypeText, DefaultBody: "started at {{ .Meta.Start }}"}
	var ctx renderCtx
	ctx.Meta.Start = "2026-06-04T10:00:00Z"
	got, err := renderDefault(s, ctx)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "started at 2026-06-04T10:00:00Z" {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultTextFuncMap(t *testing.T) {
	t.Parallel()
	orig := clock.Now
	t.Cleanup(func() { clock.Now = orig })
	clock.Now = func() time.Time { return time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC) }

	s := Section{ID: "x", Type: TypeText, DefaultBody: `created {{ now "2006-01-02" }}`}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "created 2026-06-04" {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultListItems(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: TypeList, Items: []string{"alpha", "beta"}}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "- alpha\n- beta" {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultListEmpty(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: TypeList}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "-" {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultListSingleEmptyItem(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: TypeList, Items: []string{""}}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "- " {
		t.Errorf("got %q", got)
	}
}

func TestRenderDefaultTableHeaderOnly(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: TypeTable, Columns: []string{"col1", "col2"}}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "| col1 | col2 |\n|------|------|\n|      |      |"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderDefaultTableWithRows(t *testing.T) {
	t.Parallel()
	s := Section{
		ID:      "x",
		Type:    TypeTable,
		Columns: []string{"t", "event"},
		Rows: [][]string{
			{"t0", "start"},
			{"t1", "end"},
		},
	}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "| t | event |\n|------|------|\n| t0 | start |\n| t1 | end |"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderDefaultTablePrefix(t *testing.T) {
	t.Parallel()
	s := Section{
		ID:          "x",
		Type:        TypeTable,
		DefaultBody: "(All times UTC.)",
		Columns:     []string{"t", "event"},
	}
	got, err := renderDefault(s, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.HasPrefix(got, "(All times UTC.)\n\n| t | event |") {
		t.Errorf("prefix not applied:\n%s", got)
	}
}

func TestRenderDefaultTableTemplatedRows(t *testing.T) {
	t.Parallel()
	var ctx renderCtx
	ctx.Meta.Start = "2026-06-04T10:00:00Z"
	ctx.Meta.End = "2026-06-04T11:30:00Z"
	s := Section{
		ID:      "timeline",
		Type:    TypeTable,
		Columns: []string{"Time", "Event"},
		Rows: [][]string{
			{"{{ .Meta.Start }}", "started"},
			{"{{ .Meta.End }}", "resolved"},
		},
	}
	got, err := renderDefault(s, ctx)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(got, "2026-06-04T10:00:00Z") || !strings.Contains(got, "2026-06-04T11:30:00Z") {
		t.Errorf("template rows not expanded:\n%s", got)
	}
}

func TestRenderDefaultUnknownType(t *testing.T) {
	t.Parallel()
	s := Section{ID: "x", Type: "image"}
	_, err := renderDefault(s, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
