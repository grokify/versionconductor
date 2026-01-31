package graph

import (
	"testing"
)

func TestParseGoMod_Simple(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require github.com/spf13/cobra v1.8.0
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if info.Module != "github.com/example/mymodule" {
		t.Errorf("expected module github.com/example/mymodule, got %s", info.Module)
	}

	if info.Go != "1.21" {
		t.Errorf("expected go version 1.21, got %s", info.Go)
	}

	if len(info.Require) != 1 {
		t.Fatalf("expected 1 require, got %d", len(info.Require))
	}

	if info.Require[0].Path != "github.com/spf13/cobra" {
		t.Errorf("expected require path github.com/spf13/cobra, got %s", info.Require[0].Path)
	}

	if info.Require[0].Version != "v1.8.0" {
		t.Errorf("expected require version v1.8.0, got %s", info.Require[0].Version)
	}
}

func TestParseGoMod_RequireBlock(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.0
	github.com/google/uuid v1.5.0 // indirect
)
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if len(info.Require) != 3 {
		t.Fatalf("expected 3 requires, got %d", len(info.Require))
	}

	// Check direct dependencies
	directDeps := info.DirectDependencies()
	if len(directDeps) != 2 {
		t.Errorf("expected 2 direct dependencies, got %d", len(directDeps))
	}

	// Check indirect marking
	for _, req := range info.Require {
		if req.Path == "github.com/google/uuid" {
			if !req.Indirect {
				t.Error("expected uuid to be marked as indirect")
			}
		}
	}
}

func TestParseGoMod_Replace(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require github.com/example/oldpkg v1.0.0

replace github.com/example/oldpkg => github.com/example/newpkg v2.0.0
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if len(info.Replace) != 1 {
		t.Fatalf("expected 1 replace, got %d", len(info.Replace))
	}

	replace := info.Replace[0]
	if replace.Old.Path != "github.com/example/oldpkg" {
		t.Errorf("expected old path github.com/example/oldpkg, got %s", replace.Old.Path)
	}

	if replace.New.Path != "github.com/example/newpkg" {
		t.Errorf("expected new path github.com/example/newpkg, got %s", replace.New.Path)
	}

	if replace.New.Version != "v2.0.0" {
		t.Errorf("expected new version v2.0.0, got %s", replace.New.Version)
	}
}

func TestParseGoMod_LocalReplace(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require github.com/example/localpkg v1.0.0

replace github.com/example/localpkg => ../localpkg
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if !info.HasLocalReplaces() {
		t.Error("expected HasLocalReplaces to return true")
	}

	if !info.IsLocalReplace(info.Replace[0]) {
		t.Error("expected first replace to be local")
	}
}

func TestParseGoMod_ReplaceBlock(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require (
	github.com/old/pkg1 v1.0.0
	github.com/old/pkg2 v1.0.0
)

replace (
	github.com/old/pkg1 => github.com/new/pkg1 v2.0.0
	github.com/old/pkg2 => ./local/pkg2
)
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if len(info.Replace) != 2 {
		t.Fatalf("expected 2 replaces, got %d", len(info.Replace))
	}

	if info.Replace[0].New.Path != "github.com/new/pkg1" {
		t.Errorf("expected first replace to github.com/new/pkg1, got %s", info.Replace[0].New.Path)
	}

	if info.Replace[1].New.Path != "./local/pkg2" {
		t.Errorf("expected second replace to ./local/pkg2, got %s", info.Replace[1].New.Path)
	}
}

func TestParseGoMod_Exclude(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require github.com/example/pkg v1.2.0

exclude github.com/example/pkg v1.0.0
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if len(info.Exclude) != 1 {
		t.Fatalf("expected 1 exclude, got %d", len(info.Exclude))
	}

	if info.Exclude[0].Path != "github.com/example/pkg" {
		t.Errorf("expected exclude path github.com/example/pkg, got %s", info.Exclude[0].Path)
	}

	if info.Exclude[0].Version != "v1.0.0" {
		t.Errorf("expected exclude version v1.0.0, got %s", info.Exclude[0].Version)
	}
}

func TestParseGoMod_ExcludeBlock(t *testing.T) {
	content := `module github.com/example/mymodule

go 1.21

require github.com/example/pkg v1.3.0

exclude (
	github.com/example/pkg v1.0.0
	github.com/example/pkg v1.1.0
)
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if len(info.Exclude) != 2 {
		t.Fatalf("expected 2 excludes, got %d", len(info.Exclude))
	}
}

func TestParseGoMod_Comments(t *testing.T) {
	content := `// This is a comment
module github.com/example/mymodule

// Another comment
go 1.21

require (
	// Comment in block
	github.com/spf13/cobra v1.8.0
)
`
	info, err := ParseGoMod([]byte(content))
	if err != nil {
		t.Fatalf("ParseGoMod failed: %v", err)
	}

	if info.Module != "github.com/example/mymodule" {
		t.Errorf("expected module github.com/example/mymodule, got %s", info.Module)
	}

	if len(info.Require) != 1 {
		t.Fatalf("expected 1 require, got %d", len(info.Require))
	}
}

func TestGoModInfo_IsReplaced(t *testing.T) {
	info := &GoModInfo{
		Replace: []ModuleReplace{
			{
				Old: ModuleVersion{Path: "github.com/old/pkg"},
				New: ModuleVersion{Path: "github.com/new/pkg", Version: "v1.0.0"},
			},
		},
	}

	if !info.IsReplaced("github.com/old/pkg") {
		t.Error("expected github.com/old/pkg to be replaced")
	}

	if info.IsReplaced("github.com/other/pkg") {
		t.Error("expected github.com/other/pkg to not be replaced")
	}
}

func TestGoModInfo_GetReplacement(t *testing.T) {
	info := &GoModInfo{
		Replace: []ModuleReplace{
			{
				Old: ModuleVersion{Path: "github.com/old/pkg"},
				New: ModuleVersion{Path: "github.com/new/pkg", Version: "v2.0.0"},
			},
		},
	}

	newMod, found := info.GetReplacement("github.com/old/pkg")
	if !found {
		t.Fatal("expected to find replacement")
	}

	if newMod.Path != "github.com/new/pkg" {
		t.Errorf("expected new path github.com/new/pkg, got %s", newMod.Path)
	}

	if newMod.Version != "v2.0.0" {
		t.Errorf("expected new version v2.0.0, got %s", newMod.Version)
	}

	_, found = info.GetReplacement("github.com/notfound/pkg")
	if found {
		t.Error("expected not to find replacement for non-existent path")
	}
}

func TestNewModuleID(t *testing.T) {
	tests := []struct {
		lang     Language
		name     string
		expected string
	}{
		{LanguageGo, "github.com/grokify/mogo", "go:github.com/grokify/mogo"},
		{LanguageTypeScript, "@agentplexus/core", "typescript:@agentplexus/core"},
		{LanguageSwift, "Example", "swift:Example"},
	}

	for _, tc := range tests {
		result := NewModuleID(tc.lang, tc.name)
		if result != tc.expected {
			t.Errorf("NewModuleID(%s, %s) = %s, expected %s", tc.lang, tc.name, result, tc.expected)
		}
	}
}

func TestParseModuleID(t *testing.T) {
	tests := []struct {
		id           string
		expectedLang Language
		expectedName string
	}{
		{"go:github.com/grokify/mogo", LanguageGo, "github.com/grokify/mogo"},
		{"typescript:@agentplexus/core", LanguageTypeScript, "@agentplexus/core"},
		{"swift:Example", LanguageSwift, "Example"},
		{"nocolon", "", "nocolon"},
	}

	for _, tc := range tests {
		lang, name := ParseModuleID(tc.id)
		if lang != tc.expectedLang {
			t.Errorf("ParseModuleID(%s) lang = %s, expected %s", tc.id, lang, tc.expectedLang)
		}
		if name != tc.expectedName {
			t.Errorf("ParseModuleID(%s) name = %s, expected %s", tc.id, name, tc.expectedName)
		}
	}
}

func TestExtractOrg(t *testing.T) {
	tests := []struct {
		lang     Language
		name     string
		expected string
	}{
		{LanguageGo, "github.com/grokify/mogo", "github.com/grokify"},
		{LanguageGo, "github.com/spf13/cobra", "github.com/spf13"},
		{LanguageGo, "golang.org/x/oauth2", "golang.org/x"},
		{LanguageTypeScript, "@agentplexus/core", "@agentplexus"},
		{LanguageTypeScript, "lodash", ""},
	}

	for _, tc := range tests {
		result := ExtractOrg(tc.lang, tc.name)
		if result != tc.expected {
			t.Errorf("ExtractOrg(%s, %s) = %s, expected %s", tc.lang, tc.name, result, tc.expected)
		}
	}
}

func TestLanguage_ManifestFile(t *testing.T) {
	tests := []struct {
		lang     Language
		expected string
	}{
		{LanguageGo, "go.mod"},
		{LanguageTypeScript, "package.json"},
		{LanguageSwift, "Package.swift"},
		{LanguagePython, "pyproject.toml"},
		{LanguageRust, "Cargo.toml"},
		{Language("unknown"), ""},
	}

	for _, tc := range tests {
		result := tc.lang.ManifestFile()
		if result != tc.expected {
			t.Errorf("Language(%s).ManifestFile() = %s, expected %s", tc.lang, result, tc.expected)
		}
	}
}
