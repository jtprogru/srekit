package meta

import (
	"testing"

	"github.com/spf13/viper"
)

func TestResolveFlagsWin(t *testing.T) {
	viper.Reset()
	viper.Set("author", "Viper Author")
	viper.Set("email", "viper@example.com")
	t.Cleanup(viper.Reset)

	a, err := Resolve("Flag Author", "flag@example.com")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Flag Author" || a.Email != "flag@example.com" {
		t.Fatalf("flags must win, got %+v", a)
	}
}

func TestResolveViperFallback(t *testing.T) {
	viper.Reset()
	viper.Set("author", "Viper Author")
	viper.Set("email", "viper@example.com")
	t.Cleanup(viper.Reset)

	a, err := Resolve("", "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Viper Author" || a.Email != "viper@example.com" {
		t.Fatalf("viper fallback failed, got %+v", a)
	}
}

func TestResolveGitFallback(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

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

	a, err := Resolve("", "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if a.Name != "Git Author" || a.Email != "git@example.com" {
		t.Fatalf("git fallback failed, got %+v", a)
	}
}

func TestDetectRepo(t *testing.T) {
	cases := map[string]Repo{
		"git@github.com:jtprogru/srekit.git":         {"jtprogru", "srekit"},
		"git@github.com:jtprogru/srekit":             {"jtprogru", "srekit"},
		"https://github.com/jtprogru/srekit.git":     {"jtprogru", "srekit"},
		"https://github.com/jtprogru/srekit":         {"jtprogru", "srekit"},
		"https://github.com/foo/bar-baz/":            {"foo", "bar-baz"},
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
	viper.Reset()
	t.Cleanup(viper.Reset)

	orig := gitRunner
	t.Cleanup(func() { gitRunner = orig })
	gitRunner = func(_ ...string) (string, error) { return "", nil }

	if _, err := Resolve("", ""); err == nil {
		t.Fatalf("expected error when nothing configured")
	}
}
