package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

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
