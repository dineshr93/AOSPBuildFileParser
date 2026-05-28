# AOSP Build File Parser

[![Go Version](https://img.shields.io/github/go-mod/go-version/dineshr93/AOSPBuildFileParser)](https://github.com/dineshr93/AOSPBuildFileParser)
[![License](https://img.shields.io/github/license/dineshr93/AOSPBuildFileParser)](LICENSE)
[![Releases](https://img.shields.io/github/v/release/dineshr93/AOSPBuildFileParser)](https://github.com/dineshr93/AOSPBuildFileParser/releases)

A comprehensive tool for parsing AOSP (Android Open Source Project) build files (`Android.bp` and `Android.mk`) to build a **central module registry** for **license compliance scanning** (e.g., BlackDuck, Synopsys).

Uses the **official AOSP Soong blueprint parser** and **androidmk parser** — no custom regex-based parsing.

## Features

- **Full tree scanning** — recursively find and parse all `.bp` and `.mk` files in an AOSP tree
- **Variable resolution** — resolves `foo = "bar"` assignments referenced as module properties
- **Defaults chain resolution** — `cc_defaults` inheritance for C/C++ modules
- **Transitive dependency resolution** — BFS graph traversal through all dependency chains
- **Source path extraction** — maps modules to their actual source file paths
- **Structured JSON output** — machine-readable registry for pipeline integration

## Building

```bash
go build -o aospparse .
```

Requires Go 1.18+.

### With Make

```bash
make build
```

## Usage

### 1. Scan the AOSP tree → Build registry

```bash
./aospparse scan ~/aosp > module_registry.json
```

Outputs a JSON registry with:
- Every module name, type, and build file location
- All source files, shared/static/header libs, include dirs
- `cc_defaults` chain with inherited properties
- Total file counts

### 2. Resolve transitive dependencies

```bash
./aospparse deps module_registry.json libvsomeip3
```

BFS traversal through all dependency chains — outputs every module that `libvsomeip3` depends on, transitively.

### 3. Extract source file paths

```bash
./aospparse paths module_registry.json libvsomeip3
```

For a module and all its transitive deps, outputs every source file path relative to the AOSP root.

### 4. Extract a property from a single file (legacy)

```bash
./aospparse extract Android.bp srcs
./aospparse extract Android.mk LOCAL_SRC_FILES
```

## Download Pre-built Binaries

Download ready-to-use binaries from the [Releases](https://github.com/dineshr93/AOSPBuildFileParser/releases) page:

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | AMD64 | aospparse-linux-amd64 |
| Linux | ARM64 | aospparse-linux-arm64 |
| macOS | AMD64 | aospparse-darwin-amd64 |
| macOS | ARM64 | aospparse-darwin-arm64 |
| Windows | AMD64 | aospparse-windows-amd64.exe |
| Windows | ARM64 | aospparse-windows-arm64.exe |

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.

## License

Apache 2.0 (parsers from AOSP). Main tool code is MIT.

## Release Process

This project uses semantic versioning with git tags. To create a new release:

1. Update version in code (if needed)
2. Create and push a tag:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This triggers an automated GitHub Action that:
- Runs tests
- Builds binaries for all platforms
- Creates a GitHub release with detailed release notes
- Uploads pre-built binaries as release assets

## Output Format

### `module_registry.json`

```json
{
  "version": "1.0.0",
  "root_dir": "/path/to/aosp",
  "total_bp_files": 12345,
  "total_mk_files": 6789,
  "modules": {
    "libvsomeip3": {
      "name": "libvsomeip3",
      "type": "cc_library_shared",
      "file": "some/ip/libraries/Android.bp",
      "dir": "some/ip/libraries",
      "srcs": ["src/some.cpp", "src/other.cpp"],
      "shared_libs": ["libcutils", "liblog"],
      "static_libs": ["libprotobuf-cpp-full"],
      "defaults": ["libvsomeip_defaults"],
      "include_dirs": ["some/ip/libraries/include"],
      "all_deps": ["libcutils", "liblog", "libprotobuf-cpp-full"],
      "properties": { ... }
    }
  },
  "defaults": {
    "libvsomeip_defaults": { ... }
  }
}
```

## How It Works

### Phase 1: Scan

1. Walk the AOSP tree, finding all `Android.bp`, `*.bp`, `Android.mk`, `*.mk` files
2. Parse `.bp` files with the official Soong blueprint parser
3. Parse `.mk` files with the official Soong androidmk parser
4. Extract module names, types, source files, and dependency properties
5. Resolve variable assignments referenced in properties
6. Resolve `defaults` chains (inheritance)
7. Output structured JSON registry

### Phase 2: Dependencies & Paths

1. Load the registry JSON
2. BFS through dependency graph from seed modules
3. For each module, resolve source file paths relative to AOSP root
4. Output flat list of all source files for license scanning

## Known Dependency Properties by Module Type

| Type | Dependency Properties |
|------|----------------------|
| `cc_library_shared` | `shared_libs`, `static_libs`, `header_libs`, `defaults` |
| `cc_binary` | `static_libs`, `shared_libs`, `defaults` |
| `cc_test` | `static_libs`, `shared_libs`, `defaults` |
| `java_library` | `libs`, `static_libs`, `defaults` |
| `android_app` | `libs`, `static_libs`, `defaults` |
| `go_binary` | `deps`, `defaults` |
| `cc_defaults` | (inherited by modules referencing it) |

## License Compliance Workflow

```
./aospparse scan ~/aosp > registry.json
./aospparse paths registry.json target_module > source_files.txt
# Feed source_files.txt to BlackDuck/Synopsys scan
blackduck-scan --source-files source_files.txt --project myproject
```

## Parsers

- `blueprint/parser/` — Official AOSP Soong blueprint parser (from `build/soong/androidbp`)
- `androidmk/parser/` — Official AOSP Soong androidmk parser (from `build/soong/androidmk`)

Both are Apache 2.0 licensed from Google/AOSP.

## Differences from Python Scripts (sample/)

The Python scripts in `sample/` were the original approach using `module-info.json` generated by Soong builds. The Go tool replaces them:

| Feature | Python (old) | Go (new) |
|---------|-------------|----------|
| Requires full build? | Yes (`module-info.json`) | No |
| Tree scanning | No (single file) | Yes (recursive) |
| `.mk` support | No | Yes |
| Variable resolution | Partial | Yes |
| Defaults resolution | No | Yes |
| Output format | Mixed | Structured JSON |

## License

Apache 2.0 (parsers from AOSP). Main tool code is MIT.