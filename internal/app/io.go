package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// NewTranslationManager loads the provided files and returns a TranslationManager.
func NewTranslationManager(files map[string]string) (*TranslationManager, error) {
	tm := &TranslationManager{
		files:     files,
		data:      make(map[string]map[string]interface{}),
		Languages: make([]string, 0, len(files)),
	}

	for lang, path := range files {
		tm.Languages = append(tm.Languages, lang)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		tm.data[lang] = data
	}

	sort.Strings(tm.Languages)
	return tm, nil
}
