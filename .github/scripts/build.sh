#!/bin/bash
set -e

echo "=== Go version ==="
go version

echo "=== Go env ==="
go env

echo "=== Check go.mod ==="
cat go.mod

echo "=== Check go.sum ==="
cat go.sum 2>/dev/null || echo "go.sum not found"

echo "=== Build ==="
go build -v .

echo "=== Cross-build ==="
GOOS=linux GOARCH=amd64 go build -o aospparse-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o aospparse-linux-arm64 .
GOOS=darwin GOARCH=amd64 go build -o aospparse-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o aospparse-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o aospparse-windows-amd64.exe .

echo "=== Done ==="
ls -la aospparse-*
