package graph

import (
	"fmt"
	"io"
	"strings"
)

// DOTConfig configures DOT output generation.
type DOTConfig struct {
	// Title is the graph title.
	Title string

	// RankDir is the direction of graph layout: "TB" (top-bottom), "LR" (left-right).
	RankDir string

	// ShowExternal includes external (non-managed) dependencies.
	ShowExternal bool

	// ShowVersions includes version labels on edges.
	ShowVersions bool

	// ClusterByOrg groups nodes by organization.
	ClusterByOrg bool

	// ColorManaged is the color for managed modules.
	ColorManaged string

	// ColorExternal is the color for external modules.
	ColorExternal string
}

// DefaultDOTConfig returns default DOT configuration.
func DefaultDOTConfig() DOTConfig {
	return DOTConfig{
		Title:         "Dependency Graph",
		RankDir:       "TB",
		ShowExternal:  false,
		ShowVersions:  true,
		ClusterByOrg:  true,
		ColorManaged:  "#4CAF50",
		ColorExternal: "#9E9E9E",
	}
}

// WriteDOT writes the graph in DOT format for Graphviz.
func (g *DependencyGraph) WriteDOT(w io.Writer, cfg DOTConfig) error {
	// Start graph
	fmt.Fprintf(w, "digraph dependencies {\n")
	fmt.Fprintf(w, "  label=\"%s\";\n", cfg.Title)
	fmt.Fprintf(w, "  labelloc=\"t\";\n")
	fmt.Fprintf(w, "  rankdir=\"%s\";\n", cfg.RankDir)
	fmt.Fprintf(w, "  node [shape=box, style=filled];\n")
	fmt.Fprintf(w, "\n")

	// Collect modules by org for clustering
	orgModules := make(map[string][]Module)
	for _, m := range g.modules {
		if !cfg.ShowExternal && !m.IsManaged {
			continue
		}
		orgModules[m.Org] = append(orgModules[m.Org], *m)
	}

	// Write nodes (optionally clustered by org)
	if cfg.ClusterByOrg {
		clusterNum := 0
		for org, modules := range orgModules {
			if org == "" {
				org = "external"
			}
			fmt.Fprintf(w, "  subgraph cluster_%d {\n", clusterNum)
			fmt.Fprintf(w, "    label=\"%s\";\n", escapeLabel(org))
			fmt.Fprintf(w, "    style=dashed;\n")

			for _, m := range modules {
				writeNode(w, m, cfg, "    ")
			}

			fmt.Fprintf(w, "  }\n\n")
			clusterNum++
		}
	} else {
		for _, m := range g.modules {
			if !cfg.ShowExternal && !m.IsManaged {
				continue
			}
			writeNode(w, *m, cfg, "  ")
		}
		fmt.Fprintf(w, "\n")
	}

	// Write edges
	for _, m := range g.modules {
		if !cfg.ShowExternal && !m.IsManaged {
			continue
		}

		for _, dep := range m.Dependencies {
			// Skip external deps if not showing them
			if !cfg.ShowExternal && !dep.IsManaged {
				continue
			}

			// Check if target exists in graph
			if _, ok := g.modules[dep.ID]; !ok && !cfg.ShowExternal {
				continue
			}

			fromID := nodeID(m.ID)
			toID := nodeID(dep.ID)

			if cfg.ShowVersions && dep.Version != "" {
				fmt.Fprintf(w, "  %s -> %s [label=\"%s\"];\n",
					fromID, toID, escapeLabel(dep.Version))
			} else {
				fmt.Fprintf(w, "  %s -> %s;\n", fromID, toID)
			}
		}
	}

	fmt.Fprintf(w, "}\n")
	return nil
}

// writeNode writes a single node definition.
func writeNode(w io.Writer, m Module, cfg DOTConfig, indent string) {
	id := nodeID(m.ID)
	label := shortModuleName(m.Name)

	color := cfg.ColorManaged
	if !m.IsManaged {
		color = cfg.ColorExternal
	}

	fmt.Fprintf(w, "%s%s [label=\"%s\", fillcolor=\"%s\"];\n",
		indent, id, escapeLabel(label), color)
}

// nodeID creates a valid DOT node ID from a module ID.
func nodeID(id string) string {
	// Replace special characters with underscores
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		".", "_",
		"-", "_",
		"@", "_",
	)
	return replacer.Replace(id)
}

// shortModuleName extracts a short display name from a module path.
func shortModuleName(name string) string {
	// For Go modules like "github.com/grokify/mogo", return "mogo"
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

// escapeLabel escapes a string for use in DOT labels.
func escapeLabel(s string) string {
	replacer := strings.NewReplacer(
		"\"", "\\\"",
		"\n", "\\n",
	)
	return replacer.Replace(s)
}

// ToDOT returns the graph as a DOT string.
func (g *DependencyGraph) ToDOT(cfg DOTConfig) string {
	var sb strings.Builder
	_ = g.WriteDOT(&sb, cfg)
	return sb.String()
}

// MermaidConfig configures Mermaid diagram output.
type MermaidConfig struct {
	// Direction is "TB", "BT", "LR", or "RL".
	Direction string

	// ShowExternal includes external dependencies.
	ShowExternal bool
}

// DefaultMermaidConfig returns default Mermaid configuration.
func DefaultMermaidConfig() MermaidConfig {
	return MermaidConfig{
		Direction:    "TB",
		ShowExternal: false,
	}
}

// WriteMermaid writes the graph in Mermaid format.
func (g *DependencyGraph) WriteMermaid(w io.Writer, cfg MermaidConfig) error {
	fmt.Fprintf(w, "graph %s\n", cfg.Direction)

	// Track written nodes to avoid duplicates
	writtenNodes := make(map[string]bool)

	// Write edges and collect nodes
	for _, m := range g.modules {
		if !cfg.ShowExternal && !m.IsManaged {
			continue
		}

		fromID := mermaidID(m.ID)
		fromLabel := shortModuleName(m.Name)

		// Write node if not written
		if !writtenNodes[fromID] {
			style := ""
			if m.IsManaged {
				style = ":::managed"
			}
			fmt.Fprintf(w, "    %s[%s]%s\n", fromID, fromLabel, style)
			writtenNodes[fromID] = true
		}

		for _, dep := range m.Dependencies {
			if !cfg.ShowExternal && !dep.IsManaged {
				continue
			}

			toID := mermaidID(dep.ID)
			toLabel := shortModuleName(extractModuleName(dep.ID))

			// Write target node if not written
			if !writtenNodes[toID] {
				style := ""
				if dep.IsManaged {
					style = ":::managed"
				}
				fmt.Fprintf(w, "    %s[%s]%s\n", toID, toLabel, style)
				writtenNodes[toID] = true
			}

			fmt.Fprintf(w, "    %s --> %s\n", fromID, toID)
		}
	}

	// Add styles
	fmt.Fprintf(w, "    classDef managed fill:#4CAF50,color:#fff\n")

	return nil
}

// mermaidID creates a valid Mermaid node ID.
func mermaidID(id string) string {
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		".", "_",
		"-", "_",
		"@", "_",
	)
	return replacer.Replace(id)
}

// extractModuleName extracts the module name from a module ID.
func extractModuleName(id string) string {
	_, name := ParseModuleID(id)
	return name
}

// ToMermaid returns the graph as a Mermaid string.
func (g *DependencyGraph) ToMermaid(cfg MermaidConfig) string {
	var sb strings.Builder
	_ = g.WriteMermaid(&sb, cfg)
	return sb.String()
}
