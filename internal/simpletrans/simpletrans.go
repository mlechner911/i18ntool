package simpletrans

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// Translations maps keys to string or nested maps.
type Translations map[string]interface{}

// LoadTranslations loads a JSON translation file from disk.
func LoadTranslations(filename string) (Translations, error) {
	f, err := os.Open(filename)
	if err == nil {
		defer f.Close()
		var t Translations
		if err := json.NewDecoder(f).Decode(&t); err == nil {
			return t, nil
		}
		// if decode failed, fallthrough to embedded fallback
	}

	// attempt to detect language code from filename and use embedded translations
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if m, ok := EmbeddedTranslations[name]; ok {
		return buildTranslationsFromFlat(m), nil
	}

	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filename, err)
	}
	return nil, fmt.Errorf("decode %s: failed to parse JSON", filename)
}

// buildTranslationsFromFlat converts a flat map (dotted keys) into a nested Translations
// while also keeping the flat keys available at the top-level. This ensures both
// dotted-key traversal and top-level lookups work.
func buildTranslationsFromFlat(flat map[string]string) Translations {
	out := make(Translations)
	// keep flat entries
	for k, v := range flat {
		out[k] = v
	}
	// also build nested maps for dotted keys
	for k, v := range flat {
		parts := splitKey(k)
		cur := map[string]interface{}(out)
		for i, p := range parts {
			if i == len(parts)-1 {
				cur[p] = v
			} else {
				if nxt, ok := cur[p].(map[string]interface{}); ok {
					cur = nxt
				} else {
					nm := make(map[string]interface{})
					cur[p] = nm
					cur = nm
				}
			}
		}
	}
	return out
}

// GetTranslation returns the string for key or the fallback if missing.
func GetTranslation(t Translations, key string, data map[string]interface{}, fallback string) (string, error) {
	parts := splitKey(key)
	var cur interface{} = t
	for _, p := range parts {
		switch m := cur.(type) {
		case map[string]interface{}:
			cur = m[p]
		case Translations:
			cur = mapInterface(m)[p]
		default:
			cur = nil
		}
		if cur == nil {
			break
		}
	}
	if cur == nil {
		return fallback, nil
	}
	if s, ok := cur.(string); ok {
		// perform template rendering if needed
		var err error
		if data != nil {
			var out string
			out, err = render(s, data)
			if err == nil {
				s = out
			}
		}
		// unescape common escaped sequences like \n, \t so output contains real newlines
		s = unescapeCommon(s)
		return s, nil
	}
	return fallback, nil
}

// render executes a text/template using data map.
func render(tmpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("msg").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var out string
	if err := t.Execute(&stringWriter{&out}, data); err != nil {
		return "", err
	}
	return out, nil
}

// stringWriter appends bytes to a string.
type stringWriter struct{ s *string }

func (w *stringWriter) Write(p []byte) (int, error) {
	*w.s += string(p)
	return len(p), nil
}

// splitKey splits dotted keys.
func splitKey(k string) []string {
	var parts []string
	cur := ""
	for i := 0; i < len(k); i++ {
		if k[i] == '.' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
			continue
		}
		cur += string(k[i])
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}

// mapInterface converts Translations to map[string]interface{}.
func mapInterface(t Translations) map[string]interface{} {
	m := make(map[string]interface{}, len(t))
	for k, v := range t {
		m[k] = v
	}
	return m
}

// unescapeCommon converts sequences like "\n" into real newlines.
// It tries to use strconv.Unquote on a quoted string; if that fails
// it falls back to replacing common escapes.
func unescapeCommon(s string) string {
	// Quick path: if the string contains backslash escapes, try Unquote
	if !strings.Contains(s, "\\") {
		return s
	}
	// strconv.Unquote expects a quoted string; try it first
	if q, err := strconv.Unquote("\"" + s + "\""); err == nil {
		return q
	}

	// Manual fallback: handle common escapes and unicode sequences
	// Replace escaped backslash first
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")

	// Handle \uXXXX sequences
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			hex := s[i+2 : i+6]
			// parse hex
			var r rune
			if n, err := strconv.ParseInt(hex, 16, 32); err == nil {
				r = rune(n)
				b.WriteRune(r)
				i += 5 // skip \uXXXX
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Unescape exposes unescapeCommon for callers that need to normalize
// translation strings (turning "\\n" into a newline, etc.).
func Unescape(s string) string {
	return unescapeCommon(s)
}
