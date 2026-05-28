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

echo "=== Listing build outputs ==="
ls -la aospparse-*

echo "Build successful!"
