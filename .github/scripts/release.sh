#!/bin/bash
# release.sh - Release script for CI/CD
set -e

if [ -z "$TAG" ]; then
    echo "ERROR: TAG environment variable is required"
    echo "Usage: TAG=v1.0.0 bash release.sh"
    exit 1
fi

echo "=== Release Script for $TAG ==="

# Build binaries
echo "Building release binaries..."
GOOS=linux GOARCH=amd64 go build -o aospparse-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o aospparse-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o aospparse-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o aospparse-windows-amd64.exe .

echo "Built binaries:"
ls -lh aospparse-*

echo "Release prepared for $TAG"
