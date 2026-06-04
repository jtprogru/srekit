package sections

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/jtprogru/srekit/internal/tmpl"
)

// renderDefault produces the default body for a section. Template
// directives inside default_body / items / rows are evaluated against ctx
// using the same FuncMap that the main template uses, so writers can
// reference values like {{ .Meta.Start }} or {{ now "2006-01-02" }} in
// section defaults.
func renderDefault(s Section, ctx any) (string, error) {
	switch s.Type {
	case TypeText:
		return renderTemplate(s.DefaultBody, ctx)
	case TypeList:
		return renderList(s.Items, ctx)
	case TypeTable:
		prefix, err := renderTemplate(s.DefaultBody, ctx)
		if err != nil {
			return "", err
		}
		table, err := renderTable(s.Columns, s.Rows, ctx)
		if err != nil {
			return "", err
		}
		prefix = strings.TrimRight(prefix, "\n")
		switch {
		case prefix == "":
			return table, nil
		case table == "":
			return prefix, nil
		default:
			return prefix + "\n\n" + table, nil
		}
	default:
		return "", fmt.Errorf("unknown section type %q", s.Type)
	}
}

// renderTemplate runs body through the shared FuncMap. A body with no
// directives passes through verbatim; this keeps trivial literal strings
// (e.g. an empty placeholder) cheap.
func renderTemplate(body string, ctx any) (string, error) {
	if body == "" || !strings.Contains(body, "{{") {
		return body, nil
	}
	t, err := template.New("section").Funcs(tmpl.Funcs).Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse default body: %w", err)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("execute default body: %w", err)
	}
	return buf.String(), nil
}

// renderList emits one markdown bullet per item, each rendered through
// the FuncMap. Empty items become bare bullets — useful as placeholders
// agents can fill in.
func renderList(items []string, ctx any) (string, error) {
	if len(items) == 0 {
		return "-", nil
	}
	var b strings.Builder
	for i, it := range items {
		rendered, err := renderTemplate(it, ctx)
		if err != nil {
			return "", err
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- ")
		b.WriteString(rendered)
	}
	return b.String(), nil
}

// renderTable emits a GFM-compatible markdown table. When no rows are
// provided but columns are declared, one empty placeholder row is
// rendered so the table is visually complete.
func renderTable(columns []string, rows [][]string, ctx any) (string, error) {
	if len(columns) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("|")
	for _, c := range columns {
		b.WriteString(" ")
		b.WriteString(c)
		b.WriteString(" |")
	}
	b.WriteString("\n|")
	for range columns {
		b.WriteString("------|")
	}
	b.WriteString("\n")
	if len(rows) == 0 {
		b.WriteString("|")
		for range columns {
			b.WriteString("      |")
		}
		return b.String(), nil
	}
	for ri, row := range rows {
		if ri > 0 {
			b.WriteString("\n")
		}
		b.WriteString("|")
		for ci := range columns {
			cell := ""
			if ci < len(row) {
				cell = row[ci]
			}
			rendered, err := renderTemplate(cell, ctx)
			if err != nil {
				return "", err
			}
			b.WriteString(" ")
			b.WriteString(rendered)
			b.WriteString(" |")
		}
	}
	return b.String(), nil
}
