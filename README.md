# i18n-manager - Translation Helper

Quick commands:
# i18n-manager - Translation Helper

Lightweight CLI for managing JSON translation files: checking missing translations, sorting and backing up, adding keys, and detecting unused keys.

Build
-----
From the repository root:

```bash
make build
```

Usage
-----
General form:

```bash
i18n-manager <command> [options]
```

Commands
- check: Detect missing translations across N JSON files. Language code is derived from filename (e.g. `en.json`) or parent directory; falls back to `file-1`, `file-2`, ... for reporting.

```bash
# example: check any number of files
./i18n-manager check examples/locales/en.json examples/locales/de.json examples/locales/es.json
```

- sort: Sort and save the provided JSON files (creates backups with timestamps).

```bash
./i18n-manager sort examples/locales/en.json examples/locales/de.json examples/locales/es.json
```

- unused: Find keys that are not referenced in project source files. Use `--` to separate translation files from project path(s):

```bash
# usage: i18n-manager unused <file1.json> <file2.json> -- <project-path> [<project-path>...]
./i18n-manager unused examples/locales/en.json examples/locales/de.json -- ./frontend/src
```

- add: Add a new translation key to a single file (creates a backup). Example:

```bash
./i18n-manager add examples/locales/en.json some.section.key "My string"
```

Examples
--------
- Example locales are in `examples/locales/` (en/de/es/fr). A small demo `examples/example_app` shows usage (it's only a demo).
- Run a quick smoke test (build + basic checks):

```bash
./scripts/run_examples_tests.sh
```

Cleaning up example backups
---------------------------
- Backups created by `sort` are stored next to the file as `*.backup.<timestamp>`. Use the Makefile target to remove example backups:

```bash
make examples-clean
```

Development notes
-----------------
- Code layout: CLI in `cmd/i18n-manager`, core logic in `internal/app` split among small files.
- License: MIT — see `LICENSE`.

If you want a different language-detection policy (strict two-letter codes, directory-only, or explicit `lang:filepath` syntax), I can add flags or stricter validation.
