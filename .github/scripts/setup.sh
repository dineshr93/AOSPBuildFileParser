#!/bin/bash
set -e

echo "=== Go version ==="
go version

echo "=== Initialize module ==="
go mod init github.com/dineshr93/AOSPBuildFileParser || echo "Module already initialized"

echo "=== Tidy module ==="
go mod tidy || echo "No dependencies to tidy"

echo "=== Build ==="
go build -v .
