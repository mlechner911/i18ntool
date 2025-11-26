package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TranslationManager struct {
	files     map[string]string
	data      map[string]map[string]interface{}
	languages []string
}

type MissingTranslation struct {
	Key          string
	Translations map[string]string
}

func NewTranslationManager(files map[string]string) (*TranslationManager, error) {
	tm := &TranslationManager{
		files:     files,
		data:      make(map[string]map[string]interface{}),
		languages: make([]string, 0, len(files)),
	}

	for lang, path := range files {
		tm.languages = append(tm.languages, lang)
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

	sort.Strings(tm.languages)
	return tm, nil
}

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

func (tm *TranslationManager) CheckMissing() []MissingTranslation {
	allKeys := tm.GetAllKeys()
	missing := make([]MissingTranslation, 0)

	for _, key := range allKeys {
		translations := make(map[string]string)
		hasMissing := false

		for _, lang := range tm.languages {
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

func (tm *TranslationManager) SortAndSave(createBackup bool) error {
	for lang, path := range tm.files {
		if createBackup {
			backupPath := path + ".backup." + time.Now().Format("20060102-150405")
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s for backup: %w", path, err)
			}
			if err := os.WriteFile(backupPath, content, 0644); err != nil {
				return fmt.Errorf("creating backup %s: %w", backupPath, err)
			}
			fmt.Printf("Backup created: %s\n", backupPath)
		}

		sorted := tm.sortMap(tm.data[lang])
		content, err := json.MarshalIndent(sorted, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling %s: %w", lang, err)
		}

		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		fmt.Printf("Sorted and saved: %s\n", path)
	}

	return nil
}

func (tm *TranslationManager) sortMap(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		sorted := make(map[string]interface{})
		for _, key := range keys {
			sorted[key] = tm.sortMap(v[key])
		}
		return sorted
	default:
		return v
	}
}

func (tm *TranslationManager) AddTranslation(filePath, key, value string) error {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	// Create backup
	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("creating backup %s: %w", backupPath, err)
	}
	fmt.Printf("Backup created: %s\n", backupPath)

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Check if key already exists
	if tm.keyExists(key, data) {
		return fmt.Errorf("key '%s' already exists in %s", key, filePath)
	}

	// Add the new key
	if err := tm.addNestedKey(data, key, value); err != nil {
		return fmt.Errorf("adding key '%s': %w", key, err)
	}

	// Sort and save
	sorted := tm.sortMap(data)
	newContent, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(filePath, newContent, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", filePath, err)
	}

	fmt.Printf("Added translation '%s' = '%s' to %s\n", key, value, filePath)
	return nil
}

func (tm *TranslationManager) keyExists(key string, data map[string]interface{}) bool {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if val, exists := current[part]; exists {
			if i == len(parts)-1 {
				// Last part exists
				return true
			}
			// Continue deeper
			if nested, ok := val.(map[string]interface{}); ok {
				current = nested
			} else {
				// Path blocked by non-object value
				return false
			}
		} else {
			// Key doesn't exist at this level
			return false
		}
	}
	return false
}

func (tm *TranslationManager) addNestedKey(data map[string]interface{}, key, value string) error {
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent of the final key
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			current[part] = value
			return nil
		}

		// Check if this level exists
		if val, exists := current[part]; exists {
			// Exists - make sure it's an object
			if nested, ok := val.(map[string]interface{}); ok {
				current = nested
			} else {
				return fmt.Errorf("cannot add nested key '%s': '%s' is not an object", key, strings.Join(parts[:i+1], "."))
			}
		} else {
			// Doesn't exist - create new object
			newObj := make(map[string]interface{})
			current[part] = newObj
			current = newObj
		}
	}

	return nil
}

func (tm *TranslationManager) FindUnusedKeys(projectPaths []string) ([]string, error) {
	allKeys := tm.GetAllKeys()
	usedKeys := make(map[string]bool)

	for _, projectPath := range projectPaths {
		err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
				return filepath.SkipDir
			}

			if !d.IsDir() && (strings.HasSuffix(path, ".vue") ||
				strings.HasSuffix(path, ".ts") ||
				strings.HasSuffix(path, ".js")) {

				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				contentStr := string(content)
				// Check each key directly in the file content
				for _, key := range allKeys {
					if strings.Contains(contentStr, key) {
						usedKeys[key] = true
					}
				}
			}
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("scanning %s: %w", projectPath, err)
		}
	}

	unused := make([]string, 0)
	for _, key := range allKeys {
		if !usedKeys[key] {
			unused = append(unused, key)
		}
	}

	return unused, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: i18n-manager <command> [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  check <en.json> <de.json> <es.json>  - Check for missing translations")
		fmt.Println("  sort <en.json> <de.json> <es.json>   - Sort and save JSON files (with backup)")
		fmt.Println("  unused <en.json> <de.json> <es.json> <project-path>... - Find unused keys")
		fmt.Println("  add <file.json> <key> <value>        - Add a new translation key (with backup)")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		if len(os.Args) < 5 {
			fmt.Println("Usage: i18n-manager check <en.json> <de.json> <es.json>")
			os.Exit(1)
		}

		files := map[string]string{
			"en": os.Args[2],
			"de": os.Args[3],
			"es": os.Args[4],
		}

		tm, err := NewTranslationManager(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		missing := tm.CheckMissing()
		if len(missing) == 0 {
			fmt.Println("All translations complete!")
		} else {
			fmt.Printf("Found %d missing translations:\n\n", len(missing))
			for _, m := range missing {
				fmt.Printf("key: %s: { ", m.Key)
				for i, lang := range tm.languages {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s: %s", lang, m.Translations[lang])
				}
				fmt.Println(" }")
			}
		}

	case "sort":
		if len(os.Args) < 5 {
			fmt.Println("Usage: i18n-manager sort <en.json> <de.json> <es.json>")
			os.Exit(1)
		}

		files := map[string]string{
			"en": os.Args[2],
			"de": os.Args[3],
			"es": os.Args[4],
		}

		tm, err := NewTranslationManager(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := tm.SortAndSave(true); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "unused":
		if len(os.Args) < 6 {
			fmt.Println("Usage: i18n-manager unused <en.json> <de.json> <es.json> <project-path>...")
			os.Exit(1)
		}

		files := map[string]string{
			"en": os.Args[2],
			"de": os.Args[3],
			"es": os.Args[4],
		}

		projectPaths := os.Args[5:]

		tm, err := NewTranslationManager(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		unused, err := tm.FindUnusedKeys(projectPaths)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(unused) == 0 {
			fmt.Println("All keys are used!")
		} else {
			fmt.Printf("Found %d unused keys:\n", len(unused))
			for _, key := range unused {
				fmt.Printf("  - %s\n", key)
			}
		}

	case "add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: i18n-manager add <file.json> <key> <value>")
			os.Exit(1)
		}

		filePath := os.Args[2]
		key := os.Args[3]
		value := os.Args[4]

		// Create a temporary TranslationManager just to use the AddTranslation method
		tm := &TranslationManager{}

		if err := tm.AddTranslation(filePath, key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
