package collector

import (
	"bufio"
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plexusone/versionconductor/pkg/model"
)

// GoModAnalysis contains the results of analyzing go.mod changes in a PR.
type GoModAnalysis struct {
	// Files changed
	ChangedFiles   []string
	OnlyGoModFiles bool

	// Directive changes detected in the diff
	HasReplaceChange    bool
	HasExcludeChange    bool
	HasRetractChange    bool
	HasToolchainChange  bool
	HasGoVersionChange  bool
	HasDirectiveChanges bool

	// Dependency changes
	HasNewDirectDependency bool
	NewDirectDependencies  []string
	RemovedDependencies    []string

	// Version info
	OldGoVersion string
	NewGoVersion string
}

// goModAllowedFiles is the set of files allowed in a Go dependency PR.
var goModAllowedFiles = map[string]bool{
	"go.mod": true,
	"go.sum": true,
}

// AnalyzePRForGoMod analyzes a PR to determine if it's a safe Go dependency update.
func (c *GitHubCollector) AnalyzePRForGoMod(ctx context.Context, repo model.RepoRef, prNumber int) (*GoModAnalysis, error) {
	analysis := &GoModAnalysis{}

	// Get files changed in the PR
	files, err := c.GetPRFiles(ctx, repo, prNumber)
	if err != nil {
		return nil, err
	}

	analysis.ChangedFiles = files
	analysis.OnlyGoModFiles = areOnlyGoModFiles(files)

	// Get the diff to analyze directive changes
	diff, err := c.GetPRDiff(ctx, repo, prNumber)
	if err != nil {
		return nil, err
	}

	// Analyze the diff for go.mod changes
	analyzeGoModDiff(diff, analysis)

	return analysis, nil
}

// GetPRFiles returns the list of files changed in a PR.
func (c *GitHubCollector) GetPRFiles(ctx context.Context, repo model.RepoRef, prNumber int) ([]string, error) {
	files, err := c.client.ListPullRequestFiles(ctx, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	var allFiles []string
	for _, f := range files {
		allFiles = append(allFiles, f.Filename)
	}

	return allFiles, nil
}

// GetPRDiff returns the unified diff for a PR.
func (c *GitHubCollector) GetPRDiff(ctx context.Context, repo model.RepoRef, prNumber int) (string, error) {
	return c.client.GetPullRequestDiff(ctx, repo.Owner, repo.Name, prNumber)
}

// areOnlyGoModFiles checks if all changed files are go.mod or go.sum.
func areOnlyGoModFiles(files []string) bool {
	if len(files) == 0 {
		return false
	}

	for _, f := range files {
		base := filepath.Base(f)
		if !goModAllowedFiles[base] {
			return false
		}
	}
	return true
}

// analyzeGoModDiff analyzes a unified diff to detect go.mod directive changes.
func analyzeGoModDiff(diff string, analysis *GoModAnalysis) {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	inGoMod := false

	// Regex patterns for directive detection
	replaceRe := regexp.MustCompile(`^[+-]\s*replace\s+`)
	excludeRe := regexp.MustCompile(`^[+-]\s*exclude\s+`)
	retractRe := regexp.MustCompile(`^[+-]\s*retract\s+`)
	toolchainRe := regexp.MustCompile(`^[+-]\s*toolchain\s+`)
	goVersionRe := regexp.MustCompile(`^([+-])\s*go\s+(\d+\.\d+(?:\.\d+)?)`)
	requireRe := regexp.MustCompile(`^([+-])\s*(\S+)\s+v[\d.]+`)

	// Track require block
	inRequireBlock := false
	requireBlockStart := regexp.MustCompile(`^\+?\s*require\s*\(`)
	requireBlockEnd := regexp.MustCompile(`^\+?\s*\)`)

	existingDeps := make(map[string]bool)
	newDeps := make(map[string]bool)
	removedDeps := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Track which file we're in
		if strings.HasPrefix(line, "+++ b/") || strings.HasPrefix(line, "--- a/") {
			filename := strings.TrimPrefix(line, "+++ b/")
			filename = strings.TrimPrefix(filename, "--- a/")
			inGoMod = filepath.Base(filename) == "go.mod"
			continue
		}

		if !inGoMod {
			continue
		}

		// Track require block
		if requireBlockStart.MatchString(line) {
			inRequireBlock = true
			continue
		}
		if requireBlockEnd.MatchString(line) {
			inRequireBlock = false
			continue
		}

		// Check for directive changes
		if replaceRe.MatchString(line) {
			analysis.HasReplaceChange = true
			analysis.HasDirectiveChanges = true
		}

		if excludeRe.MatchString(line) {
			analysis.HasExcludeChange = true
			analysis.HasDirectiveChanges = true
		}

		if retractRe.MatchString(line) {
			analysis.HasRetractChange = true
			analysis.HasDirectiveChanges = true
		}

		if toolchainRe.MatchString(line) {
			analysis.HasToolchainChange = true
			analysis.HasDirectiveChanges = true
		}

		// Check for go version changes
		if matches := goVersionRe.FindStringSubmatch(line); len(matches) > 2 {
			analysis.HasGoVersionChange = true
			analysis.HasDirectiveChanges = true
			if matches[1] == "-" {
				analysis.OldGoVersion = matches[2]
			} else if matches[1] == "+" {
				analysis.NewGoVersion = matches[2]
			}
		}

		// Track dependency changes in require block or single-line require
		if inRequireBlock || strings.Contains(line, "require ") {
			if matches := requireRe.FindStringSubmatch(line); len(matches) > 2 {
				op := matches[1]
				dep := matches[2]

				if op == "-" {
					// Could be a removal or version change (old line)
					existingDeps[dep] = true
					removedDeps[dep] = true
				} else if op == "+" {
					// Could be an addition or version change (new line)
					newDeps[dep] = true
					delete(removedDeps, dep) // If we see +, it's not removed
				}
			}
		}
	}

	// Determine truly new dependencies (added, not just version bumped)
	for dep := range newDeps {
		if !existingDeps[dep] {
			analysis.HasNewDirectDependency = true
			analysis.NewDirectDependencies = append(analysis.NewDirectDependencies, dep)
		}
	}

	// Determine truly removed dependencies
	for dep := range removedDeps {
		analysis.RemovedDependencies = append(analysis.RemovedDependencies, dep)
	}
}

// ToGoModContext converts GoModAnalysis to model.GoModContext.
func (a *GoModAnalysis) ToGoModContext() model.GoModContext {
	return model.GoModContext{
		HasReplaceChange:       a.HasReplaceChange,
		HasExcludeChange:       a.HasExcludeChange,
		HasRetractChange:       a.HasRetractChange,
		HasToolchainChange:     a.HasToolchainChange,
		HasGoVersionChange:     a.HasGoVersionChange,
		HasDirectiveChanges:    a.HasDirectiveChanges,
		HasNewDirectDependency: a.HasNewDirectDependency,
		NewDirectDependencies:  a.NewDirectDependencies,
		RemovedDependencies:    a.RemovedDependencies,
		OldGoVersion:           a.OldGoVersion,
		NewGoVersion:           a.NewGoVersion,
	}
}

// ToPRContextFields returns fields to update in PRContext.
func (a *GoModAnalysis) ToPRContextFields() (changedFiles []string, onlyGoModFiles bool, changedFileExts []string) {
	changedFiles = a.ChangedFiles
	onlyGoModFiles = a.OnlyGoModFiles

	// Extract unique file extensions
	extSet := make(map[string]bool)
	for _, f := range a.ChangedFiles {
		ext := filepath.Ext(f)
		if ext != "" {
			extSet[ext] = true
		} else {
			// No extension, use filename
			extSet[filepath.Base(f)] = true
		}
	}

	for ext := range extSet {
		changedFileExts = append(changedFileExts, ext)
	}

	return
}
