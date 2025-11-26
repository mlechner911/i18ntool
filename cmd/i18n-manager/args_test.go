package main

import (
	"reflect"
	"testing"
)

func TestBuildFilesMapFromPaths_Detections(t *testing.T) {
	paths := []string{
		"examples/locales/en.json",
		"locales/de.json",
		"some/path/customfile.json",
		"noext",
		"other/en.json", // different path but same basename to force unique suffix
	}

	files := buildFilesMapFromPaths(paths)

	// Expect keys: en, de, customfile, noext, en-1
	wantKeys := []string{"en", "de", "customfile", "noext", "en-1"}

	gotKeys := make([]string, 0, len(files))
	for k := range files {
		gotKeys = append(gotKeys, k)
	}

	// We only assert presence (map order not significant)
	for _, wk := range wantKeys {
		if _, ok := files[wk]; !ok {
			t.Fatalf("expected key %q in files map, not found; got: %v", wk, files)
		}
	}

	// Check that duplicate en.json paths produced different keys
	if files["en"] == files["en-1"] {
		t.Fatalf("expected en and en-1 to map to different paths; both point to %s", files["en"])
	}

	// Ensure mapping preserves original paths
	if !reflect.DeepEqual(files["de"], "locales/de.json") {
		t.Fatalf("de should map to locales/de.json, got %s", files["de"])
	}
}
