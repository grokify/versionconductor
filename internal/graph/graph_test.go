package graph

import (
	"testing"
)

func TestDependencyGraph_AddModule(t *testing.T) {
	g := NewGraph()

	module := Module{
		ID:        "go:github.com/example/pkg",
		Language:  LanguageGo,
		Name:      "github.com/example/pkg",
		Org:       "github.com/example",
		Version:   "v1.0.0",
		IsManaged: true,
	}

	g.AddModule(module)

	got, ok := g.GetModule("go:github.com/example/pkg")
	if !ok {
		t.Fatal("expected to find module after adding")
	}

	if got.Name != module.Name {
		t.Errorf("expected name %s, got %s", module.Name, got.Name)
	}
}

func TestDependencyGraph_Dependents(t *testing.T) {
	g := NewGraph()

	// Add base module
	base := Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		Org:       "github.com/grokify",
		IsManaged: true,
	}
	g.AddModule(base)

	// Add dependent module
	dependent := Module{
		ID:        "go:github.com/grokify/gogithub",
		Language:  LanguageGo,
		Name:      "github.com/grokify/gogithub",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/grokify/mogo", Version: "v0.70.0", IsManaged: true},
		},
	}
	g.AddModule(dependent)

	dependents := g.Dependents("go:github.com/grokify/mogo")
	if len(dependents) != 1 {
		t.Fatalf("expected 1 dependent, got %d", len(dependents))
	}

	if dependents[0].ID != "go:github.com/grokify/gogithub" {
		t.Errorf("expected dependent go:github.com/grokify/gogithub, got %s", dependents[0].ID)
	}
}

func TestDependencyGraph_Dependencies(t *testing.T) {
	g := NewGraph()

	// Add dependency
	dep := Module{
		ID:        "go:github.com/spf13/cobra",
		Language:  LanguageGo,
		Name:      "github.com/spf13/cobra",
		Org:       "github.com/spf13",
		IsManaged: false,
	}
	g.AddModule(dep)

	// Add module with dependencies
	module := Module{
		ID:        "go:github.com/grokify/mycli",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mycli",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/spf13/cobra", Version: "v1.8.0", IsManaged: false},
		},
	}
	g.AddModule(module)

	deps := g.Dependencies("go:github.com/grokify/mycli")
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(deps))
	}

	if deps[0].ID != "go:github.com/spf13/cobra" {
		t.Errorf("expected dependency go:github.com/spf13/cobra, got %s", deps[0].ID)
	}
}

func TestDependencyGraph_UpgradeOrder_Linear(t *testing.T) {
	g := NewGraph()

	// Create a linear dependency chain: A -> B -> C
	// C should be upgraded first, then B, then A

	moduleC := Module{
		ID:        "go:github.com/example/c",
		Language:  LanguageGo,
		Name:      "github.com/example/c",
		IsManaged: true,
	}
	g.AddModule(moduleC)

	moduleB := Module{
		ID:        "go:github.com/example/b",
		Language:  LanguageGo,
		Name:      "github.com/example/b",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/example/c", IsManaged: true},
		},
	}
	g.AddModule(moduleB)

	moduleA := Module{
		ID:        "go:github.com/example/a",
		Language:  LanguageGo,
		Name:      "github.com/example/a",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/example/b", IsManaged: true},
		},
	}
	g.AddModule(moduleA)

	order, err := g.UpgradeOrder()
	if err != nil {
		t.Fatalf("UpgradeOrder failed: %v", err)
	}

	if len(order.Modules) != 3 {
		t.Fatalf("expected 3 modules in order, got %d", len(order.Modules))
	}

	// C should be first (no dependencies)
	if order.Modules[0].ID != "go:github.com/example/c" {
		t.Errorf("expected first module to be C, got %s", order.Modules[0].ID)
	}

	// B should be second
	if order.Modules[1].ID != "go:github.com/example/b" {
		t.Errorf("expected second module to be B, got %s", order.Modules[1].ID)
	}

	// A should be last
	if order.Modules[2].ID != "go:github.com/example/a" {
		t.Errorf("expected third module to be A, got %s", order.Modules[2].ID)
	}
}

func TestDependencyGraph_UpgradeOrder_Diamond(t *testing.T) {
	g := NewGraph()

	// Create a diamond dependency:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	// D should be first, then B and C (in some order), then A

	moduleD := Module{
		ID:        "go:github.com/example/d",
		Language:  LanguageGo,
		Name:      "github.com/example/d",
		IsManaged: true,
	}
	g.AddModule(moduleD)

	moduleB := Module{
		ID:        "go:github.com/example/b",
		Language:  LanguageGo,
		Name:      "github.com/example/b",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/example/d", IsManaged: true},
		},
	}
	g.AddModule(moduleB)

	moduleC := Module{
		ID:        "go:github.com/example/c",
		Language:  LanguageGo,
		Name:      "github.com/example/c",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/example/d", IsManaged: true},
		},
	}
	g.AddModule(moduleC)

	moduleA := Module{
		ID:        "go:github.com/example/a",
		Language:  LanguageGo,
		Name:      "github.com/example/a",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/example/b", IsManaged: true},
			{ID: "go:github.com/example/c", IsManaged: true},
		},
	}
	g.AddModule(moduleA)

	order, err := g.UpgradeOrder()
	if err != nil {
		t.Fatalf("UpgradeOrder failed: %v", err)
	}

	if len(order.Modules) != 4 {
		t.Fatalf("expected 4 modules in order, got %d", len(order.Modules))
	}

	// D should be first
	if order.Modules[0].ID != "go:github.com/example/d" {
		t.Errorf("expected first module to be D, got %s", order.Modules[0].ID)
	}

	// A should be last
	if order.Modules[3].ID != "go:github.com/example/a" {
		t.Errorf("expected last module to be A, got %s", order.Modules[3].ID)
	}
}

func TestDependencyGraph_UpgradeOrder_IgnoresExternal(t *testing.T) {
	g := NewGraph()

	// External dependency (not managed)
	external := Module{
		ID:        "go:github.com/spf13/cobra",
		Language:  LanguageGo,
		Name:      "github.com/spf13/cobra",
		IsManaged: false,
	}
	g.AddModule(external)

	// Managed module depends on external
	managed := Module{
		ID:        "go:github.com/example/cli",
		Language:  LanguageGo,
		Name:      "github.com/example/cli",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/spf13/cobra", IsManaged: false},
		},
	}
	g.AddModule(managed)

	order, err := g.UpgradeOrder()
	if err != nil {
		t.Fatalf("UpgradeOrder failed: %v", err)
	}

	// Should only include managed modules
	if len(order.Modules) != 1 {
		t.Fatalf("expected 1 managed module in order, got %d", len(order.Modules))
	}

	if order.Modules[0].ID != "go:github.com/example/cli" {
		t.Errorf("expected managed module, got %s", order.Modules[0].ID)
	}
}

func TestDependencyGraph_FilterByOrg(t *testing.T) {
	g := NewGraph()

	// Add modules from different orgs
	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		Org:       "github.com/grokify",
		IsManaged: true,
	})

	g.AddModule(Module{
		ID:        "go:github.com/agentplexus/core",
		Language:  LanguageGo,
		Name:      "github.com/agentplexus/core",
		Org:       "github.com/agentplexus",
		IsManaged: true,
	})

	filtered := g.FilterByOrg("github.com/grokify")
	modules := filtered.AllModules()

	if len(modules) != 1 {
		t.Fatalf("expected 1 module after filtering, got %d", len(modules))
	}

	if modules[0].Org != "github.com/grokify" {
		t.Errorf("expected org github.com/grokify, got %s", modules[0].Org)
	}
}

func TestDependencyGraph_FilterByLanguage(t *testing.T) {
	g := NewGraph()

	// Add modules of different languages
	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		IsManaged: true,
	})

	g.AddModule(Module{
		ID:        "typescript:@agentplexus/web",
		Language:  LanguageTypeScript,
		Name:      "@agentplexus/web",
		IsManaged: true,
	})

	filtered := g.FilterByLanguage(LanguageGo)
	modules := filtered.AllModules()

	if len(modules) != 1 {
		t.Fatalf("expected 1 module after filtering, got %d", len(modules))
	}

	if modules[0].Language != LanguageGo {
		t.Errorf("expected language Go, got %s", modules[0].Language)
	}
}

func TestDependencyGraph_ManagedModules(t *testing.T) {
	g := NewGraph()

	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		IsManaged: true,
	})

	g.AddModule(Module{
		ID:        "go:github.com/spf13/cobra",
		Language:  LanguageGo,
		Name:      "github.com/spf13/cobra",
		IsManaged: false,
	})

	managed := g.ManagedModules()

	if len(managed) != 1 {
		t.Fatalf("expected 1 managed module, got %d", len(managed))
	}

	if !managed[0].IsManaged {
		t.Error("expected module to be marked as managed")
	}
}

func TestDependencyGraph_Stats(t *testing.T) {
	g := NewGraph()

	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/spf13/cobra"},
		},
	})

	g.AddModule(Module{
		ID:        "go:github.com/spf13/cobra",
		Language:  LanguageGo,
		Name:      "github.com/spf13/cobra",
		Org:       "github.com/spf13",
		IsManaged: false,
	})

	stats := g.Stats()

	if stats.TotalModules != 2 {
		t.Errorf("expected 2 total modules, got %d", stats.TotalModules)
	}

	if stats.ManagedModules != 1 {
		t.Errorf("expected 1 managed module, got %d", stats.ManagedModules)
	}

	if stats.ExternalModules != 1 {
		t.Errorf("expected 1 external module, got %d", stats.ExternalModules)
	}

	if stats.TotalEdges != 1 {
		t.Errorf("expected 1 edge, got %d", stats.TotalEdges)
	}

	if stats.ByLanguage[LanguageGo] != 2 {
		t.Errorf("expected 2 Go modules, got %d", stats.ByLanguage[LanguageGo])
	}
}

func TestDependencyGraph_Snapshot(t *testing.T) {
	g := NewGraph()
	g.portfolio = Portfolio{
		Name: "test-portfolio",
		Orgs: []string{"github.com/grokify"},
	}

	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		IsManaged: true,
	})

	snapshot := g.Snapshot()

	if snapshot.Portfolio.Name != "test-portfolio" {
		t.Errorf("expected portfolio name test-portfolio, got %s", snapshot.Portfolio.Name)
	}

	if len(snapshot.Modules) != 1 {
		t.Errorf("expected 1 module in snapshot, got %d", len(snapshot.Modules))
	}

	if _, ok := snapshot.Modules["go:github.com/grokify/mogo"]; !ok {
		t.Error("expected to find mogo in snapshot")
	}
}

func TestDependencyGraph_StaleModules(t *testing.T) {
	g := NewGraph()

	// Add external dependency
	g.AddModule(Module{
		ID:        "go:github.com/grokify/gogithub",
		Language:  LanguageGo,
		Name:      "github.com/grokify/gogithub",
		IsManaged: false,
	})

	// Add managed module using old version
	g.AddModule(Module{
		ID:        "go:github.com/grokify/mycli",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mycli",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/grokify/gogithub", Version: "v0.5.0", IsManaged: false},
		},
	})

	// Add managed module using current version
	g.AddModule(Module{
		ID:        "go:github.com/grokify/mylib",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mylib",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/grokify/gogithub", Version: "v0.7.0", IsManaged: false},
		},
	})

	stale := g.StaleModules("github.com/grokify/gogithub", "v0.7.0")

	if len(stale) != 1 {
		t.Fatalf("expected 1 stale module, got %d", len(stale))
	}

	if stale[0].Module.Name != "github.com/grokify/mycli" {
		t.Errorf("expected stale module github.com/grokify/mycli, got %s", stale[0].Module.Name)
	}

	if stale[0].Current != "v0.5.0" {
		t.Errorf("expected current version v0.5.0, got %s", stale[0].Current)
	}
}
