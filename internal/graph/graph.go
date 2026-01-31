package graph

import (
	"context"
	"fmt"
	"sort"
)

// Graph is the interface for dependency graph operations.
type Graph interface {
	// Build constructs the graph from the portfolio configuration.
	Build(ctx context.Context, portfolio Portfolio) error

	// AddModule adds a module to the graph.
	AddModule(module Module)

	// GetModule returns a module by ID.
	GetModule(id string) (*Module, bool)

	// Dependents returns all modules that depend on the given module.
	Dependents(moduleID string) []Module

	// Dependencies returns all modules that the given module depends on.
	Dependencies(moduleID string) []Module

	// UpgradeOrder returns modules in topological order for upgrades.
	// Only includes managed modules.
	UpgradeOrder() (*UpgradeOrder, error)

	// StaleModules finds managed modules using outdated versions of a dependency.
	StaleModules(dependency string, minVersion string) []StaleModule

	// FilterByOrg returns a new graph containing only modules from the specified org.
	FilterByOrg(org string) Graph

	// FilterByLanguage returns a new graph containing only modules of the specified language.
	FilterByLanguage(lang Language) Graph

	// AllModules returns all modules in the graph.
	AllModules() []Module

	// ManagedModules returns only managed modules (those in the portfolio).
	ManagedModules() []Module

	// Snapshot creates a point-in-time snapshot of the graph.
	Snapshot() *GraphSnapshot
}

// DependencyGraph is the default implementation of Graph.
type DependencyGraph struct {
	portfolio Portfolio
	modules   map[string]*Module  // keyed by module ID
	edges     map[string][]string // module ID -> dependency IDs
	reverse   map[string][]string // module ID -> dependent IDs
}

// NewGraph creates a new empty dependency graph.
func NewGraph() *DependencyGraph {
	return &DependencyGraph{
		modules: make(map[string]*Module),
		edges:   make(map[string][]string),
		reverse: make(map[string][]string),
	}
}

// Build constructs the graph from a portfolio configuration.
// Note: The actual fetching from GitHub is done by the GraphBuilder.
func (g *DependencyGraph) Build(ctx context.Context, portfolio Portfolio) error {
	g.portfolio = portfolio
	return nil
}

// AddModule adds a module to the graph.
func (g *DependencyGraph) AddModule(module Module) {
	g.modules[module.ID] = &module

	// Build edges
	for _, dep := range module.Dependencies {
		g.edges[module.ID] = append(g.edges[module.ID], dep.ID)
		g.reverse[dep.ID] = append(g.reverse[dep.ID], module.ID)
	}
}

// GetModule returns a module by ID.
func (g *DependencyGraph) GetModule(id string) (*Module, bool) {
	m, ok := g.modules[id]
	return m, ok
}

// Dependents returns all modules that depend on the given module.
func (g *DependencyGraph) Dependents(moduleID string) []Module {
	dependentIDs := g.reverse[moduleID]
	result := make([]Module, 0, len(dependentIDs))

	for _, id := range dependentIDs {
		if m, ok := g.modules[id]; ok {
			result = append(result, *m)
		}
	}

	return result
}

// Dependencies returns all modules that the given module depends on.
func (g *DependencyGraph) Dependencies(moduleID string) []Module {
	depIDs := g.edges[moduleID]
	result := make([]Module, 0, len(depIDs))

	for _, id := range depIDs {
		if m, ok := g.modules[id]; ok {
			result = append(result, *m)
		}
	}

	return result
}

// UpgradeOrder returns modules in topological order for upgrades.
// Uses Kahn's algorithm for topological sorting.
// Only includes managed modules.
func (g *DependencyGraph) UpgradeOrder() (*UpgradeOrder, error) {
	// Get only managed modules
	managed := g.ManagedModules()

	// Build in-degree map (only counting edges between managed modules)
	inDegree := make(map[string]int)
	managedSet := make(map[string]bool)

	for _, m := range managed {
		managedSet[m.ID] = true
		inDegree[m.ID] = 0
	}

	// Count incoming edges from managed modules
	for _, m := range managed {
		for _, dep := range m.Dependencies {
			if managedSet[dep.ID] {
				inDegree[m.ID]++
			}
		}
	}

	// Initialize queue with modules that have no managed dependencies
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	result := &UpgradeOrder{}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		// Take first element
		id := queue[0]
		queue = queue[1:]

		if visited[id] {
			continue
		}
		visited[id] = true

		if m, ok := g.modules[id]; ok {
			result.Modules = append(result.Modules, *m)
		}

		// Decrease in-degree of dependents
		for _, depID := range g.reverse[id] {
			if !managedSet[depID] {
				continue
			}
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
			}
		}

		// Re-sort queue for deterministic output
		sort.Strings(queue)
	}

	// Check for cycles
	if len(result.Modules) < len(managed) {
		// There's a cycle - find it
		var cycleModules []string
		for id := range managedSet {
			if !visited[id] {
				cycleModules = append(cycleModules, id)
			}
		}
		result.Cycles = append(result.Cycles, Cycle{Modules: cycleModules})
	}

	return result, nil
}

// StaleModules finds managed modules using outdated versions of a dependency.
func (g *DependencyGraph) StaleModules(dependency string, minVersion string) []StaleModule {
	var stale []StaleModule

	for _, m := range g.ManagedModules() {
		for _, dep := range m.Dependencies {
			_, name := ParseModuleID(dep.ID)
			if name == dependency && dep.Version < minVersion {
				stale = append(stale, StaleModule{
					Module:     m,
					Dependency: dependency,
					Current:    dep.Version,
					Latest:     minVersion,
				})
			}
		}
	}

	return stale
}

// FilterByOrg returns a new graph containing only modules from the specified org.
func (g *DependencyGraph) FilterByOrg(org string) Graph {
	filtered := NewGraph()
	filtered.portfolio = g.portfolio

	for _, m := range g.modules {
		if m.Org == org {
			filtered.AddModule(*m)
		}
	}

	return filtered
}

// FilterByLanguage returns a new graph containing only modules of the specified language.
func (g *DependencyGraph) FilterByLanguage(lang Language) Graph {
	filtered := NewGraph()
	filtered.portfolio = g.portfolio

	for _, m := range g.modules {
		if m.Language == lang {
			filtered.AddModule(*m)
		}
	}

	return filtered
}

// AllModules returns all modules in the graph.
func (g *DependencyGraph) AllModules() []Module {
	result := make([]Module, 0, len(g.modules))
	for _, m := range g.modules {
		result = append(result, *m)
	}

	// Sort by ID for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// ManagedModules returns only managed modules (those in the portfolio).
func (g *DependencyGraph) ManagedModules() []Module {
	var result []Module
	for _, m := range g.modules {
		if m.IsManaged {
			result = append(result, *m)
		}
	}

	// Sort by ID for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// Snapshot creates a point-in-time snapshot of the graph.
func (g *DependencyGraph) Snapshot() *GraphSnapshot {
	modules := make(map[string]Module)
	for id, m := range g.modules {
		modules[id] = *m
	}

	return &GraphSnapshot{
		Portfolio: g.portfolio,
		Modules:   modules,
	}
}

// Stats returns statistics about the graph.
func (g *DependencyGraph) Stats() GraphStats {
	stats := GraphStats{
		TotalModules: len(g.modules),
		ByLanguage:   make(map[Language]int),
		ByOrg:        make(map[string]int),
	}

	for _, m := range g.modules {
		if m.IsManaged {
			stats.ManagedModules++
		} else {
			stats.ExternalModules++
		}

		stats.ByLanguage[m.Language]++
		stats.ByOrg[m.Org]++
		stats.TotalEdges += len(m.Dependencies)
	}

	return stats
}

// GraphStats contains statistics about the dependency graph.
type GraphStats struct {
	TotalModules    int              `json:"totalModules"`
	ManagedModules  int              `json:"managedModules"`
	ExternalModules int              `json:"externalModules"`
	TotalEdges      int              `json:"totalEdges"`
	ByLanguage      map[Language]int `json:"byLanguage"`
	ByOrg           map[string]int   `json:"byOrg"`
}

// Validate checks the graph for issues.
func (g *DependencyGraph) Validate() []ValidationIssue {
	var issues []ValidationIssue

	// Check for missing dependencies
	for _, m := range g.modules {
		for _, dep := range m.Dependencies {
			if _, ok := g.modules[dep.ID]; !ok && dep.IsManaged {
				issues = append(issues, ValidationIssue{
					Type:    "missing_dependency",
					Module:  m.ID,
					Message: fmt.Sprintf("dependency %s is marked as managed but not in graph", dep.ID),
				})
			}
		}
	}

	// Check for cycles
	order, _ := g.UpgradeOrder()
	for _, cycle := range order.Cycles {
		issues = append(issues, ValidationIssue{
			Type:    "cycle",
			Module:  cycle.Modules[0],
			Message: fmt.Sprintf("cycle detected involving: %v", cycle.Modules),
		})
	}

	return issues
}

// ValidationIssue represents a problem found during graph validation.
type ValidationIssue struct {
	Type    string `json:"type"`
	Module  string `json:"module"`
	Message string `json:"message"`
}
