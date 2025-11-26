package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindUnusedKeys scans project paths and returns keys that are not referenced.
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
