package tmpl

import (
	"errors"
	"fmt"
	"io"
	"text/template"
)

// Samples is the canonical sample-data registry, keyed by builtin template
// filename. It is the source of truth for `srekit templates validate` and
// must stay in sync with the struct literals each cmd/*.go file builds.
//
//nolint:gochecknoglobals // intentional package-level fixture registry
var Samples = map[string]any{
	// task.md.tmpl was removed in v0.15.0 (migrated to task.yaml — the
	// first fresh artifact migration to the v1 format). Validation goes
	// through sections.ParseArtifact in `templates validate`.
	// incident.md.tmpl and rfc.md.tmpl were removed in v0.18.0 (migrated to
	// incident.yaml / rfc.yaml — v1 artifacts).
	// postmortem.md.tmpl was removed in v0.14.0 (migrated to postmortem.yaml,
	// the v1 single-file artifact format). The artifact path doesn't go
	// through `tmpl.Validate`; structural validation of postmortem.yaml is
	// done by sections.ParseArtifact in `srekit templates validate`.
	// runbook.md.tmpl was removed in v0.19.0 (migrated to runbook.yaml).
	// slo.md.tmpl was removed in v0.16.0 (migrated to slo.yaml — v1 artifact).
	// ebp.md.tmpl and capacity.md.tmpl were removed in v0.17.0 (migrated to
	// ebp.yaml / capacity.yaml — v1 artifacts).
	// oncall.md.tmpl was removed in v0.19.0 (migrated to oncall.yaml).
	// retro.md.tmpl was removed in v0.16.0 (migrated to retro.yaml — v1 artifact).
	// changelog.md.tmpl was removed in v0.20.0 (migrated to changelog.yaml).
	// license_*.tmpl entries were removed in v0.14.0 — license bodies are
	// inlined as Go constants in cmd/license.go and no longer flow through
	// the embed FS / Samples / templates validate pipeline.
	//
	// All embedded artifacts are now v1 YAML; the Samples registry is empty.
	// Validate() falls back to parse-only with ErrUnknownTemplate for any
	// bespoke .tmpl files users keep in their templates_dir.
}

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
