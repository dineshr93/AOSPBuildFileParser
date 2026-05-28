#!/bin/bash

set -e

TAG=${GITHUB_REF#refs/tags/}

# Validate tag format (basic semver check)
if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
  echo "Warning: Tag format may not follow semantic versioning: $TAG"
fi

# Detect prerelease based on tag name
IS_PRERELEASE=false
if [[ "$TAG" =~ -alpha|-beta|-rc|-pre ]]; then
  IS_PRERELEASE=true
  echo "Detected prerelease tag: $TAG"
fi

# Generate release notes (simplified markdown without backticks)
VERSION="${TAG#v}"

cat > RELEASE_NOTES.md << EOF
# AOSP Build File Parser v${VERSION}

## What's Changed

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

\`\`\`bash
# Scan an AOSP tree
./aospparse scan /path/to/aosp > module_registry.json

# Resolve transitive dependencies
./aospparse deps module_registry.json libvsomeip3

# Extract source file paths
./aospparse paths module_registry.json libvsomeip3
\`\`\`

## Downloads

Platform | Architecture | Binary
---------|--------------|--------
Linux | AMD64 | aospparse-linux-amd64
Linux | ARM64 | aospparse-linux-arm64
macOS | AMD64 | aospparse-darwin-amd64
macOS | ARM64 | aospparse-darwin-arm64
Windows | AMD64 | aospparse-windows-amd64.exe
Windows | ARM64 | aospparse-windows-arm64.exe

## Changelog

See CHANGELOG.md for detailed changes.

## License

Apache 2.0 (parsers from AOSP). Main tool code is MIT.
EOF

# Output to GITHUB_OUTPUT
echo "release_notes<<EOF" >> $GITHUB_OUTPUT
cat RELEASE_NOTES.md >> $GITHUB_OUTPUT
echo "EOF" >> $GITHUB_OUTPUT
echo "is_prerelease=$IS_PRERELEASE" >> $GITHUB_OUTPUT
