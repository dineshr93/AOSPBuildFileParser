# AOSP Build File Parser

A production-grade tool for parsing Android.mk and Android.bp build files without requiring a full AOSP build environment.

[![CI](https://github.com/dineshr93/AOSPBuildFileParser/actions/workflows/ci.yaml)](https://github.com/dineshr93/AOSPBuildFileParser/actions/workflows/ci.yaml)
[![Release](https://github.com/dineshr93/AOSPBuildFileParser/actions/workflows/release.yaml)](https://github.com/dineshr93/AOSPBuildFileParser/actions/workflows/release.yaml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

This project provides tools to:

1. **Parse Android.mk/Android.bp** files statically (no build env needed)
2. **Expand build variables** like `$(wildcard ...)`, `$(patsubst ...)`, etc.
3. **Extract source files** and dependencies
4. **Generate registry** of modules for dependency analysis
5. **Resolve transitive dependencies** across the build tree

### Hybrid Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   AOSP Build File Parser                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────┐      ┌────────────────────────┐   │
│  │   Static Parser     │      │   AOSP Environment     │   │
│  │   (Primary)         │─────>│   (Fallback)           │   │
│  │                     │      │                        │   │
│  │ • Python parser     │      │ • build/envsetup.sh    │   │
│  │ • No dependencies   │      │ • lunch <target>       │   │
│  │ • Offline capable   │      │ • get_build_var        │   │
│  │ • Fast execution    │      │ • Soong UI             │   │
│  └─────────────────────┘      └────────────────────────┘   │
│                                                             │
│  Auto-detects file type and falls back gracefully          │
│  when complex variable expansion is needed                 │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
AOSPBuildFileParser/
├── .github/
│   ├── workflows/          # CI/CD workflows
│   │   ├── ci.yaml        # Continuous Integration
│   │   └── release.yaml   # Release automation
│   └── scripts/           # Build and release scripts
│       ├── build.sh
│       └── setup.sh
├── bin/                    # Binaries (downloaded or built)
├── tests/                  # Test scripts
│   ├── README.md
│   ├── test_aosp_env.sh
│   ├── test_expand_vars.sh
│   ├── test_expand_vars2.sh
│   ├── test_with_m.sh
│   └── test_with_soong.sh
├── androidmk/              # Android.mk parser module
├── blueprint/              # Android.bp parser module
├── sample/                 # Sample build files for testing
├── sources-parser.sh       # Shell wrapper (auto-detects .mk/.bp)
├── sources-parser.py       # Python parser implementation
├── Makefile                # Build and test targets
├── main.go                 # Go CLI binary
├── go.mod / go.sum         # Go module dependencies
├── CHANGELOG.md            # Version history
├── LICENSE                 # MIT License
└── README.md               # This file
```

## Quick Start

### Installation

```bash
# Option 1: Build from source
git clone https://github.com/dineshr93/AOSPBuildFileParser.git
cd AOSPBuildFileParser
make build

# Option 2: Use the sources parser (no Go required)
bash sources-parser.sh path/to/Android.mk
bash sources-parser.sh path/to/Android.bp
```

### Basic Usage

```bash
# Parse Android.mk
bash sources-parser.sh external/libaom/Android.mk

# Parse Android.bp
bash sources-parser.sh external/libaom/Android.bp

# Use Go CLI for full functionality
./aospparse extract sample/Android.bp deps
./aospparse extract sample/Android.mk LOCAL_MODULE
./aospparse scan sample > registry.json
./aospparse deps registry.json libsome_module
```

## Features

### Variable Expansion

| Function | Description | Example |
|----------|-------------|---------|
| `$(wildcard pattern)` | Match files matching pattern | `$(wildcard *.cpp)` |
| `$(patsubst pattern,replacement,text)` | Pattern substitution | `$(patsubst %.c,%.o,$(src))` |
| `$(filter pattern,text)` | Filter matching words | `$(filter %.cpp,$(src))` |
| `$(notdir name)` | Extract filename | `$(notdir /path/to/file.cpp)` |
| `$(dir name)` | Extract directory | `$(dir /path/to/file.cpp)` |
| `$(call my-dir)` | Current module directory | Used in Android.mk |
| `$(LOCAL_PATH)` | Module path variable | AOSP built-in |

### Build Variable Resolution

- `LOCAL_PATH` - Current module directory
- `LOCAL_MODULE` - Module name being built
- `LOCAL_SRC_FILES` - Source files to compile
- `LOCAL_INCLUDES` - Include directories
- `LOCAL_SHARED_LIBRARIES` - Shared library dependencies
- `LOCAL_STATIC_LIBRARIES` - Static library dependencies

### File Detection

- Files found on filesystem: marked as `(EXISTS)`
- Files matching patterns (wildcards/globs): marked as `(PATTERN)` or `(GLOB PATTERN)`
- Unresolved variables: marked as `#UNRESOLVED#` or `#UNRESOLVED_VAR:NAME#`

## Requirements

### Minimum Requirements

- **Python 3.6+** (for sources-parser.py)
- **No AOSP build environment required** for basic parsing

### Optional: AOSP Build Environment

For enhanced variable expansion (AOSP build env fallback):

```
- prebuilts/build-tools/   - Build tools (awk, ninja, shlib)
- prebuilts/go/           - Go runtime for soong
- build/envsetup.sh       - AOSP environment setup
- build/soong/soong_ui.bash - Soong build system
```

To sync required prebuilts:
```bash
cd /path/to/aosp
.repo/repo/repo sync prebuilts/build-tools prebuilts/go -j4
source build/envsetup.sh
lunch <target>
```

## Usage Examples

### Parse Android.mk

```bash
$ bash sources-parser.sh external/google-breakpad/android/google_breakpad/Android.mk
=== sources-parser.py ===
File: external/google-breakpad/android/google_breakpad/Android.mk
Type: make

Parsing Android.mk: external/google-breakpad/android/google_breakpad/Android.mk
LOCAL_PATH = /home/dinesh/aosp/external/google-breakpad/android/google_breakpad/$(call my-dir)/../..
LOCAL_MODULE = breakpad_client
LOCAL_SRC_FILES [raw] = \\\n

=== Parsing complete ===
```

### Parse Android.bp

```bash
$ bash sources-parser.sh external/libaom/Android.bp
=== sources-parser.py ===
File: external/libaom/Android.bp
Type: blueprint

Parsing Android.bp: external/libaom/Android.bp
  -> /home/dinesh/aosp/external/libaom/examples/av1_dec_fuzzer.cc (EXISTS)

=== Parsing complete ===
```

### Use Go CLI for Full Features

```bash
# Build the binary
make build

# Extract properties from files
./aospparse extract sample/Android.bp srcs
./aospparse extract sample/Android.bp shared_libs
./aospparse extract sample/Android.mk LOCAL_MODULE
./aospparse extract sample/Android.mk LOCAL_SRC_FILES

# Scan entire directory
./aospparse scan sample > sample_registry.json

# Resolve dependencies
./aospparse deps sample_registry.json libvsomeip3

# Resolve source paths
./aospparse paths sample_registry.json libvsomeip3
```

## Makefile Targets

```bash
# Build targets
make build          # Build the binary
make clean          # Remove build artifacts
make test           # Run all tests

# Demo targets
make scan-extract   # Demo: extract properties from sample files
make scan-deps      # Demo: scan tree + resolve dependencies
make scan-paths     # Demo: scan tree + resolve source paths
make demo           # Run all demos

# Release targets
make release TAG=v1.0.0  # Create release (requires TAG)
```

## Supported Patterns

### Android.mk Patterns

```
$(wildcard *.cpp)              # Match all .cpp files
$(wildcard src/*.c)            # Match .c files in subdirectory
$(wildcard dir/*)              # Match all files in directory
$(wildcard dir/*.h include/*.h) # Multiple patterns
$(patsubst %.c,%.o,$(src))     # Pattern substitution
$(filter %.cpp,$(src))         # Filter by pattern
$(notdir $(src))               # Remove path prefix
$(dir $(src))                  # Extract path prefix
$(call my-dir)                 # Current module directory
```

### Android.bp Patterns

```
srcs: ["*.cpp"]                # Glob pattern for sources
srcs: ["src/*.cpp"]            # Subdirectory glob
include_paths: ["include"]     # Include directories
shared_libs: ["libsome"]       # Shared library dependencies
static_libs: ["libsome"]       # Static library dependencies
```

## Error Handling

| Error | Description | Solution |
|-------|-------------|----------|
| Unknown file type | File extension not .mk or .bp | Use correct file extension |
| File not found | Specified file doesn't exist | Check file path |
| Unresolved variables | AOSP variable not defined | May need AOSP build env for full expansion |
| (PATTERN) output | File matches wildcard but doesn't exist yet | File may be created later or in different environment |
| AOSP env not available | Required prebuilts missing | Run `repo sync prebuilts/build-tools prebuilts/go` |

## Architecture

### Shell Wrapper (sources-parser.sh)

- Auto-detects file type from extension
- Calls Python parser with file and type
- Handles basic argument validation
- Provides user-friendly output

### Python Parser (sources-parser.py)

**Parsing Android.mk:**
- Extracts LOCAL_PATH, LOCAL_MODULE, LOCAL_SRC_FILES, LOCAL_INCLUDES
- Expands build variables using regex patterns
- Checks actual filesystem for file existence
- Handles line continuations (backslash)

**Parsing Android.bp:**
- Extracts module names, srcs, include_paths
- Handles glob patterns
- Checks actual filesystem for file existence
- Parses cc_library, cc_binary, etc. modules

### Go CLI (main.go)

- Full-featured AOSP build file parser
- Supports all extraction commands (extract, scan, deps, paths)
- Generates JSON registry for dependency analysis
- Handles transitive dependency resolution

## Benefits

1. **No full AOSP build required** - Works with incomplete trees
2. **Fast parsing** - Direct file parsing without shell evaluation
3. **Works offline** - No network dependencies for static parsing
4. **Accurate** - Static analysis of actual file contents
5. **Graceful fallback** - Falls back to AOSP env only when needed
6. **Production-ready** - Includes CI/CD, tests, release workflow

## Testing

```bash
# Run all tests
make test

# Run specific test
./tests/test_aosp_env.sh
./tests/test_expand_vars.sh
./tests/test_expand_vars2.sh

# Run all integration tests
for test in tests/test_*.sh; do
    echo "Running $test..."
    bash "$test"
done
```

## CI/CD

This project uses GitHub Actions for continuous integration and release automation:

- **CI Workflow** (`/.github/workflows/ci.yaml`): Runs on PRs and pushes
- **Release Workflow** (`/.github/workflows/release.yaml`): Runs on tag creation

## Release Process

```bash
# Create and push a release tag
make release TAG=v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The release workflow will:
1. Run all tests
2. Build cross-platform binaries (Linux, macOS, Windows)
3. Create a GitHub release with assets

## Troubleshooting

### Parser shows "Unresolved" for variables

This is normal when AOSP variables aren't defined in static context. The file paths will still show correctly for actual files.

### No files found despite wildcards

The parser expands wildcards only for files that actually exist on the filesystem. Files matching patterns but not yet created will show as `(PATTERN)`.

### AOSP build env not detected

Run `repo sync prebuilts/build-tools prebuilts/go` to sync required prebuilts, then re-source `build/envsetup.sh`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test`
4. Submit a pull request

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Author

Dinesh R - [@dineshr93](https://github.com/dineshr93)
