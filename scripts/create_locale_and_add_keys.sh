#!/usr/bin/env bash
# Create a new locale JSON file and add some keys using the i18n-manager CLI.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCALES_DIR="$ROOT_DIR/examples/locales"
NEW_LOCALE="$LOCALES_DIR/example_new.json"

echo "Creating new locale file: $NEW_LOCALE"
mkdir -p "$LOCALES_DIR"
if [ ! -f "$NEW_LOCALE" ]; then
  echo "{}" > "$NEW_LOCALE"
  echo "Created empty JSON file."
else
  echo "File already exists."
fi

# Build the CLI (ensures binary exists)
cd "$ROOT_DIR"
make build >/dev/null

# Add some keys
./i18n-manager add "$NEW_LOCALE" welcome.message "Welcome to the example locale"
./i18n-manager add "$NEW_LOCALE" button.save "Save"
./i18n-manager add "$NEW_LOCALE" button.cancel "Cancel"

echo "Added sample keys to $NEW_LOCALE"
cat "$NEW_LOCALE"
