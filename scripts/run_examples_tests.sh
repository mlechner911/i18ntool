#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "Building CLI binary..."
make build

BINARY="$ROOT_DIR/i18n-manager"
LOCALES_DIR="$ROOT_DIR/examples/locales"

echo
echo "=== Test: check translations ==="
"$BINARY" check "$LOCALES_DIR/en.json" "$LOCALES_DIR/de.json" "$LOCALES_DIR/es.json" || true

echo
echo "=== Test: sort locales (will create backups in-place) ==="
"$BINARY" sort "$LOCALES_DIR/en.json" "$LOCALES_DIR/de.json" "$LOCALES_DIR/es.json" || true

echo
echo "=== Test: unused keys (scans current repo as example project) ==="
"$BINARY" unused "$LOCALES_DIR/en.json" "$LOCALES_DIR/de.json" "$LOCALES_DIR/es.json" "$ROOT_DIR" || true

echo
echo "=== Test: run example_app (expected to be a demo) ==="
if go build -o /tmp/example_app "$ROOT_DIR/examples/example_app" 2>/dev/null; then
  echo "example_app built successfully (unexpected)"
  /tmp/example_app || true
else
  echo "example_app did not build (this is expected for the demo)."
fi

echo
echo "All example tests completed (some failures may be expected)."
