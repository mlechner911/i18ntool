# i18n-manager - Translation Helper

Quick commands:
```bash
# Check for missing translations
./i18n-manager check en.json de.json es.json

# Add new translation (creates backup)
./i18n-manager add en.json "section.key" "Translation text"

# Sort and backup all files
./i18n-manager sort en.json de.json es.json

# Find unused translation keys
./i18n-manager unused en.json de.json es.json src/
```

Features:
- ✅ Automatic backups before changes
- ✅ Nested key support (aa.bb.cc)
- ✅ Never overwrites existing keys
- ✅ Sorted JSON output# i18ntool
