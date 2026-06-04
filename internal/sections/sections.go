// Package sections defines the typed-section manifest format used by
// srekit generators that ship a sidecar `<template>.sections.yaml`.
//
// A manifest declares an ordered list of typed slots (text / list / table)
// with stable IDs, optional required-flag, and default content. Commands
// that consume a manifest assemble the document body from these slots —
// either using the declared defaults or overrides provided via JSON input
// (e.g. `srekit postmortem --from input.json`).
//
// The manifest format and on-the-wire RenderedSection shape are part of
// the public --json contract; both stabilize at 1.0.
package sections

import (
	"errors"
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

// SectionType enumerates the section shapes the renderer knows how to
// build a default body for. Section content fed back via --from input.json
// is always a free-form string regardless of type — the type only affects
// how defaults are rendered.
type SectionType string

const (
	TypeText  SectionType = "text"
	TypeList  SectionType = "list"
	TypeTable SectionType = "table"
)

// Manifest is the on-disk format of a `<template>.sections.yaml` file.
type Manifest struct {
	Sections []Section `json:"sections" yaml:"sections"`
	Version  int       `json:"version"  yaml:"version"`
}

// Section is one slot in a Manifest. Default-content fields are interpreted
// per Type: text uses DefaultBody only; list uses Items; table uses Columns
// + Rows (with DefaultBody rendered as a prefix above the table).
// snake_case YAML keys are deliberate: section manifests are hand-edited by
// SRE/devops authors, where snake_case is the idiomatic convention
// (Ansible/Kubernetes/Helm precedent). JSON output keys remain camelCase
// per project contract.
//
//nolint:tagliatelle // see comment above
type Section struct {
	ID          string      `json:"id"                 yaml:"id"`
	Title       string      `json:"title"              yaml:"title"`
	Type        SectionType `json:"type"               yaml:"type"`
	DefaultBody string      `json:"-"                  yaml:"default_body,omitempty"`
	Items       []string    `json:"-"                  yaml:"items,omitempty"`
	Columns     []string    `json:"-"                  yaml:"columns,omitempty"`
	Rows        [][]string  `json:"-"                  yaml:"rows,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
}

// RenderedSection is the unit shipped to the `.tmpl` `{{ range .Sections }}`
// loop and emitted in structured `--json` output. Body is always a string;
// for list/table the renderer formats Items/Columns+Rows into a markdown
// fragment so consumers see the same value through both paths.
type RenderedSection struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Body     string `json:"body"`
	Required bool   `json:"required"`
}

// ParseManifest reads a YAML manifest and returns the validated structure.
// Returns a descriptive error on schema violations so authors get an
// actionable message when editing `.sections.yaml` by hand.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks structural invariants: supported version, unique IDs,
// non-empty title, known type. It does not validate default-content shape
// beyond type recognition — `renderDefault` surfaces type-specific issues
// at render time.
func (m *Manifest) Validate() error {
	if m.Version != 1 {
		return fmt.Errorf("unsupported manifest version %d (expected 1)", m.Version)
	}
	if len(m.Sections) == 0 {
		return errors.New("manifest has no sections")
	}
	seen := make(map[string]bool, len(m.Sections))
	for i, s := range m.Sections {
		if s.ID == "" {
			return fmt.Errorf("section #%d: id is required", i)
		}
		if seen[s.ID] {
			return fmt.Errorf("section #%d: duplicate id %q", i, s.ID)
		}
		seen[s.ID] = true
		if s.Title == "" {
			return fmt.Errorf("section %q: title is required", s.ID)
		}
		switch s.Type {
		case TypeText, TypeList, TypeTable:
		case "":
			return fmt.Errorf("section %q: type is required (text|list|table)", s.ID)
		default:
			return fmt.Errorf("section %q: unknown type %q (want text|list|table)", s.ID, s.Type)
		}
	}
	return nil
}

// IDs returns section IDs in manifest order. Used for diagnostics
// (unknown-ID errors include the known list).
func (m *Manifest) IDs() []string {
	out := make([]string, 0, len(m.Sections))
	for _, s := range m.Sections {
		out = append(out, s.ID)
	}
	return out
}
