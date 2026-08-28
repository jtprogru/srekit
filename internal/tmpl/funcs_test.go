package tmpl

import (
	"errors"
	"testing"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
)

func callString(t *testing.T, name string, args ...any) string {
	t.Helper()
	fn, ok := Funcs[name].(func(...string) string)
	if ok {
		strs := make([]string, len(args))
		for i, a := range args {
			strs[i] = a.(string)
		}
		return fn(strs...)
	}
	switch f := Funcs[name].(type) {
	case func(string) string:
		return f(args[0].(string))
	case func(string, string) string:
		return f(args[0].(string), args[1].(string))
	case func(string, int) string:
		return f(args[0].(string), args[1].(int))
	default:
		t.Fatalf("unknown signature for func %q", name)
		return ""
	}
}

func TestDefaultFunc(t *testing.T) {
	if got := callString(t, "default", "fallback", ""); got != "fallback" {
		t.Fatalf("expected fallback for empty, got %q", got)
	}
	if got := callString(t, "default", "fallback", "real"); got != "real" {
		t.Fatalf("expected real value, got %q", got)
	}
}

func TestShortIDFunc(t *testing.T) {
	if got := callString(t, "shortID", "abcdef-1234-5678", 6); got != "abcdef" {
		t.Fatalf("shortID(_, 6) = %q", got)
	}
	if got := callString(t, "shortID", "abc", 8); got != "abc" {
		t.Fatalf("shortID shorter-than-n should return input, got %q", got)
	}
}

func TestSlugifyFunc(t *testing.T) {
	if got := callString(t, "slugify", "Tail Latency Spike!"); got != "tail-latency-spike" {
		t.Fatalf("slugify = %q", got)
	}
}

func TestUpperLowerTrimFuncs(t *testing.T) {
	if got := callString(t, "upper", "hi"); got != "HI" {
		t.Fatalf("upper = %q", got)
	}
	if got := callString(t, "lower", "HI"); got != "hi" {
		t.Fatalf("lower = %q", got)
	}
	if got := callString(t, "trim", "  spaced  "); got != "spaced" {
		t.Fatalf("trim = %q", got)
	}
}

func TestNowFuncRespectsClock(t *testing.T) {
	orig := clock.Now
	t.Cleanup(func() { clock.Now = orig })
	clock.Now = func() time.Time { return time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC) }

	got := callString(t, "now")
	if got != "2026-05-26T12:00:00Z" {
		t.Fatalf("default now = %q", got)
	}

	got = callString(t, "now", "2006-01-02")
	if got != "2026-05-26" {
		t.Fatalf("custom format now = %q", got)
	}
}

func TestValidateAppliesFuncMap(t *testing.T) {
	// Every template parse must wire the FuncMap so bodies can call
	// shortID / default / slugify / now. Validate is the production entry
	// point for `.tmpl` files (`srekit templates validate`); a body using
	// the helpers must parse, and an unknown helper must not.
	body := []byte(`{{ "abc-1234" | shortID 4 }} / {{ "" | default "fallback" }}`)
	if err := Validate("fixture.tmpl", body); err != nil && !errors.Is(err, ErrUnknownTemplate) {
		t.Fatalf("fixture should parse with FuncMap: %v", err)
	}

	bogus := []byte(`{{ "x" | nosuchfunc }}`)
	if err := Validate("bogus.tmpl", bogus); err == nil || errors.Is(err, ErrUnknownTemplate) {
		t.Fatalf("an unknown helper must fail to parse, got: %v", err)
	}
}

// TestJoinFunc covers the helper that lets a template turn a []string
// meta field into a YAML flow sequence — the tasker artifact's `level`.
// Separator first, so the pipe form reads naturally.
func TestJoinFunc(t *testing.T) {
	fn, ok := Funcs["join"].(func(string, []string) string)
	if !ok {
		t.Fatalf("join has an unexpected signature: %T", Funcs["join"])
	}
	if got := fn(", ", []string{"middle", "senior"}); got != "middle, senior" {
		t.Fatalf("join = %q", got)
	}
	if got := fn(", ", nil); got != "" {
		t.Fatalf("join of nothing = %q, want empty", got)
	}
	if got := fn(", ", []string{"junior"}); got != "junior" {
		t.Fatalf("single-item join = %q", got)
	}
}
