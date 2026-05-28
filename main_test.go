package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", ""}
	result := uniqueStrings(input)

	if len(result) != 3 {
		t.Errorf("uniqueStrings: got %d items, expected 3", len(result))
	}

	seen := make(map[string]bool)
	for _, s := range result {
		if seen[s] {
			t.Errorf("uniqueStrings: duplicate %q", s)
		}
		seen[s] = true
	}

	if !seen["a"] || !seen["b"] || !seen["c"] {
		t.Errorf("uniqueStrings: missing expected values, got %v", result)
	}
}

func TestCombineDeps(t *testing.T) {
	shared := []string{"libshared"}
	statics := []string{"libstatic"}
	headers := []string{"libheader"}

	result := combineDeps(shared, statics, headers)
	if len(result) != 3 {
		t.Errorf("combineDeps: got %d deps, expected 3", len(result))
	}
}

func TestAppendUnique(t *testing.T) {
	base := []string{"a", "b"}
	add := []string{"b", "c"}

	result := appendUnique(base, add)
	if len(result) != 3 {
		t.Errorf("appendUnique: got %d items, expected 3: %v", len(result), result)
	}
}

func TestResolveDefaultsChain(t *testing.T) {
	registry := &Registry{
		Modules:  make(map[string]ModuleInfo),
		Defaults: make(map[string]ModuleInfo),
	}

	registry.Defaults["base_defaults"] = ModuleInfo{
		Name:   "base_defaults",
		Type:   "cc_defaults",
		Srcs:   []string{"base.cpp"},
		Dir:    "base",
		File:   "base/Android.bp",
	}

	registry.Defaults["extended_defaults"] = ModuleInfo{
		Name:     "extended_defaults",
		Type:     "cc_defaults",
		Srcs:     []string{"extended.cpp"},
		Defaults: []string{"base_defaults"},
		Dir:      "extended",
		File:     "extended/Android.bp",
	}

	globalRegistry = registry

	chain := resolveDefaultsChain([]string{"extended_defaults"}, registry, make(map[string]bool))

	if len(chain) != 2 {
		t.Errorf("resolveDefaultsChain: got %d items, expected 2: %v", len(chain), chain)
	}

	foundExtended := false
	foundBase := false
	for _, name := range chain {
		if name == "extended_defaults" {
			foundExtended = true
		}
		if name == "base_defaults" {
			foundBase = true
		}
	}

	if !foundExtended {
		t.Error("resolveDefaultsChain: missing extended_defaults")
	}
	if !foundBase {
		t.Error("resolveDefaultsChain: missing inherited base_defaults")
	}
}

func TestResolveDefaultsFull(t *testing.T) {
	registry := &Registry{
		Modules:  make(map[string]ModuleInfo),
		Defaults: make(map[string]ModuleInfo),
	}

	registry.Defaults["lib_defaults"] = ModuleInfo{
		Name:       "lib_defaults",
		Type:       "cc_defaults",
		SharedLibs: []string{"libinherited"},
		StaticLibs: []string{"libstatic_inherited"},
	}

	registry.Modules["mylib"] = ModuleInfo{
		Name:        "mylib",
		Type:        "cc_library_shared",
		SharedLibs:  []string{"libown"},
		StaticLibs:  []string{"libown_static"},
		Defaults:    []string{"lib_defaults"},
		Srcs:        []string{"mylib.cpp"},
		Dir:         "mylib",
		File:        "mylib/Android.bp",
	}

	globalRegistry = registry
	resolveDefaults(registry)

	module := registry.Modules["mylib"]

	// Should have inherited deps
	hasInherited := false
	hasStaticInherited := false
	for _, dep := range module.SharedLibs {
		if dep == "libinherited" {
			hasInherited = true
		}
	}
	for _, dep := range module.StaticLibs {
		if dep == "libstatic_inherited" {
			hasStaticInherited = true
		}
	}

	if !hasInherited {
		t.Error("resolveDefaults: module should inherit shared_libs")
	}
	if !hasStaticInherited {
		t.Error("resolveDefaults: module should inherit static_libs")
	}
}

func TestIsDefaultsModule(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"cc_defaults", true},
		{"java_defaults", true},
		{"go_defaults", true},
		{"custom_defaults", true},
		{"cc_library_shared", false},
		{"cc_binary", false},
		{"java_library", false},
	}

	for _, tc := range testCases {
		result := isDefaultsModule(tc.input)
		if result != tc.expected {
			t.Errorf("isDefaultsModule(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestParseMKSrcFiles(t *testing.T) {
	src := "src/foo.cpp src/bar.cpp $(wildcard src/*.cpp) src/baz.cpp"
	result := parseMKSrcFiles(src)

	// Should have the non-wildcard entries
	foundFoo := false
	foundBar := false
	foundBaz := false
	for _, s := range result {
		if strings.HasSuffix(s, "foo.cpp") {
			foundFoo = true
		}
		if strings.HasSuffix(s, "bar.cpp") {
			foundBar = true
		}
		if strings.HasSuffix(s, "baz.cpp") {
			foundBaz = true
		}
	}

	if !foundFoo || !foundBar || !foundBaz {
		t.Errorf("parseMKSrcFiles: missing expected srcs, got %v", result)
	}
}

func TestExtractStringsNil(t *testing.T) {
	result := extractStrings(nil)
	if len(result) != 0 {
		t.Errorf("extractStrings(nil) = %v, expected empty", result)
	}
}

func TestCombineDepsEmpty(t *testing.T) {
	result := combineDeps(nil, nil, nil)
	if len(result) != 0 {
		t.Errorf("combineDeps(nil, nil, nil) = %v, expected empty", result)
	}
}

func TestUniqueStringsEmpty(t *testing.T) {
	result := uniqueStrings([]string{})
	if len(result) != 0 {
		t.Errorf("uniqueStrings(empty) = %v, expected empty", result)
	}
}

func TestUniqueStringsAllEmpty(t *testing.T) {
	result := uniqueStrings([]string{"", "", ""})
	if len(result) != 0 {
		t.Errorf("uniqueStrings(all empty) = %v, expected empty", result)
	}
}

func TestResolveDefaultsChainEmpty(t *testing.T) {
	registry := &Registry{
		Modules:  make(map[string]ModuleInfo),
		Defaults: make(map[string]ModuleInfo),
	}
	globalRegistry = registry

	chain := resolveDefaultsChain([]string{}, registry, make(map[string]bool))
	if len(chain) != 0 {
		t.Errorf("resolveDefaultsChain(empty) = %v, expected empty", chain)
	}
}

func TestResolveDefaultsCycleProtection(t *testing.T) {
	registry := &Registry{
		Modules:  make(map[string]ModuleInfo),
		Defaults: make(map[string]ModuleInfo),
	}

	registry.Defaults["a"] = ModuleInfo{
		Name:     "a",
		Type:     "cc_defaults",
		Defaults: []string{"b"},
	}
	registry.Defaults["b"] = ModuleInfo{
		Name:     "b",
		Type:     "cc_defaults",
		Defaults: []string{"a"}, // Circular reference
	}

	globalRegistry = registry

	chain := resolveDefaultsChain([]string{"a"}, registry, make(map[string]bool))

	// Should not loop forever - should return at most 2
	if len(chain) > 2 {
		t.Errorf("resolveDefaultsChain(circular) = %v, expected at most 2 to prevent infinite loop", chain)
	}
}

func TestModuleInfoPaths(t *testing.T) {
	info := ModuleInfo{
		Dir:  "hardware/interfaces/radio",
		File: "hardware/interfaces/radio/Android.bp",
		Srcs: []string{"src/Radio.cpp", "src/RadioConfig.cpp"},
	}

	for _, src := range info.Srcs {
		fullPath := filepath.Join(info.Dir, src)
		if !strings.HasPrefix(fullPath, "hardware/") {
			t.Errorf("Expected path to start with hardware/, got %s", fullPath)
		}
	}
}
