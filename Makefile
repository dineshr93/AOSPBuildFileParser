# AOSP Build File Parser Makefile
# Production-grade build system with CI/CD integration

.PHONY: all build clean test test-go test-sh test-all scan-extract scan-deps scan-paths \
        demo release release-dry-run help

# Configuration
BINARY_NAME = aospparse
PYTHON_PARSER = sources-parser.py
PYTHON_SHELL = sources-parser.sh
GO = go
GOFLAGS = -v
TEST_TIMEOUT = 60s

# Auto-detect Python
PYTHON ?= python3

# Directories
SRC_DIR = .
TEST_DIR = tests
SAMPLE_DIR = sample

# Go version check
GO_VERSION = $(shell $(GO) version 2>/dev/null | sed -E 's/.*go version go([0-9]+\.[0-9]+).*/\1/')
MIN_GO_VERSION = 1.18

all: build

## Build targets
build: check-go-version
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Built $(BINARY_NAME) successfully"
	@ls -lh $(BINARY_NAME)

check-go-version:
	@echo "Checking Go version..."
	@if [ -z "$(GO_VERSION)" ]; then \
		echo "ERROR: Go not found. Please install Go $(MIN_GO_VERSION)+"; \
		exit 1; \
	fi
	@if [ "$$(echo $(GO_VERSION) | cut -d. -f1)" -lt 1 ] || \
	   ([ "$$(echo $(GO_VERSION) | cut -d. -f1)" -eq 1 ] && \
	    [ "$$(echo $(GO_VERSION) | cut -d. -f2)" -lt $(MIN_GO_VERSION) ]); then \
		echo "ERROR: Go version $(GO_VERSION) is below minimum $(MIN_GO_VERSION)"; \
		exit 1; \
	fi
	@echo "✓ Go $(GO_VERSION) detected"

## Clean targets
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	$(GO) clean -cache -testcache
	rm -rf __pycache__
	rm -f *.pyc
	rm -f sample_registry.json
	@echo "✓ Clean complete"

## Test targets
test: test-go test-sh
	@echo "✓ All tests passed"

test-go:
	@echo "=== Running Go tests ==="
	$(GO) test ./... -v -timeout=$(TEST_TIMEOUT) -coverprofile=coverage-go.txt
	@echo ""
	@echo "Go test coverage:"
	$(GO) tool cover -func=coverage-go.txt | grep total
	@echo "✓ Go tests complete"

test-sh:
	@echo "=== Running shell tests ==="
	@test -d $(TEST_DIR) || { echo "ERROR: tests directory not found"; exit 1; }
	@test -f $(TEST_DIR)/test_aosp_env.sh || { echo "ERROR: test_aosp_env.sh not found"; exit 1; }
	bash $(TEST_DIR)/test_aosp_env.sh
	bash $(TEST_DIR)/test_expand_vars.sh
	bash $(TEST_DIR)/test_expand_vars2.sh
	bash $(TEST_DIR)/test_with_m.sh
	bash $(TEST_DIR)/test_with_soong.sh
	@echo "✓ Shell tests complete"

test-all: test-go test-sh

## Demo targets (for development)
scan-extract: build
	@echo "=== Extracting srcs from sample/Android.bp ==="
	./$(BINARY_NAME) extract $(SAMPLE_DIR)/Android.bp deps 2>&1 || echo "Note: sample/Android.bp structure may not match expected format"
	@echo ""
	@echo "=== Extracting LOCAL_MODULE from sample/Android.mk ==="
	./$(BINARY_NAME) extract $(SAMPLE_DIR)/Android.mk LOCAL_MODULE
	@echo ""
	@echo "=== Extracting LOCAL_SRC_FILES from sample/Android.mk ==="
	./$(BINARY_NAME) extract $(SAMPLE_DIR)/Android.mk LOCAL_SRC_FILES

scan-deps: build
	@echo "=== Full scan of sample directory ==="
	./$(BINARY_NAME) scan $(SAMPLE_DIR) > sample_registry.json 2>/dev/null
	@echo ""
	@echo "=== Transitive deps for libvsomeip3 ==="
	./$(BINARY_NAME) deps sample_registry.json libvsomeip3 2>/dev/null

scan-paths: build scan-deps
	@echo ""
	@echo "=== Source paths for libvsomeip3 ==="
	./$(BINARY_NAME) paths sample_registry.json libvsomeip3 2>/dev/null

demo: scan-extract scan-deps scan-paths

## Sources parser targets
sources-test:
	@echo "=== Testing sources parser ==="
	@test -f $(PYTHON_SHELL) || { echo "ERROR: $(PYTHON_SHELL) not found"; exit 1; }
	@test -f $(PYTHON_PARSER) || { echo "ERROR: $(PYTHON_PARSER) not found"; exit 1; }
	bash $(PYTHON_SHELL) $(SAMPLE_DIR)/Android.bp 2>&1
	bash $(PYTHON_SHELL) $(SAMPLE_DIR)/Android.mk 2>&1
	@echo "✓ Sources parser tests complete"

## Release targets
release: check-go-version test release-dry-run
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required. Usage: make release TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating release $(TAG)..."
	@echo "Building release binaries..."
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Built release binaries successfully"
	@echo ""
	@echo "Next steps:"
	@echo "  1. git tag -a $(TAG) -m \"Release $(TAG)\""
	@echo "  2. git push origin $(TAG)"
	@echo ""
	@echo "Or use GitHub CLI:"
	@echo "  gh release create $(TAG) --title \"Release $(TAG)\" --notes \"Release $(TAG)\""

release-dry-run:
	@echo "=== Release dry run ==="
	@echo "Checking for uncommitted changes..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "ERROR: Uncommitted changes detected:"; \
		git status --porcelain; \
		exit 1; \
	fi
	@echo "✓ No uncommitted changes"
	@echo "Checking for upstream divergence..."
	@if [ "$$(git rev-list --count HEAD..origin/main 2>/dev/null || echo 0)" -gt 0 ]; then \
		echo "WARNING: Local branch is behind origin/main"; \
	fi
	@echo "✓ Ready for release"

## Help target
help:
	@echo "AOSP Build File Parser - Makefile targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build              - Build the binary (requires Go $(MIN_GO_VERSION)+)"
	@echo "  check-go-version   - Verify Go version"
	@echo "  clean              - Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test               - Run all tests (Go + shell)"
	@echo "  test-go            - Run Go unit tests"
	@echo "  test-sh            - Run shell integration tests"
	@echo "  test-all           - Run all tests (alias for test)"
	@echo "  sources-test       - Test sources parser only"
	@echo ""
	@echo "Demo targets (for development):"
	@echo "  scan-extract       - Demo: extract properties from sample files"
	@echo "  scan-deps          - Demo: scan tree + resolve dependencies"
	@echo "  scan-paths         - Demo: scan tree + resolve source paths"
	@echo "  demo               - Run all demos"
	@echo ""
	@echo "Release targets:"
	@echo "  release TAG=vX.Y.Z - Create GitHub release (requires TAG)"
	@echo "  release-dry-run    - Pre-release validation checks"
	@echo ""
	@echo "Sources parser (no Go required):"
	@echo "  sources-test       - Test sources parser only"
	@echo ""
	@echo "Examples:"
	@echo "  make build                              - Build aospparse binary"
	@echo "  make test                               - Run all tests"
	@echo "  make demo                               - Run all demos"
	@echo "  make release TAG=v1.0.0                 - Create release"
	@echo "  bash sources-parser.sh sample/Android.bp - Parse Android.bp"
	@echo ""
