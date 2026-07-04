package collector

import (
	"reflect"
	"sort"
	"testing"
)

func TestAreOnlyGoModFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{
			name:  "only go.mod",
			files: []string{"go.mod"},
			want:  true,
		},
		{
			name:  "only go.sum",
			files: []string{"go.sum"},
			want:  true,
		},
		{
			name:  "go.mod and go.sum",
			files: []string{"go.mod", "go.sum"},
			want:  true,
		},
		{
			name:  "with other file",
			files: []string{"go.mod", "go.sum", "main.go"},
			want:  false,
		},
		{
			name:  "only other file",
			files: []string{"main.go"},
			want:  false,
		},
		{
			name:  "empty",
			files: []string{},
			want:  false,
		},
		{
			name:  "nested go.mod",
			files: []string{"subdir/go.mod"},
			want:  true,
		},
		{
			name:  "nested go.mod and go.sum",
			files: []string{"pkg/v2/go.mod", "pkg/v2/go.sum"},
			want:  true,
		},
		{
			name:  "vendor directory",
			files: []string{"go.mod", "vendor/modules.txt"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := areOnlyGoModFiles(tt.files)
			if got != tt.want {
				t.Errorf("areOnlyGoModFiles(%v) = %v, want %v", tt.files, got, tt.want)
			}
		})
	}
}

func TestAnalyzeGoModDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		wantFunc func(*testing.T, *GoModAnalysis)
	}{
		{
			name: "simple version bump",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -5,7 +5,7 @@ require (
-	github.com/spf13/cobra v1.7.0
+	github.com/spf13/cobra v1.8.0
 )`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if a.HasDirectiveChanges {
					t.Error("expected no directive changes")
				}
				if a.HasNewDirectDependency {
					t.Error("expected no new dependencies")
				}
			},
		},
		{
			name: "replace directive added",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -10,3 +10,5 @@ require (
 )
+
+replace github.com/old/pkg => github.com/new/pkg v1.0.0`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasReplaceChange {
					t.Error("expected replace change")
				}
				if !a.HasDirectiveChanges {
					t.Error("expected directive changes")
				}
			},
		},
		{
			name: "exclude directive added",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -10,3 +10,5 @@ require (
 )
+
+exclude github.com/bad/pkg v1.2.3`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasExcludeChange {
					t.Error("expected exclude change")
				}
				if !a.HasDirectiveChanges {
					t.Error("expected directive changes")
				}
			},
		},
		{
			name: "retract directive added",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -10,3 +10,5 @@ require (
 )
+
+retract v1.0.0`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasRetractChange {
					t.Error("expected retract change")
				}
				if !a.HasDirectiveChanges {
					t.Error("expected directive changes")
				}
			},
		},
		{
			name: "toolchain directive added",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -1,4 +1,5 @@
 module example.com/foo

 go 1.21
+toolchain go1.21.5`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasToolchainChange {
					t.Error("expected toolchain change")
				}
				if !a.HasDirectiveChanges {
					t.Error("expected directive changes")
				}
			},
		},
		{
			name: "go version change",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -1,4 +1,4 @@
 module example.com/foo

-go 1.20
+go 1.21`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasGoVersionChange {
					t.Error("expected go version change")
				}
				if a.OldGoVersion != "1.20" {
					t.Errorf("OldGoVersion = %q, want %q", a.OldGoVersion, "1.20")
				}
				if a.NewGoVersion != "1.21" {
					t.Errorf("NewGoVersion = %q, want %q", a.NewGoVersion, "1.21")
				}
				if !a.HasDirectiveChanges {
					t.Error("expected directive changes")
				}
			},
		},
		{
			name: "new dependency added",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -3,5 +3,6 @@ module example.com/foo
 go 1.21

 require (
 	github.com/spf13/cobra v1.7.0
+	github.com/spf13/viper v1.18.0
 )`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if !a.HasNewDirectDependency {
					t.Error("expected new dependency")
				}
				if len(a.NewDirectDependencies) != 1 {
					t.Errorf("NewDirectDependencies length = %d, want 1", len(a.NewDirectDependencies))
				}
				if len(a.NewDirectDependencies) > 0 && a.NewDirectDependencies[0] != "github.com/spf13/viper" {
					t.Errorf("NewDirectDependencies[0] = %q, want %q", a.NewDirectDependencies[0], "github.com/spf13/viper")
				}
			},
		},
		{
			name: "dependency removed",
			diff: `diff --git a/go.mod b/go.mod
--- a/go.mod
+++ b/go.mod
@@ -3,6 +3,5 @@ module example.com/foo
 go 1.21

 require (
 	github.com/spf13/cobra v1.7.0
-	github.com/old/dep v1.0.0
 )`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if len(a.RemovedDependencies) != 1 {
					t.Errorf("RemovedDependencies length = %d, want 1", len(a.RemovedDependencies))
				}
				if len(a.RemovedDependencies) > 0 && a.RemovedDependencies[0] != "github.com/old/dep" {
					t.Errorf("RemovedDependencies[0] = %q, want %q", a.RemovedDependencies[0], "github.com/old/dep")
				}
			},
		},
		{
			name: "non go.mod file ignored",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,4 +1,4 @@
+replace something
+exclude something
+retract something`,
			wantFunc: func(t *testing.T, a *GoModAnalysis) {
				if a.HasDirectiveChanges {
					t.Error("expected no directive changes in non-go.mod file")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &GoModAnalysis{}
			analyzeGoModDiff(tt.diff, analysis)
			tt.wantFunc(t, analysis)
		})
	}
}

func TestGoModAnalysis_ToGoModContext(t *testing.T) {
	analysis := &GoModAnalysis{
		HasReplaceChange:       true,
		HasExcludeChange:       false,
		HasRetractChange:       true,
		HasToolchainChange:     false,
		HasGoVersionChange:     true,
		HasDirectiveChanges:    true,
		HasNewDirectDependency: true,
		NewDirectDependencies:  []string{"github.com/new/pkg"},
		RemovedDependencies:    []string{"github.com/old/pkg"},
		OldGoVersion:           "1.20",
		NewGoVersion:           "1.21",
	}

	ctx := analysis.ToGoModContext()

	if ctx.HasReplaceChange != analysis.HasReplaceChange {
		t.Errorf("HasReplaceChange = %v, want %v", ctx.HasReplaceChange, analysis.HasReplaceChange)
	}
	if ctx.HasExcludeChange != analysis.HasExcludeChange {
		t.Errorf("HasExcludeChange = %v, want %v", ctx.HasExcludeChange, analysis.HasExcludeChange)
	}
	if ctx.HasRetractChange != analysis.HasRetractChange {
		t.Errorf("HasRetractChange = %v, want %v", ctx.HasRetractChange, analysis.HasRetractChange)
	}
	if ctx.HasToolchainChange != analysis.HasToolchainChange {
		t.Errorf("HasToolchainChange = %v, want %v", ctx.HasToolchainChange, analysis.HasToolchainChange)
	}
	if ctx.HasGoVersionChange != analysis.HasGoVersionChange {
		t.Errorf("HasGoVersionChange = %v, want %v", ctx.HasGoVersionChange, analysis.HasGoVersionChange)
	}
	if ctx.HasDirectiveChanges != analysis.HasDirectiveChanges {
		t.Errorf("HasDirectiveChanges = %v, want %v", ctx.HasDirectiveChanges, analysis.HasDirectiveChanges)
	}
	if ctx.HasNewDirectDependency != analysis.HasNewDirectDependency {
		t.Errorf("HasNewDirectDependency = %v, want %v", ctx.HasNewDirectDependency, analysis.HasNewDirectDependency)
	}
	if !reflect.DeepEqual(ctx.NewDirectDependencies, analysis.NewDirectDependencies) {
		t.Errorf("NewDirectDependencies = %v, want %v", ctx.NewDirectDependencies, analysis.NewDirectDependencies)
	}
	if !reflect.DeepEqual(ctx.RemovedDependencies, analysis.RemovedDependencies) {
		t.Errorf("RemovedDependencies = %v, want %v", ctx.RemovedDependencies, analysis.RemovedDependencies)
	}
	if ctx.OldGoVersion != analysis.OldGoVersion {
		t.Errorf("OldGoVersion = %q, want %q", ctx.OldGoVersion, analysis.OldGoVersion)
	}
	if ctx.NewGoVersion != analysis.NewGoVersion {
		t.Errorf("NewGoVersion = %q, want %q", ctx.NewGoVersion, analysis.NewGoVersion)
	}
}

func TestGoModAnalysis_ToPRContextFields(t *testing.T) {
	analysis := &GoModAnalysis{
		ChangedFiles:   []string{"go.mod", "go.sum"},
		OnlyGoModFiles: true,
	}

	changedFiles, onlyGoModFiles, changedFileExts := analysis.ToPRContextFields()

	if !reflect.DeepEqual(changedFiles, analysis.ChangedFiles) {
		t.Errorf("changedFiles = %v, want %v", changedFiles, analysis.ChangedFiles)
	}
	if onlyGoModFiles != analysis.OnlyGoModFiles {
		t.Errorf("onlyGoModFiles = %v, want %v", onlyGoModFiles, analysis.OnlyGoModFiles)
	}

	// Sort for comparison since map iteration is random
	sort.Strings(changedFileExts)
	expectedExts := []string{".mod", ".sum"}
	sort.Strings(expectedExts)

	if !reflect.DeepEqual(changedFileExts, expectedExts) {
		t.Errorf("changedFileExts = %v, want %v", changedFileExts, expectedExts)
	}
}

func TestGoModAnalysis_ToPRContextFields_NoExtension(t *testing.T) {
	analysis := &GoModAnalysis{
		ChangedFiles:   []string{"Makefile", "README"},
		OnlyGoModFiles: false,
	}

	_, _, changedFileExts := analysis.ToPRContextFields()

	// Files without extensions should use the filename
	sort.Strings(changedFileExts)
	expectedExts := []string{"Makefile", "README"}
	sort.Strings(expectedExts)

	if !reflect.DeepEqual(changedFileExts, expectedExts) {
		t.Errorf("changedFileExts = %v, want %v", changedFileExts, expectedExts)
	}
}

func TestGoModAnalysis_ToPRContextFields_Mixed(t *testing.T) {
	analysis := &GoModAnalysis{
		ChangedFiles:   []string{"main.go", "go.mod", "Makefile"},
		OnlyGoModFiles: false,
	}

	_, _, changedFileExts := analysis.ToPRContextFields()

	sort.Strings(changedFileExts)
	expectedExts := []string{".go", ".mod", "Makefile"}
	sort.Strings(expectedExts)

	if !reflect.DeepEqual(changedFileExts, expectedExts) {
		t.Errorf("changedFileExts = %v, want %v", changedFileExts, expectedExts)
	}
}
