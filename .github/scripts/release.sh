#!/bin/bash
set -e

echo "=== Building binaries ==="
GOOS=linux GOARCH=amd64 go build -v -o aospparse-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -v -o aospparse-linux-arm64 .
GOOS=darwin GOARCH=amd64 go build -v -o aospparse-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -v -o aospparse-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -v -o aospparse-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -v -o aospparse-windows-arm64.exe .

echo "=== Verifying builds ==="
ls -la aospparse-*

echo "=== Creating release notes ==="
TAG=${GITHUB_REF#refs/tags/}

cat > RELEASE_NOTES.md << 'EOF'
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

## Downloads

Linux AMD64: aospparse-linux-amd64
Linux ARM64: aospparse-linux-arm64
macOS AMD64: aospparse-darwin-amd64
macOS ARM64: aospparse-darwin-arm64
Windows AMD64: aospparse-windows-amd64.exe
Windows ARM64: aospparse-windows-arm64.exe
EOF

sed -i "s/TAG_PLACEHOLDER/$TAG/" RELEASE_NOTES.md

echo "is_prerelease=false" >> $GITHUB_OUTPUT
echo "tag_name=$TAG" >> $GITHUB_OUTPUT
