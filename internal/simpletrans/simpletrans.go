package simpletrans

import (
    "encoding/json"
    "fmt"
    "os"
    "text/template"
)

// Translations maps keys to string or nested maps.
type Translations map[string]interface{}

// LoadTranslations loads a JSON translation file from disk.
func LoadTranslations(filename string) (Translations, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("open %s: %w", filename, err)
    }
    defer f.Close()
    var t Translations
    if err := json.NewDecoder(f).Decode(&t); err != nil {
        return nil, fmt.Errorf("decode %s: %w", filename, err)
    }
    return t, nil
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
        if data == nil {
            return s, nil
        }
        out, err := render(s, data)
        if err != nil {
            return s, nil
        }
        return out, nil
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
