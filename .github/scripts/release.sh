#!/bin/bash
set -e

echo "=== Building for Linux AMD64 ==="
GOOS=linux GOARCH=amd64 go build -v -o aospparse-linux-amd64 .

echo "=== Building for Linux ARM64 ==="
GOOS=linux GOARCH=arm64 go build -v -o aospparse-linux-arm64 .

echo "=== Building for macOS AMD64 ==="
GOOS=darwin GOARCH=amd64 go build -v -o aospparse-darwin-amd64 .

echo "=== Building for macOS ARM64 ==="
GOOS=darwin GOARCH=arm64 go build -v -o aospparse-darwin-arm64 .

echo "=== Building for Windows AMD64 ==="
GOOS=windows GOARCH=amd64 go build -v -o aospparse-windows-amd64.exe .

echo "=== Building for Windows ARM64 ==="
GOOS=windows GOARCH=arm64 go build -v -o aospparse-windows-arm64.exe .

echo "=== Validating build outputs ==="
if [[ ! -f "aospparse-linux-amd64" ]]; then echo "ERROR: aospparse-linux-amd64 not found"; exit 1; fi
if [[ ! -f "aospparse-linux-arm64" ]]; then echo "ERROR: aospparse-linux-arm64 not found"; exit 1; fi
if [[ ! -f "aospparse-darwin-amd64" ]]; then echo "ERROR: aospparse-darwin-amd64 not found"; exit 1; fi
if [[ ! -f "aospparse-darwin-arm64" ]]; then echo "ERROR: aospparse-darwin-arm64 not found"; exit 1; fi
if [[ ! -f "aospparse-windows-amd64.exe" ]]; then echo "ERROR: aospparse-windows-amd64.exe not found"; exit 1; fi
if [[ ! -f "aospparse-windows-arm64.exe" ]]; then echo "ERROR: aospparse-windows-arm64.exe not found"; exit 1; fi
echo "All binaries validated successfully!"

echo "=== Listing build outputs ==="
ls -la aospparse-*

echo "=== Generating release notes ==="
TAG=${GITHUB_REF#refs/tags/}

cat > RELEASE_NOTES.md << 'NOTES_EOF'
# AOSP Build File Parser TAG_PLACEHOLDER

This release includes:
- Full AOSP tree scanning for .bp and .mk files
- Variable resolution and defaults chain inheritance
- Transitive dependency resolution via BFS traversal
- Source file path extraction for license compliance

## Features

- Full tree scanning - Recursively parse all Android.bp and Android.mk files
- Variable resolution - Resolve variable assignments referenced in module properties
- Defaults chain resolution - Properly handle cc_defaults and similar inheritance
- Transitive dependency resolution - BFS traversal through all dependency chains
- Source path extraction - Map modules to actual source file paths
- Structured JSON output - Machine-readable registry for pipeline integration

## Usage

# Scan an AOSP tree
./aospparse scan /path/to/aosp > module_registry.json

# Resolve transitive dependencies
./aospparse deps module_registry.json libvsomeip3

# Extract source file paths
./aospparse paths module_registry.json libvsomeip3

## Downloads

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | AMD64 | aospparse-linux-amd64 |
| Linux | ARM64 | aospparse-linux-arm64 |
| macOS | AMD64 | aospparse-darwin-amd64 |
| macOS | ARM64 | aospparse-darwin-arm64 |
| Windows | AMD64 | aospparse-windows-amd64.exe |
| Windows | ARM64 | aospparse-windows-arm64.exe |

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.

## License

Apache 2.0 (parsers from AOSP). Main tool code is MIT.
NOTES_EOF

sed -i "s/TAG_PLACEHOLDER/$TAG/" RELEASE_NOTES.md

echo "tag_name=$TAG" >> $GITHUB_OUTPUT
echo "is_prerelease=false" >> $GITHUB_OUTPUT
