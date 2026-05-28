#!/bin/bash

set -e

echo "=== Go version ==="
go version

echo "=== Go env ==="
go env

echo "=== go.mod ==="
cat go.mod

echo "=== go.sum ==="
cat go.sum 2>/dev/null || echo "go.sum not found"

echo "=== Files ==="
ls -la

echo "=== Building ==="
go build -v -o aospparse .

echo "=== Build completed ==="
ls -la aospparse
