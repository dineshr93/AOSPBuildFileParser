#!/bin/bash
# build.sh - Build script for AOSP Build File Parser
set -e

echo "Building aospparse..."
go build -v -o aospparse .
echo "✓ Build successful"
ls -lh aospparse
