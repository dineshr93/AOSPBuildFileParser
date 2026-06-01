// AOSP Build File Parser - License Compliance Tool
//
// Scans AOSP .bp and .mk files to build a module registry for license compliance.
// Uses the official Soong blueprint and androidmk parsers.
//
// Usage:
//   aospparse scan   <root_dir>            Scan tree, output module_registry.json
//   aospparse deps   <registry.json> <modules...>  Resolve transitive deps
//   aospparse paths  <registry.json> <modules...>  Resolve source file paths
//   aospparse extract <file.bp|mk> <key>   Extract a single property (legacy)

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	bkparser "github.com/dineshr93/AOSPBuildFileParser/blueprint/parser"
	mkparser "github.com/dineshr93/AOSPBuildFileParser/androidmk/parser"
)

// ModuleInfo represents a parsed module from a .bp or .mk file.
type ModuleInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	File        string   `json:"file"`            // relative path to .bp/.mk
	Dir         string   `json:"dir"`             // directory containing the build file
	Srcs        []string `json:"srcs,omitempty"`
	SharedLibs  []string `json:"shared_libs,omitempty"`
	StaticLibs  []string `json:"static_libs,omitempty"`
	HeaderLibs  []string `json:"header_libs,omitempty"`
	Defaults    []string `json:"defaults,omitempty"`
	AllDeps     []string `json:"all_deps,omitempty"` // combined dependency list
	IncludeDirs []string `json:"include_dirs,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"` // raw properties
}

// Registry is the output of a full tree scan.
type Registry struct {
	Version   string            `json:"version"`
	RootDir   string            `json:"root_dir"`
	Modules   map[string]ModuleInfo `json:"modules"` // keyed by module name
	Defaults  map[string]ModuleInfo `json:"defaults"` // keyed by module name (cc_defaults etc)
	TotalBP   int               `json:"total_bp_files"`
	TotalMK   int               `json:"total_mk_files"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "scan":
		cmdScan(os.Args[2:])
	case "deps":
		cmdDeps(os.Args[2:])
	case "paths":
		cmdPaths(os.Args[2:])
	case "sources":
		cmdSources(os.Args[2:])
	case "extract":
		cmdExtract(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`AOSP Build File Parser - License Compliance Tool

Usage:
  aospparse scan   <root_dir>              Scan AOSP tree, output module_registry.json
  aospparse deps   <registry.json> <mods>  Resolve transitive dependencies
  aospparse paths  <registry.json> <mods>  Resolve source file paths
  aospparse sources <registry.json> <mods> Resolve all sources (with dedup and gather)
  aospparse extract <file> <key>           Extract a property from a single file

Examples:
  aospparse scan ~/aosp > module_registry.json
  aospparse deps module_registry.json libvsomeip3
  aospparse paths module_registry.json libvsomeip3
  aospparse sources module_registry.json libvsomeip3 --gather --dest /tmp/sources
  aospparse extract Android.bp srcs`)
}

// ============================================================================
// SCAN command - full tree scan
// ============================================================================

func cmdScan(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: aospparse scan <root_dir>")
		os.Exit(1)
	}

	rootDir := args[0]
	if abs, err := filepath.Abs(rootDir); err == nil {
		rootDir = abs
	}

	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory %s does not exist\n", rootDir)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Scanning AOSP tree at: %s\n", rootDir)

	registry := Registry{
		Version: "1.0.0",
		RootDir: rootDir,
		Modules: make(map[string]ModuleInfo),
		Defaults: make(map[string]ModuleInfo),
	}

	// Find and parse all .bp files
	bpCount := 0
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			// Skip hidden dirs and build output
			dirName := info.Name()
			if dirName == ".git" || dirName == "out" || dirName == ".repo" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode().IsRegular() {
			base := filepath.Base(path)
			if base == "Android.bp" || strings.HasSuffix(base, ".bp") {
				relPath, _ := filepath.Rel(rootDir, path)
				parseBPFile(path, relPath, &registry)
				bpCount++
			}
			// Also handle .mk files
			if base == "Android.mk" || strings.HasSuffix(base, ".mk") {
				relPath, _ := filepath.Rel(rootDir, path)
				parseMKFile(path, relPath, &registry)
				registry.TotalMK++
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error walking directory: %v\n", err)
	}

	registry.TotalBP = bpCount

	// Resolve defaults chains
	resolveDefaults(&registry)

	// Write JSON output
	output, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
	fmt.Fprintf(os.Stderr, "\nScan complete: %d .bp files, %d .mk files\n", bpCount, registry.TotalMK)
	fmt.Fprintf(os.Stderr, "Found %d modules, %d defaults\n", len(registry.Modules), len(registry.Defaults))
}

// ============================================================================
// Parse .bp file
// ============================================================================

func parseBPFile(filePath, relPath string, reg *Registry) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open %s: %v\n", filePath, err)
		return
	}
	defer f.Close()

	dir := filepath.Dir(relPath)

	file, errs := bkparser.Parse(filePath, f, bkparser.NewScope(nil))
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d parsing errors in %s\n", len(errs), relPath)
	}

	// First pass: collect variable assignments for resolution
	variables := make(map[string]bkparser.Expression)
	for _, def := range file.Defs {
		if assign, ok := def.(*bkparser.Assignment); ok {
			variables[assign.Name] = assign.Value
		}
	}

	// Second pass: parse modules
	for _, def := range file.Defs {
		switch d := def.(type) {
		case *bkparser.Module:
			moduleInfo := extractModuleInfo(d, relPath, dir, filePath, variables)

			// Check if it's a defaults module
			if isDefaultsModule(d.Type) {
				reg.Defaults[moduleInfo.Name] = moduleInfo
			} else {
				reg.Modules[moduleInfo.Name] = moduleInfo
			}
		}
	}
}

// isDefaultsModule returns true if the module type is a defaults definition.
func isDefaultsModule(typ string) bool {
	return typ == "cc_defaults" || typ == "java_defaults" || typ == "go_defaults" ||
		typ == "aidl_defaults" || typ == "proto_defaults" ||
		strings.HasSuffix(typ, "_defaults")
}

// extractModuleInfo extracts all relevant properties from a blueprint Module.
func extractModuleInfo(mod *bkparser.Module, relPath, dir, filePath string, variables map[string]bkparser.Expression) ModuleInfo {
	info := ModuleInfo{
		Type:       mod.Type,
		File:       relPath,
		Dir:        dir,
		Properties: make(map[string]interface{}),
	}

	// Get module name
	if nameProp, found := mod.GetProperty("name"); found {
		if str, ok := nameProp.Value.(*bkparser.String); ok {
			info.Name = str.Value
		}
	}

	// Extract all properties
	for _, prop := range mod.Properties {
		value := resolveExpression(prop.Value, variables)
		info.Properties[prop.Name] = expressionToJSON(value)

		// Extract known dependency properties
		switch prop.Name {
		case "name":
			if str, ok := value.(*bkparser.String); ok {
				info.Name = str.Value
			}
		case "srcs":
			info.Srcs = extractStrings(value)
		case "shared_libs":
			info.SharedLibs = extractStrings(value)
		case "static_libs":
			info.StaticLibs = extractStrings(value)
		case "header_libs":
			info.HeaderLibs = extractStrings(value)
		case "defaults":
			info.Defaults = extractStrings(value)
		case "local_include_dirs":
			info.IncludeDirs = append(info.IncludeDirs, extractStrings(value)...)
		case "export_include_dirs":
			info.IncludeDirs = append(info.IncludeDirs, extractStrings(value)...)
		case "deps":
			// General deps property (used by some module types)
			deps := extractStrings(value)
			if info.AllDeps == nil {
				info.AllDeps = make([]string, 0)
			}
			info.AllDeps = append(info.AllDeps, deps...)
		}
	}

	// Build combined deps list if not already set
	if len(info.AllDeps) == 0 {
		info.AllDeps = combineDeps(info.SharedLibs, info.StaticLibs, info.HeaderLibs)
	}

	return info
}

// resolveExpression resolves variable references in an expression.
func resolveExpression(expr bkparser.Expression, variables map[string]bkparser.Expression) bkparser.Expression {
	switch v := expr.(type) {
	case *bkparser.Variable:
		if resolved, ok := variables[v.Name]; ok {
			return resolveExpression(resolved, variables)
		}
		return expr
	case *bkparser.List:
		result := *v
		result.Values = make([]bkparser.Expression, len(v.Values))
		for i, val := range v.Values {
			result.Values[i] = resolveExpression(val, variables)
		}
		return &result
	default:
		return expr
	}
}

// extractStrings extracts all string values from an expression.
// Filters out variable substitutions like $(VAR_NAME) which are not valid file paths.
func extractStrings(expr bkparser.Expression) []string {
	if expr == nil {
		return nil
	}

	var strings []string
	switch v := expr.(type) {
	case *bkparser.String:
		if !isVariableReference(v.Value) {
			strings = append(strings, v.Value)
		}
	case *bkparser.List:
		strings = make([]string, 0, len(v.Values))
		for _, val := range v.Values {
			strings = append(strings, extractStrings(val)...)
		}
	case *bkparser.Operator:
		// Handle list concatenation
		strings = append(strings, extractStrings(v.Args[0])...)
		strings = append(strings, extractStrings(v.Args[1])...)
	}

	// Deduplicate
	return uniqueStrings(strings)
}

// isVariableReference checks if a string is a variable substitution like $(VAR)
func isVariableReference(s string) bool {
	return strings.HasPrefix(s, "$(") && strings.HasSuffix(s, ")")
}

// expandWildcard attempts to expand $(wildcard ...) patterns
func expandWildcard(pattern, rootDir string) []string {
	if !strings.Contains(pattern, "$(wildcard") {
		return nil
	}

	// Extract the glob pattern from $(wildcard ...)
	start := strings.Index(pattern, "$(wildcard")
	if start == -1 {
		return nil
	}
	end := strings.Index(pattern[start:], ")")
	if end == -1 {
		return nil
	}
	wildcardPattern := pattern[start+len("$(wildcard") : start+end]
	wildcardPattern = strings.TrimSpace(wildcardPattern)

	// Expand the glob
	fullPattern := filepath.Join(rootDir, wildcardPattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil
	}
	return matches
}

// expressionToJSON converts an expression to a JSON-serializable value.
func expressionToJSON(expr bkparser.Expression) interface{} {
	if expr == nil {
		return nil
	}

	switch v := expr.(type) {
	case *bkparser.String:
		return v.Value
	case *bkparser.Bool:
		return v.Value
	case *bkparser.Int64:
		return v.Value
	case *bkparser.List:
		result := make([]interface{}, len(v.Values))
		for i, val := range v.Values {
			result[i] = expressionToJSON(val)
		}
		return result
	case *bkparser.Map:
		result := make(map[string]interface{})
		for _, prop := range v.Properties {
			result[prop.Name] = expressionToJSON(prop.Value)
		}
		return result
	case *bkparser.Variable:
		// Return variable name as a hint
		return fmt.Sprintf("<%s>", v.Name)
	case *bkparser.Operator:
		return expressionToJSON(v.Value)
	default:
		return fmt.Sprintf("<%T>", v)
	}
}

// ============================================================================
// Parse .mk file
// ============================================================================

func parseMKFile(filePath, relPath string, reg *Registry) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open %s: %v\n", filePath, err)
		return
	}
	defer f.Close()

	dir := filepath.Dir(relPath)

	p := mkparser.NewParser(filePath, f)
	nodes, errs := p.Parse()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d parsing errors in %s\n", len(errs), relPath)
	}

	// Parse Makefile-style modules
	// Pattern: include $(CLEAR_VARS) ... LOCAL_MODULE := name ... include $(BUILD_*)
	parseMKModules(nodes, relPath, dir, reg)
}

func parseMKModules(nodes []mkparser.Node, relPath, dir string, reg *Registry) {
	type mkModule struct {
		name    string
		srcs    []string
		statics []string
		shared  []string
		includes []string
		props   map[string]string
	}

	var current *mkModule
	buildTargets := map[string]bool{
		"BUILD_STATIC_LIBRARY":   true,
		"BUILD_SHARED_LIBRARY":   true,
		"BUILD_EXECUTABLE":       true,
		"BUILD_PREBUILT":         true,
		"BUILD_PREBUILT_SHARED_LIBRARY": true,
		"JAVA_LIBRARY":           true,
		"STATIC_JAVA_LIBRARY":    true,
		"SHARED_JAVA_LIBRARY":    true,
		"BUILD_PACKAGE":          true,
	}

	for _, node := range nodes {
		switch n := node.(type) {
		case *mkparser.Directive:
			// Check for include $(CLEAR_VARS)
			if strings.Contains(n.Args.Dump(), "CLEAR_VARS") {
				current = &mkModule{
					props: make(map[string]string),
				}
			}
			// Check for include $(BUILD_*)
			if current != nil {
				dump := n.Args.Dump()
				for target := range buildTargets {
					if strings.Contains(dump, target) {
						// Finalize module
						info := ModuleInfo{
							Name:        current.name,
							Type:        target,
							File:        relPath,
							Dir:         dir,
							Srcs:        current.srcs,
							StaticLibs:  current.statics,
							SharedLibs:  current.shared,
							IncludeDirs: current.includes,
							Properties:  make(map[string]interface{}),
						}
						info.AllDeps = combineDeps(current.shared, current.statics, nil)

						if current.name != "" {
							// Copy properties
							for k, v := range current.props {
								info.Properties[k] = v
							}
							reg.Modules[current.name] = info
						}
						current = nil
						break
					}
				}
			}

		case *mkparser.Assignment:
			if current == nil {
				continue
			}
			name := n.Name.Dump()
			value := n.Value.Dump()

			current.props[name] = value

			switch name {
			case "LOCAL_MODULE":
				current.name = value
			case "LOCAL_SRC_FILES":
				current.srcs = parseMKSrcFiles(value)
			case "LOCAL_STATIC_LIBRARIES":
				current.statics = strings.Fields(value)
			case "LOCAL_SHARED_LIBRARIES":
				current.shared = strings.Fields(value)
			case "LOCAL_EXPORT_C_INCLUDE_DIRS", "LOCAL_C_INCLUDES":
				current.includes = strings.Fields(value)
			}
		}
	}
}

func parseMKSrcFiles(value string) []string {
	// Simple parsing - split by spaces, ignoring $(wildcard ...) functions
	var srcs []string
	for _, part := range strings.Fields(value) {
		if !strings.HasPrefix(part, "$(") {
			srcs = append(srcs, part)
		}
	}
	return srcs
}

// ============================================================================
// DEPS command - resolve transitive dependencies
// ============================================================================

func cmdDeps(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: aospparse deps <registry.json> <module1> [module2 ...]")
		os.Exit(1)
	}

	registry := loadRegistry(args[0])
	modules := args[1:]

	visited := make(map[string]bool)
	queue := make([]string, len(modules))
	copy(queue, modules)

	result := make([]string, 0)

	for len(queue) > 0 {
		module := queue[0]
		queue = queue[1:]

		if visited[module] {
			continue
		}
		visited[module] = true
		result = append(result, module)

		if info, ok := registry.Modules[module]; ok {
			for _, dep := range getAllDeps(info) {
				if !visited[dep] {
					queue = append(queue, dep)
				}
			}
		}
	}

	sort.Strings(result)
	for _, m := range result {
		fmt.Println(m)
	}
	fmt.Fprintf(os.Stderr, "\nTotal transitive dependencies: %d\n", len(result))
}

func getAllDeps(info ModuleInfo) []string {
	combined := combineDeps(info.SharedLibs, info.StaticLibs, info.HeaderLibs)
	combined = append(combined, info.AllDeps...)
	// Also include defaults which may add more deps
	for _, defName := range info.Defaults {
		if defInfo, ok := registryFromContext().Defaults[defName]; ok {
			combined = append(combined, getAllDeps(defInfo)...)
		}
	}
	return uniqueStrings(combined)
}

var globalRegistry *Registry

func registryFromContext() *Registry {
	return globalRegistry
}

// ============================================================================
// PATHS command - resolve source file paths
// ============================================================================

func cmdPaths(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: aospparse paths <registry.json> <module1> [module2 ...]")
		os.Exit(1)
	}

	registry := loadRegistry(args[0])
	globalRegistry = &registry
	modules := args[1:]

	// First resolve all transitive deps
	visited := make(map[string]bool)
	queue := make([]string, len(modules))
	copy(queue, modules)
	allModules := make([]string, 0)

	for len(queue) > 0 {
		module := queue[0]
		queue = queue[1:]

		if visited[module] {
			continue
		}
		visited[module] = true
		allModules = append(allModules, module)

		if info, ok := registry.Modules[module]; ok {
			for _, dep := range getAllDeps(info) {
				if !visited[dep] {
					queue = append(queue, dep)
				}
			}
		}
	}

	// Output source paths
	for _, moduleName := range allModules {
		if info, ok := registry.Modules[moduleName]; ok {
			fmt.Printf("%s (%s):\n", moduleName, info.File)
			for _, src := range info.Srcs {
				// Resolve relative path
				fullPath := filepath.Join(info.Dir, src)
				fmt.Printf("  %s\n", fullPath)
			}
		}
	}
}

// ============================================================================
// SOURCES command - resolve all sources with dedup and gather option
// ============================================================================

// SourceOutput represents the deduplicated source collection
type SourceOutput struct {
	Version    string              `json:"version"`
	RootDir    string              `json:"root_dir"`
	Modules    map[string][]string `json:"modules"` // module name -> list of source files
	AllSources []string            `json:"all_sources"` // deduplicated list of all sources
}

func cmdSources(args []string) {
	// Parse flags: --gather and --dest
	var gather bool
	var dest string

	// Filter args to separate module names from flags
	var moduleArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--gather" {
			gather = true
		} else if args[i] == "--dest" && i+1 < len(args) {
			dest = args[i+1]
			i++
		} else {
			moduleArgs = append(moduleArgs, args[i])
		}
	}

	if len(moduleArgs) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: aospparse sources <registry.json> <module1> [module2 ...] [--gather] [--dest <path>]")
		os.Exit(1)
	}

	registry := loadRegistry(moduleArgs[0])
	globalRegistry = &registry
	modules := moduleArgs[1:]

	// Collect all transitive dependencies
	visited := make(map[string]bool)
	queue := make([]string, len(modules))
	copy(queue, modules)
	allModules := make([]string, 0)

	for len(queue) > 0 {
		module := queue[0]
		queue = queue[1:]

		if visited[module] {
			continue
		}
		visited[module] = true
		allModules = append(allModules, module)

		if info, ok := registry.Modules[module]; ok {
			for _, dep := range getAllDeps(info) {
				if !visited[dep] {
					queue = append(queue, dep)
				}
			}
		}
	}

	// Build deduplicated source mapping
	moduleSources := make(map[string][]string)
	seenSources := make(map[string]bool)
	allSources := make([]string, 0)

	for _, moduleName := range allModules {
		if info, ok := registry.Modules[moduleName]; ok {
			moduleSrcs := make([]string, 0)

			// Expand glob patterns to actual files
			for _, pattern := range info.Srcs {
				pathWithDir := filepath.Join(info.Dir, pattern)
				expanded := expandGlob(pathWithDir, registry.RootDir)
				for _, file := range expanded {
					relPath, _ := filepath.Rel(registry.RootDir, file)
					if !seenSources[relPath] {
						seenSources[relPath] = true
						allSources = append(allSources, relPath)
					}
					moduleSrcs = append(moduleSrcs, relPath)
				}
			}

			moduleSources[moduleName] = moduleSrcs
		}
	}

	// Output JSON
	output := SourceOutput{
		Version:    "1.0.0",
		RootDir:    registry.RootDir,
		Modules:    moduleSources,
		AllSources: allSources,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))

	// If --gather flag is set, copy files to dest directory
	if gather {
		if dest == "" {
			fmt.Fprintln(os.Stderr, "Error: --dest is required when using --gather")
			os.Exit(1)
		}

		if err := os.MkdirAll(dest, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating dest directory: %v\n", err)
			os.Exit(1)
		}

		copied := 0
		for _, src := range allSources {
			srcPath := filepath.Join(registry.RootDir, src)
			destPath := filepath.Join(dest, src)

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot create dir for %s: %v\n", destPath, err)
				continue
			}

			if err := copyFile(srcPath, destPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot copy %s: %v\n", srcPath, err)
				continue
			}
			copied++
		}

		fmt.Fprintf(os.Stderr, "\nCopied %d files to %s\n", copied, dest)
	}

	fmt.Fprintf(os.Stderr, "\nTotal unique source files: %d\n", len(allSources))
	fmt.Fprintf(os.Stderr, "Total modules with sources: %d\n", len(moduleSources))
}

// expandGlob expands a glob pattern to actual file paths
// Handles both regular globs and $(wildcard ...) patterns
func expandGlob(pattern, rootDir string) []string {
	// First try to expand wildcard patterns
	if matches := expandWildcard(pattern, rootDir); len(matches) > 0 {
		return matches
	}

	// Then try regular glob
	fullPattern := filepath.Join(rootDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: glob error for %s: %v\\n", fullPattern, err)
		return nil
	}
	return matches
}

// copyFile copies a file from src to dest
func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// ============================================================================
// EXTRACT command - legacy single file extraction
// ============================================================================

func cmdExtract(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: aospparse extract <file> <key>")
		os.Exit(1)
	}

	filename := args[0]
	keyName := args[1]

	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)

	switch ext {
	case ".bp":
		extractBP(filename, keyName)
	case ".mk":
		extractMK(filename, keyName)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported file type: %s (use .bp or .mk)\n", ext)
		os.Exit(1)
	}
}

func extractBP(filename, keyName string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	file, errs := bkparser.Parse(filename, f, bkparser.NewScope(nil))
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "%d parsing errors\n", len(errs))
	}

	variables := make(map[string]bkparser.Expression)
	for _, def := range file.Defs {
		if assign, ok := def.(*bkparser.Assignment); ok {
			variables[assign.Name] = assign.Value
		}
	}

	result := make([]string, 0)

	for _, def := range file.Defs {
		switch d := def.(type) {
		case *bkparser.Module:
			if p, found := d.GetProperty(keyName); found {
				resolved := resolveExpression(p.Value, variables)
				result = append(result, extractStrings(resolved)...)
			}
		case *bkparser.Assignment:
			if d.Name == keyName {
				resolved := resolveExpression(d.Value, variables)
				result = append(result, extractStrings(resolved)...)
			}
		}
	}

	if len(result) > 0 {
		for _, v := range uniqueStrings(result) {
			fmt.Println(v)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Key '%s' not found in %s\n", keyName, filename)
	}
}

func extractMK(filename, keyName string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p := mkparser.NewParser(filename, f)
	nodes, errs := p.Parse()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "%d parsing errors\n", len(errs))
	}

	for _, node := range nodes {
		if assign, ok := node.(*mkparser.Assignment); ok {
			if assign.Name.Dump() == keyName {
				fmt.Println(assign.Value.Dump())
			}
		}
	}
}

// ============================================================================
// Defaults resolution
// ============================================================================

func resolveDefaults(reg *Registry) {
	globalRegistry = reg

	// For each module, resolve its defaults chain and inherit properties
	for name, info := range reg.Modules {
		if len(info.Defaults) == 0 {
			continue
		}

		resolved := resolveDefaultsChain(info.Defaults, reg, make(map[string]bool))
		for _, defName := range resolved {
			defInfo, ok := reg.Defaults[defName]
			if !ok {
				continue
			}

			// Inherit srcs
			info.Srcs = appendUnique(info.Srcs, defInfo.Srcs)
			// Inherit shared_libs
			info.SharedLibs = appendUnique(info.SharedLibs, defInfo.SharedLibs)
			// Inherit static_libs
			info.StaticLibs = appendUnique(info.StaticLibs, defInfo.StaticLibs)
			// Inherit header_libs
			info.HeaderLibs = appendUnique(info.HeaderLibs, defInfo.HeaderLibs)
			// Inherit include_dirs
			info.IncludeDirs = appendUnique(info.IncludeDirs, defInfo.IncludeDirs)
		}

		// Rebuild all_deps
		info.AllDeps = combineDeps(info.SharedLibs, info.StaticLibs, info.HeaderLibs)
		reg.Modules[name] = info
	}
}

func resolveDefaultsChain(defaults []string, reg *Registry, visited map[string]bool) []string {
	var result []string
	for _, defName := range defaults {
		if visited[defName] {
			continue
		}
		visited[defName] = true
		result = append(result, defName)

		if defInfo, ok := reg.Defaults[defName]; ok {
			result = append(result, resolveDefaultsChain(defInfo.Defaults, reg, visited)...)
		}
	}
	return uniqueStrings(result)
}

// ============================================================================
// Load registry from JSON
// ============================================================================

func loadRegistry(path string) Registry {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading registry: %v\n", err)
		os.Exit(1)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing registry JSON: %v\n", err)
		os.Exit(1)
	}

	globalRegistry = &registry
	return registry
}

// ============================================================================
// Utility functions
// ============================================================================

func combineDeps(shared, statics, headers []string) []string {
	var deps []string
	deps = append(deps, shared...)
	deps = append(deps, statics...)
	deps = append(deps, headers...)
	return uniqueStrings(deps)
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, str := range s {
		if str != "" && !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

func appendUnique(base, add []string) []string {
	return uniqueStrings(append(base, add...))
}
