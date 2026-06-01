#!/bin/bash
# setup.sh - Setup script for CI/CD environment
set -e

echo "=== CI/CD Setup ==="

# Check Go
echo "Checking Go..."
go version || { echo "ERROR: Go not found"; exit 1; }

# Check Python
echo "Checking Python..."
python3 --version || { echo "ERROR: Python3 not found"; exit 1; }

# Check bash
echo "Checking Bash..."
bash --version | head -1

# Verify directory structure
echo "Verifying directory structure..."
test -d tests || { echo "ERROR: tests directory missing"; exit 1; }
test -f sources-parser.sh || { echo "ERROR: sources-parser.sh missing"; exit 1; }
test -f sources-parser.py || { echo "ERROR: sources-parser.py missing"; exit 1; }
test -f Makefile || { echo "ERROR: Makefile missing"; exit 1; }
test -f README.md || { echo "ERROR: README.md missing"; exit 1; }

echo "✓ Setup complete"
