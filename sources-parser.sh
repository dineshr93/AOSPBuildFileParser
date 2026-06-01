#!/bin/bash
# sources-parser.sh - Shell wrapper for AOSP build file parser
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_SCRIPT="$SCRIPT_DIR/sources-parser.py"

# Auto-detect file type
detect_file_type() {
    local file="$1"
    if [[ "$file" == *.mk ]]; then
        echo "make"
    elif [[ "$file" == *.bp ]]; then
        echo "blueprint"
    else
        echo "unknown"
    fi
}

if [[ $# -lt 1 ]]; then
    echo "Usage: sources-parser.sh <Android.mk|Android.bp>"
    echo ""
    echo "Auto-detects file type from extension."
    exit 1
fi

file="$1"
file_type=$(detect_file_type "$file")

if [[ "$file_type" == "unknown" ]]; then
    echo "Error: Unknown file type. Expected .mk or .bp file."
    exit 1
fi

python3 "$PYTHON_SCRIPT" "$file" "$file_type"
echo ""
echo "=== Parsing complete ==="
