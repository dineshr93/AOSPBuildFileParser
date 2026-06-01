# Tests Directory

This directory contains test scripts for the AOSP Build File Parser.

## Test Scripts

### Unit Tests (Go)

Run with: `make test`

### Integration Tests

- `test_aosp_env.sh` - Tests AOSP environment detection
- `test_expand_vars.sh` - Tests variable expansion (basic)
- `test_expand_vars2.sh` - Tests variable expansion (advanced)
- `test_with_m.sh` - Tests integration with AOSP `m` command
- `test_with_soong.sh` - Tests integration with AOSP Soong build system

## Running Tests

```bash
# All tests
make test

# Specific test
./tests/test_aosp_env.sh

# All integration tests
bash tests/test_aosp_env.sh
bash tests/test_expand_vars.sh
bash tests/test_expand_vars2.sh
```

## Test Coverage

- Variable expansion (wildcard, patsubst, filter, notdir, dir)
- File existence detection
- Android.mk parsing
- Android.bp parsing
- AOSP environment fallback
