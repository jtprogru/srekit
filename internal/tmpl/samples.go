package tmpl

import (
	"errors"
	"fmt"
	"io"
	"text/template"
)

// Samples is the canonical sample-data registry that drove the
// Samples-based `templates validate` typo-check on shipped `.tmpl`
// artifacts. As of v0.20.0 the YAML-first migration is complete — no
// `.tmpl` ships in embed, so the registry is empty. Validate() returns
// ErrUnknownTemplate (parse-only) for any bespoke `.tmpl` users keep in
// their templates_dir; structural validation of `.yaml` artifacts is
// done by sections.ParseArtifact in `srekit templates validate`.
//
// The variable is kept (rather than removed) so external tooling can
// register fixtures for custom `.tmpl` artifacts shipped via plugins
// without forking. It may be removed in v2.0 if no plugin model
// materializes.
//
//nolint:gochecknoglobals // intentional package-level fixture registry
var Samples = map[string]any{}

// ErrUnknownTemplate is returned by Validate when name is not a builtin —
// validate falls back to parse-only and returns this sentinel so callers
// can flag the result as "syntax-only, not field-checked".
var ErrUnknownTemplate = errors.New("not a known builtin template; parse-only validation")

// Validate parses body with Funcs applied, and if name is a known builtin
// template, executes it against Samples[name] to catch references to
// fields that don't exist in the canonical struct shape.
//
// Returns ErrUnknownTemplate (wrapped) if parse succeeded but the name is
// not in Samples — callers should treat this as a soft warning, not a
// failure.
func Validate(name string, body []byte) error {
	t, err := template.New(name).Funcs(Funcs).Parse(string(body))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	sample, ok := Samples[name]
	if !ok {
		return ErrUnknownTemplate
	}
	if err := t.Execute(io.Discard, sample); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
