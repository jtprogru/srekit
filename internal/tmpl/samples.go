package tmpl

import (
	"errors"
	"fmt"
	"io"
	"text/template"
)

// authorSample mirrors meta.Author without importing internal/meta — keeps
// the tmpl package free of cross-cuts from cmd/meta.
type authorSample struct {
	Name  string
	Email string
}

// Samples is the canonical sample-data registry, keyed by builtin template
// filename. It is the source of truth for `srekit templates validate` and
// must stay in sync with the struct literals each cmd/*.go file builds.
//
//nolint:gochecknoglobals // intentional package-level fixture registry
var Samples = map[string]any{
	"task.md.tmpl": struct {
		ID, Title, CreationDate, ModificationDate string
	}{
		ID:               "11111111-2222-3333-4444-555555555555",
		Title:            "Sample title",
		CreationDate:     "2026-01-01T00:00:00Z",
		ModificationDate: "2026-01-01T00:00:00Z",
	},
	"incident.md.tmpl": struct {
		ID, Title, Severity, Lead, Status, Now string
	}{
		ID:       "11111111-2222-3333-4444-555555555555",
		Title:    "Sample incident",
		Severity: "SEV-2",
		Lead:     "alice",
		Status:   "active",
		Now:      "2026-01-01T00:00:00Z",
	},
	"postmortem.md.tmpl": struct {
		ID, Title, Severity, Start, End, Owner, Now string
	}{
		ID:       "11111111-2222-3333-4444-555555555555",
		Title:    "Sample postmortem",
		Severity: "SEV-3",
		Start:    "2026-01-01T00:00:00Z",
		End:      "2026-01-01T01:00:00Z",
		Owner:    "@oncall",
		Now:      "2026-01-01T02:00:00Z",
	},
	"runbook.md.tmpl": struct {
		ID, Title, Service, Alert, Now string
	}{
		ID:      "11111111-2222-3333-4444-555555555555",
		Title:   "Sample runbook",
		Service: "api-gw",
		Alert:   "APIHighLatency",
		Now:     "2026-01-01T00:00:00Z",
	},
	"rfc.md.tmpl": struct {
		ID, Title, Status, Now string
		Author                 authorSample
	}{
		ID:     "11111111-2222-3333-4444-555555555555",
		Title:  "Sample RFC",
		Status: "proposed",
		Now:    "2026-01-01T00:00:00Z",
		Author: authorSample{Name: "Sample Author", Email: "sample@example.com"},
	},
	"slo.md.tmpl": struct {
		ID, Service, Target, Window, LatencyTarget, Now string
	}{
		ID:            "11111111-2222-3333-4444-555555555555",
		Service:       "api-gw",
		Target:        "99.9%",
		Window:        "30d",
		LatencyTarget: "300ms",
		Now:           "2026-01-01T00:00:00Z",
	},
	"ebp.md.tmpl": struct {
		ID, Service, Now string
	}{
		ID:      "11111111-2222-3333-4444-555555555555",
		Service: "api-gw",
		Now:     "2026-01-01T00:00:00Z",
	},
	"capacity.md.tmpl": struct {
		ID, Service, Horizon, Now string
	}{
		ID:      "11111111-2222-3333-4444-555555555555",
		Service: "api-gw",
		Horizon: "1y",
		Now:     "2026-01-01T00:00:00Z",
	},
	"oncall.md.tmpl": struct {
		ID, Team, Start, End, Now string
		Author                    authorSample
	}{
		ID:     "11111111-2222-3333-4444-555555555555",
		Team:   "platform",
		Start:  "2026-01-05",
		End:    "2026-01-11",
		Now:    "2026-01-12T00:00:00Z",
		Author: authorSample{Name: "Sample Oncaller", Email: "oncall@example.com"},
	},
	"retro.md.tmpl": struct {
		ID, Team, Sprint, Now string
	}{
		ID:     "11111111-2222-3333-4444-555555555555",
		Team:   "platform",
		Sprint: "2026-W01",
		Now:    "2026-01-01T00:00:00Z",
	},
	"changelog.md.tmpl": struct {
		Today, Repo, InitialVersion string
	}{
		Today:          "2026-01-01",
		Repo:           "owner/repo",
		InitialVersion: "0.1.0",
	},
	"license_wtfpl.tmpl": struct {
		Year   int
		Author authorSample
	}{
		Year:   2026,
		Author: authorSample{Name: "Sample Author", Email: "sample@example.com"},
	},
	"license_mit.tmpl": struct {
		Year   int
		Author authorSample
	}{
		Year:   2026,
		Author: authorSample{Name: "Sample Author", Email: "sample@example.com"},
	},
	"license_apache2.tmpl": struct {
		Year   int
		Author authorSample
	}{
		Year:   2026,
		Author: authorSample{Name: "Sample Author", Email: "sample@example.com"},
	},
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
