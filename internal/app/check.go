package app

import "fmt"

// CheckMissing returns a list of keys that have missing translations.
func (tm *TranslationManager) CheckMissing() []MissingTranslation {
	allKeys := tm.GetAllKeys()
	missing := make([]MissingTranslation, 0)

	for _, key := range allKeys {
		translations := make(map[string]string)
		hasMissing := false

		for _, lang := range tm.Languages {
			flat := tm.flattenKeys("", tm.data[lang])
			if val, exists := flat[key]; exists && val != nil {
				translations[lang] = fmt.Sprintf("%v", val)
			} else {
				translations[lang] = "null"
				hasMissing = true
			}
		}

		if hasMissing {
			missing = append(missing, MissingTranslation{
				Key:          key,
				Translations: translations,
			})
		}
	}

	return missing
}
