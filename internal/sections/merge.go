package sections

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// CheckUnknownIDs returns an error if input contains section IDs not
// present in the manifest. It is the rendering-free variant of Merge's
// typo guard — useful for `--validate` and other places that want to
// reject typos without triggering default-body template evaluation
// (which may fail when the validate-time ctx is empty).
func CheckUnknownIDs(manifest *Manifest, input map[string]string) error {
	if manifest == nil {
		return errors.New("manifest is nil")
	}
	known := make(map[string]bool, len(manifest.Sections))
	for _, s := range manifest.Sections {
		known[s.ID] = true
	}
	var unknown []string
	for id := range input {
		if !known[id] {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf(
		"unknown section IDs: [%s]; known: [%s]",
		strings.Join(unknown, ", "),
		strings.Join(manifest.IDs(), ", "),
	)
}

// Merge produces a list of RenderedSection in manifest order. For each
// section, the body is the value from input[section.ID] if present, or
// the rendered default otherwise. Input bodies are used verbatim — no
// template evaluation — so consumers can round-trip arbitrary markdown
// without surprise expansion of {{ ... }} sequences.
//
// If input contains IDs that are not in the manifest, Merge returns an
// error listing the offending IDs plus the known set. This guards against
// agent typos (e.g. a misspelled section ID) silently disappearing into
// the void.
func Merge(manifest *Manifest, input map[string]string, ctx any) ([]RenderedSection, error) {
	if err := CheckUnknownIDs(manifest, input); err != nil {
		return nil, err
	}

	out := make([]RenderedSection, 0, len(manifest.Sections))
	for _, s := range manifest.Sections {
		body, ok := input[s.ID]
		if !ok {
			rendered, err := renderDefault(s, ctx)
			if err != nil {
				return nil, fmt.Errorf("section %q: %w", s.ID, err)
			}
			body = rendered
		}
		title := s.Title
		if strings.Contains(title, "{{") {
			rendered, err := renderTemplate(title, ctx)
			if err != nil {
				return nil, fmt.Errorf("section %q title: %w", s.ID, err)
			}
			title = rendered
		}
		out = append(out, RenderedSection{
			ID:       s.ID,
			Title:    title,
			Type:     string(s.Type),
			Required: s.Required,
			Body:     body,
		})
	}
	return out, nil
}
