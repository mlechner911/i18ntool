package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mlechner911/i18ntool/internal/app"
	"github.com/mlechner911/i18ntool/internal/simpletrans"
)

// main is the CLI entrypoint for i18n-manager.
func main() {
	// parse optional global flags first (simple scan for --lang or -l)
	lang := "en" // default language when not specified
	args := make([]string, 0, len(os.Args))
	args = append(args, os.Args[0])
	skipNext := false
	for i := 1; i < len(os.Args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		a := os.Args[i]
		if a == "--lang" || a == "-l" {
			if i+1 < len(os.Args) {
				lang = os.Args[i+1]
				skipNext = true
				continue
			}
		}
		args = append(args, a)
	}

	// load translations for selected language (best-effort search locations)
	var translations simpletrans.Translations
	var loadErr error
	// try common locations relative to repo
	candidates := []string{
		filepath.Join("examples", "locales", fmt.Sprintf("%s.json", lang)),
		filepath.Join("locales", fmt.Sprintf("%s.json", lang)),
		filepath.Join("testfiles", fmt.Sprintf("%s.json", lang)),
	}
	for _, c := range candidates {
		translations, loadErr = simpletrans.LoadTranslations(c)
		if loadErr == nil {
			break
		}
	}

	// helper to translate a message key or fallback to given literal
	translate := func(msg string) string {
		if translations != nil {
			// try dotted key lookup first
			if out, _ := simpletrans.GetTranslation(translations, msg, nil, ""); out != "" {
				return out
			}
			// try top-level literal key
			if v, ok := translations[msg]; ok {
				if s, ok := v.(string); ok {
					return simpletrans.Unescape(s)
				}
			}
		}
		return msg
	}

	// convenience print helpers
	tprintf := func(format string, a ...interface{}) {
		// translate the format string then apply formatting
		tf := translate(format)
		fmt.Printf(tf, a...)
	}
	tprintln := func(s string) {
		fmt.Println(translate(s))
	}

	if len(args) < 2 {
		tprintln(translate("usage.general"))
		os.Exit(1)
	}

	command := args[1]

	switch command {
	case "help":
		// print basic usage and available commands
		tprintln(translate("usage.general"))
		tprintln(translate("usage.check"))
		tprintln(translate("usage.sort"))
		tprintln(translate("usage.unused"))
		tprintln(translate("usage.add"))
		tprintln(translate("usage.simple"))
		os.Exit(0)

	case "check":
		if len(args) < 3 {
			tprintln(translate("usage.check"))
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
			fmt.Fprintf(os.Stderr, translate("Error: %v\n"), err)
			os.Exit(1)
		}

		missing := tm.CheckMissing()
		if len(missing) == 0 {
			tprintln("All translations complete!")
		} else {
			tprintf("Found %d missing translations:\n\n", len(missing))
			for _, m := range missing {
				tprintf("key: %s: { ", m.Key)
				for i, lang := range tm.Languages {
					if i > 0 {
						fmt.Print(", ")
					}
					tprintf("%s: %s", lang, m.Translations[lang])
				}
				tprintln(" }")
			}
		}

	case "sort":
		if len(args) < 3 {
			tprintln(translate("usage.sort"))
			os.Exit(1)
		}

		files := buildFilesMapFromPaths(os.Args[2:])

		tm, err := app.NewTranslationManager(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, translate("Error: %v\n"), err)
			os.Exit(1)
		}

		if err := tm.SortAndSave(true); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "unused":
		// Usage: i18n-manager unused <file1.json> <file2.json> -- <project-path> [<project-path>...]
		if len(args) < 4 {
			tprintln(translate("usage.unused"))
			os.Exit(1)
		}

		// find separator `--`
		sep := -1
		for i, a := range args[2:] {
			if a == "--" {
				sep = i + 2
				break
			}
		}

		if sep == -1 {
			tprintln(translate("usage.unused"))
			os.Exit(1)
		}

		fileArgs := args[2:sep]
		projectPaths := args[sep+1:]

		if len(fileArgs) == 0 || len(projectPaths) == 0 {
			tprintln(translate("usage.unused"))
			os.Exit(1)
		}

		files := buildFilesMapFromPaths(fileArgs)

		tm, err := app.NewTranslationManager(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, translate("Error: %v\n"), err)
			os.Exit(1)
		}

		unused, err := tm.FindUnusedKeys(projectPaths)
		if err != nil {
			fmt.Fprintf(os.Stderr, translate("Error: %v\n"), err)
			os.Exit(1)
		}

		if len(unused) == 0 {
			tprintln(translate("unused.all_used"))
		} else {
			tprintf(translate("unused.found_count"), len(unused))
			for _, key := range unused {
				tprintf(translate("unused.item"), key)
			}
		}

	case "add":
		if len(args) < 5 {
			tprintln(translate("usage.add"))
			os.Exit(1)
		}

		filePath := args[2]
		key := args[3]
		value := args[4]

		tm := &app.TranslationManager{}

		if err := tm.AddTranslation(filePath, key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "simple":
		// Usage: i18n-manager simple <translation.json> <key> [<fallback>]
		if len(args) < 4 {
			tprintln(translate("usage.simple"))
			os.Exit(1)
		}

		transPath := args[2]
		key := args[3]
		fallback := ""
		if len(args) >= 5 {
			fallback = args[4]
		}

		// load translation via simpletrans
		t, err := simpletrans.LoadTranslations(transPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, translate("error.loading_translations"), err)
			os.Exit(1)
		}

		out, err := simpletrans.GetTranslation(t, key, nil, fallback)
		if err != nil {
			fmt.Fprintf(os.Stderr, translate("error.rendering_translation"), err)
			os.Exit(1)
		}
		// translation values returned by GetTranslation are final strings; print as-is
		fmt.Println(out)

	default:
		fmt.Fprintf(os.Stderr, translate("error.unknown_command"), command)
		os.Exit(1)
	}
}

// buildFilesMapFromPaths accepts a slice of paths to JSON files and returns a map
// of language -> path. It uses the same tolerant detection rules used in the
// check command (2-letter codes, parent dir, basename, fallback to file-<n>).
// buildFilesMapFromPaths derives language identifiers from paths and returns a map[lang]path.
func buildFilesMapFromPaths(paths []string) map[string]string {
	files := make(map[string]string)
	used := make(map[string]bool)
	langRe := regexp.MustCompile(`^[A-Za-z]{2}([_-][A-Za-z]{2})?$`)

	for idx, p := range paths {
		base := filepath.Base(p)
		name := strings.TrimSuffix(base, filepath.Ext(base))

		var lang string
		if langRe.MatchString(name) {
			lang = name
		} else {
			parent := filepath.Base(filepath.Dir(p))
			if langRe.MatchString(parent) {
				lang = parent
			}
		}

		if lang == "" {
			if name != "" {
				lang = name
			} else {
				lang = fmt.Sprintf("file-%d", idx+1)
			}
		}

		orig := lang
		i := 1
		for used[lang] {
			lang = fmt.Sprintf("%s-%d", orig, i)
			i++
		}
		used[lang] = true
		files[lang] = p
	}

	return files
}
