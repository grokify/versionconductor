package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/versionconductor/pkg/model"
)

// Builder constructs a dependency graph from GitHub repositories.
type Builder struct {
	client    *github.Client
	portfolio Portfolio
}

// NewBuilder creates a new graph builder with GitHub authentication.
func NewBuilder(token string) *Builder {
	client := github.NewClient(nil)
	if token != "" {
		client = client.WithAuthToken(token)
	}
	return &Builder{
		client: client,
	}
}

// Build constructs a dependency graph from the portfolio configuration.
func (b *Builder) Build(ctx context.Context, portfolio Portfolio) (*DependencyGraph, error) {
	b.portfolio = portfolio
	graph := NewGraph()
	graph.portfolio = portfolio

	// Build set of managed orgs
	managedOrgs := make(map[string]bool)
	for _, org := range portfolio.Orgs {
		managedOrgs[org] = true
	}

	// Collect repos from all orgs
	for _, org := range portfolio.Orgs {
		// Extract owner from org (e.g., "github.com/grokify" -> "grokify")
		owner := extractOwner(org)
		if owner == "" {
			continue
		}

		repos, err := b.listRepos(ctx, owner)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for %s: %w", org, err)
		}

		for _, repo := range repos {
			// Check for Go modules
			if containsLanguage(portfolio.Languages, string(LanguageGo)) || len(portfolio.Languages) == 0 {
				gomod, err := b.fetchGoMod(ctx, owner, repo.GetName(), repo.GetDefaultBranch())
				if err != nil {
					// No go.mod, skip
					continue
				}

				// Parse go.mod
				modInfo, err := ParseGoMod(gomod)
				if err != nil {
					continue
				}

				// Create module
				module := b.createModule(org, repo, modInfo, managedOrgs)
				graph.AddModule(module)
			}
		}
	}

	return graph, nil
}

// listRepos lists all repositories for an owner.
func (b *Builder) listRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Type:        "owner",
	}

	for {
		repos, resp, err := b.client.Repositories.ListByUser(ctx, owner, opts)
		if err != nil {
			// Try as organization
			orgOpts := &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{PerPage: 100},
				Type:        "all",
			}
			repos, resp, err = b.client.Repositories.ListByOrg(ctx, owner, orgOpts)
			if err != nil {
				return nil, err
			}
		}

		// Filter out archived and forked repos
		for _, repo := range repos {
			if !repo.GetArchived() && !repo.GetFork() {
				allRepos = append(allRepos, repo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// fetchGoMod fetches the go.mod file from a repository.
func (b *Builder) fetchGoMod(ctx context.Context, owner, repo, branch string) ([]byte, error) {
	content, _, resp, err := b.client.Repositories.GetContents(
		ctx, owner, repo, "go.mod",
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("go.mod not found")
	}

	// Decode content using the built-in method
	decodedContent, err := content.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return []byte(decodedContent), nil
}

// createModule creates a Module from repo and go.mod info.
func (b *Builder) createModule(org string, repo *github.Repository, modInfo *GoModInfo, managedOrgs map[string]bool) Module {
	moduleName := modInfo.Module
	moduleID := NewModuleID(LanguageGo, moduleName)

	// Determine if this module is managed
	moduleOrg := ExtractOrg(LanguageGo, moduleName)
	isManaged := managedOrgs["github.com/"+extractOwner(moduleOrg)]

	// Build dependencies
	var deps []ModuleRef
	for _, req := range modInfo.DirectDependencies() {
		depOrg := ExtractOrg(LanguageGo, req.Path)
		depManaged := managedOrgs["github.com/"+extractOwner(depOrg)]

		deps = append(deps, ModuleRef{
			ID:        NewModuleID(LanguageGo, req.Path),
			Version:   req.Version,
			IsManaged: depManaged,
		})
	}

	return Module{
		ID:       moduleID,
		Language: LanguageGo,
		Name:     moduleName,
		Org:      org,
		Version:  getLatestVersion(repo),
		Repo: &model.Repo{
			Owner:         repo.GetOwner().GetLogin(),
			Name:          repo.GetName(),
			FullName:      repo.GetFullName(),
			Description:   repo.GetDescription(),
			DefaultBranch: repo.GetDefaultBranch(),
			Private:       repo.GetPrivate(),
			Archived:      repo.GetArchived(),
			Language:      repo.GetLanguage(),
			HTMLURL:       repo.GetHTMLURL(),
		},
		IsManaged:    isManaged,
		Dependencies: deps,
	}
}

// extractOwner extracts the owner from an org string.
// "github.com/grokify" -> "grokify"
// "grokify" -> "grokify"
func extractOwner(org string) string {
	if strings.HasPrefix(org, "github.com/") {
		return strings.TrimPrefix(org, "github.com/")
	}
	parts := strings.Split(org, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return org
}

// getLatestVersion gets the latest version tag from a repo.
// For now, just returns the default branch name. TODO: fetch actual tags.
func getLatestVersion(repo *github.Repository) string {
	// TODO: Fetch actual tags and find latest semver
	return repo.GetDefaultBranch()
}

// containsLanguage checks if a language is in the list.
func containsLanguage(langs []string, target string) bool {
	for _, l := range langs {
		if strings.EqualFold(l, target) {
			return true
		}
	}
	return false
}

// BuildFromSnapshot reconstructs a graph from a snapshot.
func BuildFromSnapshot(snapshot *GraphSnapshot) *DependencyGraph {
	graph := NewGraph()
	graph.portfolio = snapshot.Portfolio

	for _, m := range snapshot.Modules {
		graph.AddModule(m)
	}

	return graph
}
