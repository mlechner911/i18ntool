package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mlechner911/i18ntool/internal/app"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: i18n-manager <command> [options]")
        os.Exit(1)
    }

    command := os.Args[1]

    switch command {
    case "check":
        if len(os.Args) < 3 {
            fmt.Println("Usage: i18n-manager check <file1.json> <file2.json> [...]")
            os.Exit(1)
        }

        // Accept N json files. Try to derive language code from filename (e.g. "de.json")
        // or from a parent directory name (e.g. "locales/de/en.json"). If detection
        // fails, fall back to a safe identifier (basename or fileN). Language detection
        // is only for reporting, so failures are NOT fatal.
        files := make(map[string]string)
        used := make(map[string]bool)
        langRe := regexp.MustCompile(`^[A-Za-z]{2}([_-][A-Za-z]{2})?$`)

        for idx, p := range os.Args[2:] {
            base := filepath.Base(p)
            name := strings.TrimSuffix(base, filepath.Ext(base))

            var lang string
            if langRe.MatchString(name) {
                lang = name
            } else {
                // check parent directory
                parent := filepath.Base(filepath.Dir(p))
                if langRe.MatchString(parent) {
                    lang = parent
                }
            }

            if lang == "" {
                // fallback to basename if useful, otherwise file-<n>
                if name != "" {
                    lang = name
                } else {
                    lang = fmt.Sprintf("file-%d", idx+1)
                }
            }

            // ensure uniqueness using hyphen suffixes
            orig := lang
            i := 1
            for used[lang] {
                lang = fmt.Sprintf("%s-%d", orig, i)
                i++
            }
            used[lang] = true
            files[lang] = p
        }

        tm, err := app.NewTranslationManager(files)
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
                for i, lang := range tm.Languages {
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

        tm, err := app.NewTranslationManager(files)
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

        tm, err := app.NewTranslationManager(files)
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

        tm := &app.TranslationManager{}

        if err := tm.AddTranslation(filePath, key, value); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
        os.Exit(1)
    }
}
