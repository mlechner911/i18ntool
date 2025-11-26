package app

import (
	"reflect"
	"testing"
)

func TestFlattenKeys_SimpleNested(t *testing.T) {
	tm := &TranslationManager{}

	nested := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "value",
		},
		"x": "y",
	}

	got := tm.flattenKeys("", nested)

	want := map[string]interface{}{
		"a.b": "value",
		"x":   "y",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flattenKeys result mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestCheckMissing_Basic(t *testing.T) {
	tm := &TranslationManager{
		data: map[string]map[string]interface{}{
			"en": {
				"a": map[string]interface{}{"b": "hello"},
				"c": "present",
			},
			"de": {
				"a": map[string]interface{}{"b": nil},
				// "c" missing entirely
			},
		},
		Languages: []string{"en", "de"},
	}

	missing := tm.CheckMissing()

	// Build a map for easier assertions
	missMap := make(map[string]MissingTranslation)
	for _, m := range missing {
		missMap[m.Key] = m
	}

	// Expect two missing keys: a.b (de has nil) and c (de missing)
	if _, ok := missMap["a.b"]; !ok {
		t.Fatalf("expected missing key 'a.b' not found; missing list: %#v", missing)
	}
	if _, ok := missMap["c"]; !ok {
		t.Fatalf("expected missing key 'c' not found; missing list: %#v", missing)
	}

	// Check that translations map uses "null" for missing entries
	mAB := missMap["a.b"]
	if mAB.Translations["de"] != "null" {
		t.Errorf("expected de translation for a.b to be 'null', got '%s'", mAB.Translations["de"])
	}
	if mAB.Translations["en"] == "null" {
		t.Errorf("expected en translation for a.b to be present, got 'null'")
	}

	mC := missMap["c"]
	if mC.Translations["de"] != "null" {
		t.Errorf("expected de translation for c to be 'null', got '%s'", mC.Translations["de"])
	}
	if mC.Translations["en"] == "null" {
		t.Errorf("expected en translation for c to be present, got 'null'")
	}
}
