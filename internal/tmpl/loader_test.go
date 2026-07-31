package tmpl

import (
	"testing"
)

// TestLoader_DirOverridesEmbed was removed in v0.30.0 with Loader.Parse.
// Source priority is covered by TestLoadArtifactBytes_DirOverridesEmbed,
// which walks the same chain.

func TestLoader_FallbackToEmbed(t *testing.T) {
	// Dir exists but doesn't contain the requested artifact — should fall
	// through to embed. After v0.20.0 every embedded artifact is a v1
	// YAML, so we exercise the fallback via LoadArtifactBytes (the .tmpl
	// path is no longer represented in embed).
	dir := t.TempDir()
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	body, err := loader.LoadArtifactBytes("postmortem")
	if err != nil {
		t.Fatalf("expected embed fallback, got error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body from embed fallback")
	}
}

// TestLoader_NotFoundInAnySource was removed in v0.30.0 with
// Loader.Parse. The exhausted-sources case is covered by
// TestLoadArtifactBytes_NotFound.

// TestParseFile_AppliesFuncs was removed in v0.30.0 with ParseFile itself.
// Funcs application is covered by TestValidateAppliesFuncMap and the rest
// of the funcs_test suite.

func TestAddDirSource_Prepends(t *testing.T) {
	// Default order is [Embed]. After AddDirSource, it becomes [Dir, Embed].
	loader := &Loader{Sources: []Source{EmbedSource{}}}
	loader.AddDirSource("/tmp/whatever")
	if _, ok := loader.Sources[0].(DirSource); !ok {
		t.Fatalf("expected DirSource first, got %T", loader.Sources[0])
	}
	if _, ok := loader.Sources[1].(EmbedSource); !ok {
		t.Fatalf("expected EmbedSource second, got %T", loader.Sources[1])
	}
}
