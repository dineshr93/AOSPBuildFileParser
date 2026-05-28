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
VERSION="${TAG#v}"

# Create release notes inline
echo "# AOSP Build File Parser v$VERSION" > RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "This release includes:" >> RELEASE_NOTES.md
echo "- Full AOSP tree scanning for .bp and .mk files" >> RELEASE_NOTES.md
echo "- Variable resolution and defaults chain inheritance" >> RELEASE_NOTES.md
echo "- Transitive dependency resolution via BFS traversal" >> RELEASE_NOTES.md
echo "- Source file path extraction for license compliance" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## Features" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "- Full tree scanning - Recursively parse all Android.bp and Android.mk files" >> RELEASE_NOTES.md
echo "- Variable resolution - Resolve variable assignments referenced in module properties" >> RELEASE_NOTES.md
echo "- Defaults chain resolution - Properly handle cc_defaults and similar inheritance" >> RELEASE_NOTES.md
echo "- Transitive dependency resolution - BFS traversal through all dependency chains" >> RELEASE_NOTES.md
echo "- Source path extraction - Map modules to actual source file paths" >> RELEASE_NOTES.md
echo "- Structured JSON output - Machine-readable registry for pipeline integration" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## Usage" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "Scan an AOSP tree: ./aospparse scan /path/to/aosp > module_registry.json" >> RELEASE_NOTES.md
echo "Resolve dependencies: ./aospparse deps module_registry.json libvsomeip3" >> RELEASE_NOTES.md
echo "Extract paths: ./aospparse paths module_registry.json libvsomeip3" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## Downloads" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "| Platform | Architecture | Binary |" >> RELEASE_NOTES.md
echo "|----------|--------------|--------|" >> RELEASE_NOTES.md
echo "| Linux | AMD64 | aospparse-linux-amd64 |" >> RELEASE_NOTES.md
echo "| Linux | ARM64 | aospparse-linux-arm64 |" >> RELEASE_NOTES.md
echo "| macOS | AMD64 | aospparse-darwin-amd64 |" >> RELEASE_NOTES.md
echo "| macOS | ARM64 | aospparse-darwin-arm64 |" >> RELEASE_NOTES.md
echo "| Windows | AMD64 | aospparse-windows-amd64.exe |" >> RELEASE_NOTES.md
echo "| Windows | ARM64 | aospparse-windows-arm64.exe |" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## Changelog" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "See [CHANGELOG.md](CHANGELOG.md) for detailed changes." >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "## License" >> RELEASE_NOTES.md
echo "" >> RELEASE_NOTES.md
echo "Apache 2.0 (parsers from AOSP). Main tool code is MIT." >> RELEASE_NOTES.md

echo "tag_name=$TAG" >> $GITHUB_OUTPUT
echo "is_prerelease=false" >> $GITHUB_OUTPUT
