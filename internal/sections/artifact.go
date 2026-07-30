package sections

import (
	"errors"
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

// Artifact is the v1 single-file on-disk format for a generator: header
// (frontmatter, H1 title, meta bullets, optional freeform header_body)
// + the existing typed-sections list. Replaces v0.13.x's split
// `<name>.md.tmpl` + `<name>.sections.yaml` pair with one `<name>.yaml`.
//
// `Frontmatter` is a yaml.Node (not a Go map) so author-chosen key order
// survives parse → render. The renderer walks the node to evaluate each
// scalar through the shared template engine.
//
//nolint:tagliatelle // snake_case in YAML is author-facing; see Section docs
type Artifact struct {
	Version     int       `json:"version"               yaml:"version"`
	Frontmatter yaml.Node `json:"-"                     yaml:"frontmatter,omitempty"`
	Title       string    `json:"title,omitempty"       yaml:"title,omitempty"`
	MetaBullets []string  `json:"metaBullets,omitempty" yaml:"meta_bullets,omitempty"`
	HeaderBody  string    `json:"headerBody,omitempty"  yaml:"header_body,omitempty"`
	Sections    []Section `json:"sections"              yaml:"sections"`
}

// ParseArtifact decodes data and runs structural validation: version=1,
// non-empty sections, valid section types/IDs (delegated to existing
// Manifest.Validate via a synthetic Manifest constructed from a.Sections).
func ParseArtifact(data []byte) (*Artifact, error) {
	var a Artifact
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse artifact: %w", err)
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return &a, nil
}

// Validate enforces structural invariants. Body-level checks (template
// syntax inside values, frontmatter scalar types) are deferred to render
// time so authors see them when they matter, with context.
func (a *Artifact) Validate() error {
	if a.Version != 1 {
		return fmt.Errorf("unsupported artifact version %d (expected 1)", a.Version)
	}
	if len(a.Sections) == 0 {
		return errors.New("artifact has no sections")
	}
	// Reuse the existing per-section validator by wrapping into a Manifest.
	m := &Manifest{Version: a.Version, Sections: a.Sections}
	if err := m.Validate(); err != nil {
		return err
	}
	if a.Frontmatter.Kind != 0 && a.Frontmatter.Kind != yaml.MappingNode {
		return fmt.Errorf("frontmatter must be a YAML mapping, got kind %d", a.Frontmatter.Kind)
	}
	return nil
}

// SectionIDs returns section IDs in artifact order. Mirrors Manifest.IDs
// so existing diagnostic code (unknown-ID errors etc.) can be reused.
func (a *Artifact) SectionIDs() []string {
	out := make([]string, 0, len(a.Sections))
	for _, s := range a.Sections {
		out = append(out, s.ID)
	}
	return out
}
