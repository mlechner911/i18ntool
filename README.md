# i18n-manager - Translation Helper
Author: Michael Lechner
Copyright: 2025

Lightweight CLI for managing JSON translation files: checking missing translations, sorting and backing up, adding keys, and detecting unused keys.

Motivation
-----
I'm constantly updating translations for a new project and frequently run into
missing or obsolete keys. This tool helps identify unused keys and missing
translations, and automates much of the routine translation work.

I plan to add a lightweight wrapper so it can be integrated into AI-assisted workflows.

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

- simple: Load a single translation JSON and print a key's value (supports optional fallback).

```bash
# usage: i18n-manager simple <translation.json> <key> [<fallback>]
./i18n-manager simple locales/en.json messages.welcome "[MISSING]"
./i18n-manager simple locales/de.json messages.welcome "[MISSING]"
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

Embedding translations into the binary
-------------------------------------
You can embed the example locale JSON files into Go source so translations are available
even when external files are not present. This is handy for builds where you want a
compiled fallback.

- Generate embedded translations (writes `internal/simpletrans/embedded_translations.go`):

```bash
make embed-locales
```

- The generator script is `scripts/embed_locales_to_go.py` and can be run directly:

```bash
python3 scripts/embed_locales_to_go.py
```

- After generating, rebuild the project:

```bash
go build ./...
```

The runtime will prefer external JSON files when present, and fall back to the embedded
map if a requested language file is missing.

Install & Packaging
-------------------
You can install the built binary and manpage to your system or create a staged package.

- Install the binary and manpage to `/usr/local` (default):

```bash
sudo make install
```

What `make install` installs
----------------------------
- **Binary:** `$(PREFIX)/bin/i18n-manager` (installed to `/usr/local/bin` by default)
- **Manpage:** `$(PREFIX)/share/man/man1/i18n-manager.1` (installed to `/usr/local/share/man/man1` by default)

- The `install` target does **not** copy the `Makefile`, `README.md`, `LICENSE` or other repository files. Use `DESTDIR`/`package` targets if you need to collect additional files into a staged archive.

- Use `DESTDIR` for staged installs (useful for packaging):

```bash
make install DESTDIR=/tmp/package-root
# binary will be at /tmp/package-root/usr/local/bin/i18n-manager
```

- Create a tarball package (stages a `DESTDIR` install into `dist/root`):

```bash
make package
# -> dist/i18n-manager-YYYYMMDD-HHMMSS.tar.gz
```

- The package target also writes a SHA256 checksum next to the tarball.

Default Make Behavior
---------------------
- The Makefile default goal is `build` to avoid accidental installs when running `make` with no arguments.  Use `make install` explicitly to install.

Cleaning
--------
- `make clean` removes the built binary, example backup files and the `dist/` directory produced by `make package`.

Manpage
-------
The manpage is provided at `man/man1/i18n-manager.1` and is installed by `make install`.

Notes
-----
- `make build` automatically regenerates embedded translations with `make embed-locales`.
- To regenerate embedded translations manually:

```bash
python3 scripts/embed_locales_to_go.py
```


Most unit tests, the docs and the man pages are AI generated.
