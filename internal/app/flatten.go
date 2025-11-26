package app

import (
	"sort"
)

// flattenKeys flattens nested JSON-like structures into dot-separated keys.
func (tm *TranslationManager) flattenKeys(prefix string, data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			for k, val := range tm.flattenKeys(fullKey, value) {
				result[k] = val
			}
		}
	default:
		result[prefix] = v
	}

	return result
}

// GetAllKeys returns the set of all flattened keys across loaded languages.
func (tm *TranslationManager) GetAllKeys() []string {
	keySet := make(map[string]bool)

	for _, data := range tm.data {
		flat := tm.flattenKeys("", data)
		for key := range flat {
			keySet[key] = true
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
