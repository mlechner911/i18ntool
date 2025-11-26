package simpletrans

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTranslations_EmbeddedFallback(t *testing.T) {
	// Attempt to load a non-existent file but with a name matching an embedded locale
	cwd, _ := os.Getwd()
	// create a path that does not exist but whose basename is en.json so
	// LoadTranslations will try to use the embedded "en" translations.
	fake := filepath.Join(cwd, "no_such_dir", "en.json")
	// ensure it does not exist
	_ = os.Remove(fake)
	tr, err := LoadTranslations(fake)
	if err == nil {
		// Should succeed via embedded fallback
		if tr == nil {
			t.Fatalf("expected non-nil translations from embedded fallback")
		}
		// check a known key exists
		if v, ok := tr["usage.general"]; !ok || v == nil {
			t.Fatalf("expected usage.general in embedded translations, got %v", v)
		}
	} else {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
}

func TestGetTranslation_Unescape(t *testing.T) {
	// Use embedded map directly
	m := EmbeddedTranslations["en"]
	tr := buildTranslationsFromFlat(m)
	// create a string with escaped newline in embedded map if not present
	key := "test.unescape"
	m[key] = "line1\\nline2"
	// rebuild translations
	tr = buildTranslationsFromFlat(m)
	out, err := GetTranslation(tr, key, nil, "")
	if err != nil {
		t.Fatalf("GetTranslation error: %v", err)
	}
	if out != "line1\nline2" {
		t.Fatalf("expected unescaped newline, got %q", out)
	}
}
