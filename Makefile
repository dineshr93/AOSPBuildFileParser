.PHONY: all build clean test scan-extract scan-deps scan-paths

BINARY_NAME = aospparse
GO = go
GOFLAGS = -v

all: build

build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "Built $(BINARY_NAME) successfully"

clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

test:
	$(GO) test ./... -v

# Legacy extract command
scan-extract: build
	@echo "=== Extracting srcs from sample/Android.bp ==="
	./$(BINARY_NAME) extract sample/Android.bp srcs
	@echo ""
	@echo "=== Extracting deps from sample/Android.bp ==="
	./$(BINARY_NAME) extract sample/Android.bp shared_libs
	@echo ""
	@echo "=== Extracting LOCAL_MODULE from sample/Android.mk ==="
	./$(BINARY_NAME) extract sample/Android.mk LOCAL_MODULE
	@echo ""
	@echo "=== Extracting LOCAL_SRC_FILES from sample/Android.mk ==="
	./$(BINARY_NAME) extract sample/Android.mk LOCAL_SRC_FILES

# Full scan demo
scan-deps: build
	@echo "=== Full scan of sample directory ==="
	./$(BINARY_NAME) scan sample > sample_registry.json 2>/dev/null
	@echo ""
	@echo "=== Transitive deps for libvsomeip3 ==="
	./$(BINARY_NAME) deps sample_registry.json libvsomeip3 2>/dev/null

scan-paths: build scan-deps
	@echo ""
	@echo "=== Source paths for libvsomeip3 ==="
	./$(BINARY_NAME) paths sample_registry.json libvsomeip3 2>/dev/null

# Run all demos
demo: scan-extract scan-deps scan-paths

help:
	@echo "AOSP Build File Parser - Makefile targets:"
	@echo "  build        - Build the binary"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  scan-extract - Demo: extract properties from sample files"
	@echo "  scan-deps    - Demo: scan tree + resolve dependencies"
	@echo "  scan-paths   - Demo: scan tree + resolve source paths"
	@echo "  demo         - Run all demos"
	@echo "  help         - Show this help"
