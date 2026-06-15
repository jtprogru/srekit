package meta

import (
	"testing"

	"github.com/jtprogru/srekit/internal/config"
)

func newConfig(kv map[string]string) *config.Config {
	c := config.New()
	for k, val := range kv {
		c.Set(k, val)
	}
	return c
}

func TestResolveFlagsWin(t *testing.T) {
	t.Parallel()
	v := newConfig(map[string]string{"author": "Config Author", "email": "config@example.com"})

	a, err := Resolve(v, "Flag Author", "flag@example.com")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Flag Author" || a.Email != "flag@example.com" {
		t.Fatalf("flags must win, got %+v", a)
	}
}

func TestResolveConfigFallback(t *testing.T) {
	t.Parallel()
	v := newConfig(map[string]string{"author": "Config Author", "email": "config@example.com"})

	a, err := Resolve(v, "", "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Config Author" || a.Email != "config@example.com" {
		t.Fatalf("config fallback failed, got %+v", a)
	}
}

func TestResolveGitFallback(t *testing.T) {
	orig := gitRunner
	t.Cleanup(func() { gitRunner = orig })
	gitRunner = func(args ...string) (string, error) {
		switch args[len(args)-1] {
		case "user.name":
			return "Git Author", nil
		case "user.email":
			return "git@example.com", nil
		}
		return "", nil
	}

	a, err := Resolve(config.New(), "", "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Git Author" || a.Email != "git@example.com" {
		t.Fatalf("git fallback failed, got %+v", a)
	}
}

func TestDetectRepo(t *testing.T) {
	cases := map[string]Repo{
		"git@github.com:jtprogru/srekit.git":     {"jtprogru", "srekit"},
		"git@github.com:jtprogru/srekit":         {"jtprogru", "srekit"},
		"https://github.com/jtprogru/srekit.git": {"jtprogru", "srekit"},
		"https://github.com/jtprogru/srekit":     {"jtprogru", "srekit"},
		"https://github.com/foo/bar-baz/":        {"foo", "bar-baz"},
	}
	orig := gitRunner
	t.Cleanup(func() { gitRunner = orig })
	for url, want := range cases {
		gitRunner = func(_ ...string) (string, error) { return url, nil }
		got, err := DetectRepo()
		if err != nil {
			t.Fatalf("%s: %v", url, err)
		}
		if got != want {
			t.Errorf("%s: got %+v want %+v", url, got, want)
		}
	}

	gitRunner = func(_ ...string) (string, error) { return "", nil }
	if _, err := DetectRepo(); err == nil {
		t.Errorf("expected error on empty remote")
	}
	gitRunner = func(_ ...string) (string, error) { return "ssh://gitlab.com/x/y.git", nil }
	if _, err := DetectRepo(); err == nil {
		t.Errorf("expected error on non-github URL")
	}
}

func TestResolveMissing(t *testing.T) {
	orig := gitRunner
	t.Cleanup(func() { gitRunner = orig })
	gitRunner = func(_ ...string) (string, error) { return "", nil }

	if _, err := Resolve(config.New(), "", ""); err == nil {
		t.Fatalf("expected error when nothing configured")
	}
}
