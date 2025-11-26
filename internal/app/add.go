package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

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
