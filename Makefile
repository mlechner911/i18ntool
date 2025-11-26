.PHONY: build clean test test-check test-sort test-unused help precheckin examples-test examples-clean

# Absolute paths (can be overridden). Default to repository root.
BASE_DIR ?= $(CURDIR)
TOOLS_DIR=$(BASE_DIR)

BINARY_NAME=$(TOOLS_DIR)/i18n-manager
SOURCE_DIR=$(TOOLS_DIR)/cmd/i18n-manager
TEST_DIR=$(TOOLS_DIR)/testfiles
PROJECT_SRC=$(BASE_DIR)/examples/example_app

# If frontend/src doesn't exist in this workspace (examples/tests), fall back
ifeq (,$(wildcard $(PROJECT_SRC)))
PROJECT_SRC=$(TEST_DIR)
endif

# Real i18n files
I18N_DIR=$(BASE_DIR)/examples/locales
I18N_EN=$(I18N_DIR)/en.json
I18N_DE=$(I18N_DIR)/de.json
I18N_ES=$(I18N_DIR)/es.json

# Installation prefix (can be overridden):
PREFIX ?= /usr/local

install: build
	@echo "Installing $(BINARY_NAME) to $(PREFIX)/bin"
	cp -f $(BINARY_NAME) $(PREFIX)/bin/
	@echo "Installing manpage to $(PREFIX)/share/man/man1"
	mkdir -p $(PREFIX)/share/man/man1
	cp -f man/man1/i18n-manager.1 $(PREFIX)/share/man/man1/
	@echo "Install complete."

uninstall:
	@echo "Removing $(PREFIX)/bin/i18n-manager and manpage"
	rm -f $(PREFIX)/bin/i18n-manager
	rm -f $(PREFIX)/share/man/man1/i18n-manager.1
	@echo "Uninstall complete."

help:
	@echo "Available targets:"
	@echo "  make build        - Build the Go binary"
	@echo "  make test         - Run all tests (testfiles)"
	@echo "  make test-check   - Test missing translations check"
	@echo "  make test-sort    - Test JSON sorting"
	@echo "  make test-unused  - Test unused keys detection"
	@echo "  make examples-test - Run example tests (examples/locales)"
	@echo "  make examples-clean - Remove backups from examples/locales"
	@echo "  make precheckin   - Sort, check and validate real i18n files"
	@echo "  make clean        - Remove binary and backup files"


build: embed-locales
	@echo "Building $(BINARY_NAME)..."
	@cd $(TOOLS_DIR) && if [ ! -f "go.mod" ]; then \
		echo "Initializing Go module..."; \
		go mod init github.com/mlechner911/i18ntool; \
	fi
	cd $(TOOLS_DIR) && go build -o $(BINARY_NAME) ./cmd/i18n-manager
	@echo "Build complete!"

test: build test-check test-sort test-unused

test-check: build
	@echo "\n=== Testing: Check for missing translations ==="
	$(BINARY_NAME) check $(TEST_DIR)/en.json $(TEST_DIR)/de.json $(TEST_DIR)/es.json

test-sort: build
	@echo "\n=== Testing: Sort JSON files ==="
	$(BINARY_NAME) sort $(TEST_DIR)/en.json $(TEST_DIR)/de.json $(TEST_DIR)/es.json

test-unused: build
	@echo "\n=== Testing: Find unused keys ==="
	$(BINARY_NAME) unused $(TEST_DIR)/en.json $(TEST_DIR)/de.json $(TEST_DIR)/es.json -- $(PROJECT_SRC)


precheckin: build
	@echo "\n=== Pre-Checkin: Processing example i18n files ==="
	@echo "\n1. Sorting and backing up files..."
	$(BINARY_NAME) sort $(I18N_EN) $(I18N_DE) $(I18N_ES)
	@echo "\n2. Checking for missing translations..."
	$(BINARY_NAME) check $(I18N_EN) $(I18N_DE) $(I18N_ES)
	# @echo "\n3. Finding unused translation keys..."
	# $(BINARY_NAME) unused $(I18N_EN) $(I18N_DE) $(I18N_ES) $(PROJECT_SRC)
	@echo "\n=== Pre-Checkin complete ==="

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f $(TEST_DIR)/*.backup.*
	@echo "Clean complete!"

examples-test: build
	@echo "Running example tests (examples/locales)"
	./scripts/run_examples_tests.sh

examples-clean:
	@echo "Removing backup files from examples/locales"
	rm -f $(TOOLS_DIR)/examples/locales/*.backup.*
	@echo "examples/locales cleaned"

embed-locales:
	@echo "Embedding example locales into Go source (internal/simpletrans/embedded_translations.go)"
	python3 scripts/embed_locales_to_go.py
	@echo "Embedded translations generated."
