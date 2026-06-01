#!/bin/bash
# test.sh - Test runner for CI/CD
set -e

echo "=== Running Tests ==="

# Test Go
echo "Testing Go..."
go test ./... -v -timeout=60s || { echo "ERROR: Go tests failed"; exit 1; }

# Test shell
echo "Testing shell..."
bash tests/test_aosp_env.sh || { echo "ERROR: test_aosp_env.sh failed"; exit 1; }
bash tests/test_expand_vars.sh || { echo "ERROR: test_expand_vars.sh failed"; exit 1; }
bash tests/test_expand_vars2.sh || { echo "ERROR: test_expand_vars2.sh failed"; exit 1; }
bash tests/test_with_m.sh || { echo "ERROR: test_with_m.sh failed"; exit 1; }
bash tests/test_with_soong.sh || { echo "ERROR: test_with_soong.sh failed"; exit 1; }

echo "✓ All tests passed"
