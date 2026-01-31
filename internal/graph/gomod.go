package graph

import (
	"bufio"
	"strings"
)

// ParseGoMod parses a go.mod file content and returns structured information.
func ParseGoMod(content []byte) (*GoModInfo, error) {
	info := &GoModInfo{}
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	var inRequireBlock bool
	var inReplaceBlock bool
	var inExcludeBlock bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle block starts
		if line == "require (" {
			inRequireBlock = true
			continue
		}
		if line == "replace (" {
			inReplaceBlock = true
			continue
		}
		if line == "exclude (" {
			inExcludeBlock = true
			continue
		}

		// Handle block ends
		if line == ")" {
			inRequireBlock = false
			inReplaceBlock = false
			inExcludeBlock = false
			continue
		}

		// Parse module directive
		if strings.HasPrefix(line, "module ") {
			info.Module = strings.TrimPrefix(line, "module ")
			info.Module = strings.TrimSpace(info.Module)
			continue
		}

		// Parse go version directive
		if strings.HasPrefix(line, "go ") {
			info.Go = strings.TrimPrefix(line, "go ")
			info.Go = strings.TrimSpace(info.Go)
			continue
		}

		// Parse single-line require
		if strings.HasPrefix(line, "require ") && !inRequireBlock {
			mv := parseModuleVersion(strings.TrimPrefix(line, "require "))
			if mv.Path != "" {
				info.Require = append(info.Require, mv)
			}
			continue
		}

		// Parse single-line replace
		if strings.HasPrefix(line, "replace ") && !inReplaceBlock {
			mr := parseModuleReplace(strings.TrimPrefix(line, "replace "))
			if mr.Old.Path != "" {
				info.Replace = append(info.Replace, mr)
			}
			continue
		}

		// Parse single-line exclude
		if strings.HasPrefix(line, "exclude ") && !inExcludeBlock {
			mv := parseModuleVersion(strings.TrimPrefix(line, "exclude "))
			if mv.Path != "" {
				info.Exclude = append(info.Exclude, mv)
			}
			continue
		}

		// Parse block contents
		if inRequireBlock {
			mv := parseModuleVersion(line)
			if mv.Path != "" {
				info.Require = append(info.Require, mv)
			}
		}

		if inReplaceBlock {
			mr := parseModuleReplace(line)
			if mr.Old.Path != "" {
				info.Replace = append(info.Replace, mr)
			}
		}

		if inExcludeBlock {
			mv := parseModuleVersion(line)
			if mv.Path != "" {
				info.Exclude = append(info.Exclude, mv)
			}
		}
	}

	return info, scanner.Err()
}

// parseModuleVersion parses a module path and version from a line.
// Format: "github.com/example/pkg v1.2.3" or "github.com/example/pkg v1.2.3 // indirect"
func parseModuleVersion(line string) ModuleVersion {
	line = strings.TrimSpace(line)

	// Check for indirect comment
	indirect := strings.Contains(line, "// indirect")
	if indirect {
		line = strings.Split(line, "//")[0]
		line = strings.TrimSpace(line)
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ModuleVersion{}
	}

	return ModuleVersion{
		Path:     parts[0],
		Version:  parts[1],
		Indirect: indirect,
	}
}

// parseModuleReplace parses a replace directive.
// Format: "github.com/old/pkg => github.com/new/pkg v1.2.3"
// Or: "github.com/old/pkg v1.0.0 => github.com/new/pkg v1.2.3"
// Or: "github.com/old/pkg => ./local/path"
func parseModuleReplace(line string) ModuleReplace {
	line = strings.TrimSpace(line)

	parts := strings.Split(line, " => ")
	if len(parts) != 2 {
		return ModuleReplace{}
	}

	oldPart := strings.TrimSpace(parts[0])
	newPart := strings.TrimSpace(parts[1])

	// Parse old module (may or may not have version)
	oldFields := strings.Fields(oldPart)
	old := ModuleVersion{Path: oldFields[0]}
	if len(oldFields) > 1 {
		old.Version = oldFields[1]
	}

	// Parse new module (may be a path or module with version)
	newFields := strings.Fields(newPart)
	newMod := ModuleVersion{Path: newFields[0]}
	if len(newFields) > 1 {
		newMod.Version = newFields[1]
	}

	return ModuleReplace{
		Old: old,
		New: newMod,
	}
}

// DirectDependencies returns only the direct (non-indirect) dependencies.
func (g *GoModInfo) DirectDependencies() []ModuleVersion {
	var direct []ModuleVersion
	for _, req := range g.Require {
		if !req.Indirect {
			direct = append(direct, req)
		}
	}
	return direct
}

// AllDependencies returns all dependencies including indirect ones.
func (g *GoModInfo) AllDependencies() []ModuleVersion {
	return g.Require
}

// IsReplaced checks if a module path is replaced.
func (g *GoModInfo) IsReplaced(path string) bool {
	for _, r := range g.Replace {
		if r.Old.Path == path {
			return true
		}
	}
	return false
}

// GetReplacement returns the replacement for a module path, if any.
func (g *GoModInfo) GetReplacement(path string) (ModuleVersion, bool) {
	for _, r := range g.Replace {
		if r.Old.Path == path {
			return r.New, true
		}
	}
	return ModuleVersion{}, false
}

// IsLocalReplace checks if a replace directive points to a local path.
func (g *GoModInfo) IsLocalReplace(r ModuleReplace) bool {
	return strings.HasPrefix(r.New.Path, ".") || strings.HasPrefix(r.New.Path, "/")
}

// HasLocalReplaces checks if the module has any local replace directives.
func (g *GoModInfo) HasLocalReplaces() bool {
	for _, r := range g.Replace {
		if g.IsLocalReplace(r) {
			return true
		}
	}
	return false
}
