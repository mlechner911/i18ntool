package app

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
    Languages []string
}

type MissingTranslation struct {
    Key          string
    Translations map[string]string
}

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
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("reading %s: %w", filePath, err)
    }

    backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
    if err := os.WriteFile(backupPath, content, 0644); err != nil {
        return fmt.Errorf("creating backup %s: %w", backupPath, err)
    }
    fmt.Printf("Backup created: %s\n", backupPath)

    var data map[string]interface{}
    if err := json.Unmarshal(content, &data); err != nil {
        return fmt.Errorf("parsing %s: %w", filePath, err)
    }

    if tm.keyExists(key, data) {
        return fmt.Errorf("key '%s' already exists in %s", key, filePath)
    }

    if err := tm.addNestedKey(data, key, value); err != nil {
        return fmt.Errorf("adding key '%s': %w", key, err)
    }

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
                return true
            }
            if nested, ok := val.(map[string]interface{}); ok {
                current = nested
            } else {
                return false
            }
        } else {
            return false
        }
    }
    return false
}

func (tm *TranslationManager) addNestedKey(data map[string]interface{}, key, value string) error {
    parts := strings.Split(key, ".")
    current := data

    for i, part := range parts {
        if i == len(parts)-1 {
            current[part] = value
            return nil
        }

        if val, exists := current[part]; exists {
            if nested, ok := val.(map[string]interface{}); ok {
                current = nested
            } else {
                return fmt.Errorf("cannot add nested key '%s': '%s' is not an object", key, strings.Join(parts[:i+1], "."))
            }
        } else {
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
