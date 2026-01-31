package graph

import (
	"strings"
	"testing"
)

func TestDependencyGraph_ToDOT(t *testing.T) {
	g := NewGraph()

	// Add modules
	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		Org:       "github.com/grokify",
		IsManaged: true,
	})

	g.AddModule(Module{
		ID:        "go:github.com/grokify/gogithub",
		Language:  LanguageGo,
		Name:      "github.com/grokify/gogithub",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/grokify/mogo", Version: "v0.70.0", IsManaged: true},
		},
	})

	cfg := DefaultDOTConfig()
	dot := g.ToDOT(cfg)

	// Verify DOT format basics
	if !strings.HasPrefix(dot, "digraph dependencies {") {
		t.Error("expected DOT to start with digraph")
	}

	if !strings.Contains(dot, "go_github_com_grokify_mogo") {
		t.Error("expected mogo node in DOT output")
	}

	if !strings.Contains(dot, "go_github_com_grokify_gogithub") {
		t.Error("expected gogithub node in DOT output")
	}

	if !strings.Contains(dot, "->") {
		t.Error("expected edge in DOT output")
	}
}

func TestDependencyGraph_ToDOT_WithClustering(t *testing.T) {
	g := NewGraph()

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

	cfg := DOTConfig{
		ClusterByOrg: true,
		ColorManaged: "#4CAF50",
	}
	dot := g.ToDOT(cfg)

	// Verify clustering
	if !strings.Contains(dot, "subgraph cluster_") {
		t.Error("expected subgraph clusters in DOT output")
	}

	if !strings.Contains(dot, "github.com/grokify") {
		t.Error("expected grokify cluster label")
	}

	if !strings.Contains(dot, "github.com/agentplexus") {
		t.Error("expected agentplexus cluster label")
	}
}

func TestDependencyGraph_ToDOT_ShowExternal(t *testing.T) {
	g := NewGraph()

	g.AddModule(Module{
		ID:        "go:github.com/grokify/mycli",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mycli",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/spf13/cobra", Version: "v1.8.0", IsManaged: false},
		},
	})

	g.AddModule(Module{
		ID:        "go:github.com/spf13/cobra",
		Language:  LanguageGo,
		Name:      "github.com/spf13/cobra",
		Org:       "github.com/spf13",
		IsManaged: false,
	})

	// Without external
	cfg := DOTConfig{ShowExternal: false}
	dot := g.ToDOT(cfg)

	if strings.Contains(dot, "spf13") {
		t.Error("external module should not appear when ShowExternal is false")
	}

	// With external
	cfg.ShowExternal = true
	dot = g.ToDOT(cfg)

	if !strings.Contains(dot, "spf13") {
		t.Error("external module should appear when ShowExternal is true")
	}
}

func TestDependencyGraph_ToMermaid(t *testing.T) {
	g := NewGraph()

	g.AddModule(Module{
		ID:        "go:github.com/grokify/mogo",
		Language:  LanguageGo,
		Name:      "github.com/grokify/mogo",
		Org:       "github.com/grokify",
		IsManaged: true,
	})

	g.AddModule(Module{
		ID:        "go:github.com/grokify/gogithub",
		Language:  LanguageGo,
		Name:      "github.com/grokify/gogithub",
		Org:       "github.com/grokify",
		IsManaged: true,
		Dependencies: []ModuleRef{
			{ID: "go:github.com/grokify/mogo", Version: "v0.70.0", IsManaged: true},
		},
	})

	cfg := DefaultMermaidConfig()
	mermaid := g.ToMermaid(cfg)

	// Verify Mermaid format
	if !strings.HasPrefix(mermaid, "graph TB") {
		t.Error("expected Mermaid to start with graph TB")
	}

	if !strings.Contains(mermaid, "mogo") {
		t.Error("expected mogo node in Mermaid output")
	}

	if !strings.Contains(mermaid, "gogithub") {
		t.Error("expected gogithub node in Mermaid output")
	}

	if !strings.Contains(mermaid, "-->") {
		t.Error("expected edge in Mermaid output")
	}
}

func TestNodeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go:github.com/grokify/mogo", "go_github_com_grokify_mogo"},
		{"typescript:@agentplexus/core", "typescript__agentplexus_core"},
	}

	for _, tc := range tests {
		result := nodeID(tc.input)
		if result != tc.expected {
			t.Errorf("nodeID(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestShortModuleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com/grokify/mogo", "mogo"},
		{"github.com/spf13/cobra", "cobra"},
		{"mogo", "mogo"},
	}

	for _, tc := range tests {
		result := shortModuleName(tc.input)
		if result != tc.expected {
			t.Errorf("shortModuleName(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestEscapeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello "world"`, `hello \"world\"`},
		{"line1\nline2", `line1\nline2`},
		{"normal", "normal"},
	}

	for _, tc := range tests {
		result := escapeLabel(tc.input)
		if result != tc.expected {
			t.Errorf("escapeLabel(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}
