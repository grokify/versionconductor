# VersionConductor - Technical Requirements Document

## Overview

VersionConductor is a Go CLI application for automated dependency PR management and maintenance releases. It uses the GitHub API to discover and manage PRs, Cedar for policy evaluation, and follows the same architectural patterns as PipelineConductor.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      VersionConductor CLI                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  Collector   │  │    Policy    │  │      Merger          │  │
│  │ - PR Scanner │  │    Engine    │  │ - Squash/Merge/Rebase│  │
│  │ - Metadata   │  │ - Cedar      │  │ - Branch Protection  │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Releaser   │  │    Graph     │  │      Report          │  │
│  │ - Semver     │  │ - Dep Graph  │  │ - JSON               │  │
│  │ - Tags       │  │ - Topo Sort  │  │ - Markdown           │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                   │
│                    ┌──────────────────────┐                      │
│                    │      pkg/model       │                      │
│                    │ - PullRequest        │                      │
│                    │ - Release, Graph     │                      │
│                    └──────────────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
versionconductor/
├── cmd/versionconductor/
│   ├── main.go
│   └── cmd/
│       ├── root.go           # Cobra root + Viper config
│       ├── scan.go           # List dependency PRs (read-only)
│       ├── review.go         # Auto-review PRs based on policy
│       ├── merge.go          # Auto-merge approved PRs
│       ├── release.go        # Create maintenance releases
│       ├── graph.go          # Dependency graph commands (v0.2.0)
│       └── version.go        # CLI version info
├── internal/
│   ├── collector/
│   │   ├── collector.go      # Interface definition
│   │   └── github.go         # GitHub PR collection
│   ├── graph/                # Dependency graph (v0.2.0)
│   │   ├── graph.go          # Graph data structure
│   │   ├── builder.go        # Build graph from go.mod files
│   │   ├── walker.go         # Walk up/down the graph
│   │   └── topo.go           # Topological sort for upgrade order
│   ├── policy/
│   │   ├── engine.go         # Cedar policy evaluation
│   │   ├── context.go        # PR context builder
│   │   ├── profiles.go       # Merge profiles
│   │   └── loader.go         # Policy loader
│   ├── merger/
│   │   ├── merger.go         # PR merge orchestration
│   │   └── github.go         # GitHub merge implementation
│   ├── releaser/
│   │   ├── releaser.go       # Release creation orchestration
│   │   ├── semver.go         # Version bumping logic
│   │   └── github.go         # GitHub release/tag creation
│   └── report/
│       ├── report.go         # Formatter interface
│       ├── json.go
│       └── markdown.go
├── pkg/model/
│   ├── repo.go               # Repository model
│   ├── pullrequest.go        # PR model with dependency info
│   ├── release.go            # Release/tag model
│   ├── result.go             # Scan/merge results
│   ├── graph.go              # Graph node and edge models
│   └── policy.go             # Policy context for Cedar
├── policies/
│   └── examples/
│       └── auto-merge-patch.cedar
├── configs/
│   └── profiles/
├── go.mod
└── go.sum
```

## Core Components

### 1. Collector (`internal/collector`)

Discovers dependency PRs across GitHub organizations.

```go
type Collector interface {
    ListRepos(ctx context.Context, orgs []string, filter RepoFilter) ([]Repo, error)
    ListDependencyPRs(ctx context.Context, repo Repo) ([]PullRequest, error)
    GetPRDetails(ctx context.Context, repo Repo, prNumber int) (*PullRequest, error)
    GetPRChecks(ctx context.Context, repo Repo, prNumber int) ([]CheckRun, error)
}
```

**Dependency Bot Detection:**

| Bot | Author Patterns |
|-----|-----------------|
| Renovate | `renovate[bot]`, `renovate-bot` |
| Dependabot | `dependabot[bot]` |

**PR Title Parsing:**

Extract dependency info from PR titles:
- `chore(deps): update golang.org/x/oauth2 to v0.15.0`
- `Bump github.com/spf13/cobra from 1.8.0 to 1.9.0`

### 2. Policy Engine (`internal/policy`)

Evaluates Cedar policies to determine merge eligibility.

```go
type Engine interface {
    Evaluate(ctx PolicyContext, action string) (bool, error)
    LoadPolicy(policy string) error
}

type PolicyContext struct {
    Repo         RepoContext       `json:"repo"`
    PR           PRContext         `json:"pr"`
    Dependency   DependencyContext `json:"dependency"`
    CI           CIContext         `json:"ci"`
}
```

**Cedar Actions:**

| Action | Description |
|--------|-------------|
| `Action::"review"` | Can this PR be auto-reviewed? |
| `Action::"merge"` | Can this PR be auto-merged? |
| `Action::"release"` | Can a release be created? |

**Example Policy:**

```cedar
// Auto-merge patch updates when tests pass and PR is 24h old
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.isDependencyPR == true &&
    context.testsPassed == true &&
    context.prAgeHours >= 24 &&
    context.updateType == "patch"
};
```

### 3. Merger (`internal/merger`)

Orchestrates PR review and merge operations.

```go
type Merger interface {
    Review(ctx context.Context, repo Repo, pr PullRequest) error
    Merge(ctx context.Context, repo Repo, pr PullRequest, strategy MergeStrategy) error
}

type MergeStrategy string

const (
    MergeStrategySquash MergeStrategy = "squash"
    MergeStrategyMerge  MergeStrategy = "merge"
    MergeStrategyRebase MergeStrategy = "rebase"
)
```

**Safety Checks:**

1. Verify PR is still open
2. Check branch protection requirements
3. Verify CI status is passing
4. Confirm PR is mergeable (no conflicts)

### 4. Releaser (`internal/releaser`)

Creates maintenance releases with semantic version bumping.

```go
type Releaser interface {
    GetLatestRelease(ctx context.Context, repo Repo) (*Release, error)
    BumpVersion(current string, bumpType BumpType) (string, error)
    CreateRelease(ctx context.Context, repo Repo, version string, changelog string) error
}

type BumpType string

const (
    BumpPatch BumpType = "patch"  // v1.2.3 -> v1.2.4
    BumpMinor BumpType = "minor"  // v1.2.3 -> v1.3.0
    BumpMajor BumpType = "major"  // v1.2.3 -> v2.0.0
)
```

### 5. Dependency Graph (`internal/graph`) - v0.2.0

Builds and analyzes dependency relationships across repos in multiple GitHub orgs.

```go
type Graph interface {
    // Build graph from repos
    Build(ctx context.Context, repos []Repo) error

    // Get all modules that depend on the given module
    Dependents(module string) []Node

    // Get all modules that the given module depends on
    Dependencies(module string) []Node

    // Get upgrade order (topological sort)
    UpgradeOrder() []Node

    // Find modules using outdated versions of a dependency
    StaleModules(dependency string, minVersion string) []Node
}

type Node struct {
    Module      string            `json:"module"`      // e.g., github.com/grokify/mogo
    Repo        Repo              `json:"repo"`        // GitHub repo info
    Version     string            `json:"version"`     // Current version tag
    GoMod       GoModInfo         `json:"goMod"`       // Parsed go.mod
    Dependents  []string          `json:"dependents"`  // Modules that depend on this
    Dependencies []string         `json:"dependencies"` // Modules this depends on
}

type GoModInfo struct {
    Module   string              `json:"module"`
    Go       string              `json:"go"`
    Require  []ModuleVersion     `json:"require"`
    Replace  []ModuleReplace     `json:"replace"`
}
```

**Key Operations:**

| Operation | Description | Use Case |
|-----------|-------------|----------|
| `Build` | Parse go.mod files, build graph | Initial setup |
| `Dependents` | Walk up graph from a module | "What uses mogo?" |
| `Dependencies` | Walk down graph from a module | "What does pipelineconductor need?" |
| `UpgradeOrder` | Topological sort | "What order to upgrade 50 repos?" |
| `StaleModules` | Find outdated dependents | "Who's still on gogithub v0.6.0?" |

**Example Workflow:**

```bash
# Build dependency graph across orgs
versionconductor graph build --orgs grokify,mycompany

# Show what depends on mogo
versionconductor graph dependents github.com/grokify/mogo

# Show upgrade order for all repos
versionconductor graph order --orgs grokify

# Find repos using old version of gogithub
versionconductor graph stale github.com/grokify/gogithub --min-version v0.7.0
```

**Graph Building Process:**

1. List all repos across specified orgs
2. Fetch `go.mod` from each repo's default branch
3. Parse `go.mod` to extract module path and dependencies
4. Build directed graph: edge from A→B means A depends on B
5. Detect cycles (Go modules shouldn't have cycles, but report if found)
6. Cache graph for subsequent operations

**Cascade Update Detection:**

When module X releases a new version:
1. Find all dependents of X using `Dependents(X)`
2. For each dependent, check if it has a PR updating X
3. If no PR exists, flag as "needs update"
4. After dependent Y is updated and released, recursively check Y's dependents

**Scheduled Graph Upgrade Orchestration:**

For organizations wanting to upgrade their entire dependency graph on a schedule:

```go
type UpgradeOrchestrator interface {
    // Run full graph upgrade in topological order
    UpgradeGraph(ctx context.Context, opts UpgradeOptions) (*UpgradeReport, error)
}

type UpgradeOptions struct {
    Orgs           []string      // GitHub orgs to include
    Schedule       string        // Cron expression (e.g., "0 2 * * TUE")
    DryRun         bool          // Preview without executing
    CreateReleases bool          // Create patch releases after merge
    WaitForCI      bool          // Wait for CI to pass before releasing
    Timeout        time.Duration // Max time per module
}

type UpgradeReport struct {
    StartTime       time.Time
    EndTime         time.Time
    ModulesUpgraded []ModuleUpgrade
    ReleasesCreated []Release
    Failures        []UpgradeFailure
}
```

**Upgrade Algorithm:**

```
1. Build dependency graph G for all repos in specified orgs
2. Compute topological order T = toposort(G)
3. For each module M in T (bottom-up):
   a. List open dependency PRs for M
   b. For each PR that passes policy:
      - Merge PR
      - Wait for CI
   c. If any PRs were merged and CI passes:
      - Bump patch version
      - Create GitHub release
      - Log release to report
   d. Wait for Renovate/Dependabot to detect new release
      (or trigger via API if supported)
4. Return UpgradeReport with summary
```

**CLI Commands for Orchestration:**

```bash
# Preview full graph upgrade (dry-run)
versionconductor upgrade --orgs grokify --dry-run

# Execute full graph upgrade with releases
versionconductor upgrade --orgs grokify --create-releases

# Schedule recurring upgrades (outputs cron job config)
versionconductor upgrade schedule --orgs grokify --cron "0 2 * * TUE"

# Check status of last upgrade run
versionconductor upgrade status --orgs grokify
```

### 6. Data Models (`pkg/model`)

```go
// PullRequest represents a dependency update PR
type PullRequest struct {
    Number        int           `json:"number"`
    Title         string        `json:"title"`
    State         string        `json:"state"`
    Author        string        `json:"author"`
    IsDependency  bool          `json:"isDependency"`
    DependencyBot string        `json:"dependencyBot"`
    Dependency    Dependency    `json:"dependency"`
    Checks        []CheckRun    `json:"checks"`
    Mergeable     bool          `json:"mergeable"`
    CreatedAt     time.Time     `json:"createdAt"`
    AgeHours      int           `json:"ageHours"`
}

type Dependency struct {
    Name        string `json:"name"`
    FromVersion string `json:"fromVersion"`
    ToVersion   string `json:"toVersion"`
    UpdateType  string `json:"updateType"`  // major, minor, patch
    Ecosystem   string `json:"ecosystem"`   // go, npm, pip, etc.
}
```

## CLI Commands

### scan

List dependency PRs across organizations (read-only).

```bash
versionconductor scan --orgs myorg --format json
versionconductor scan --orgs org1,org2 --bot renovate
versionconductor scan --orgs myorg --update-type patch,minor
```

### review

Auto-review PRs based on Cedar policies.

```bash
versionconductor review --orgs myorg --dry-run
versionconductor review --orgs myorg --profile balanced
```

### merge

Merge approved, passing PRs.

```bash
versionconductor merge --orgs myorg --dry-run
versionconductor merge --orgs myorg --strategy squash
```

### release

Create maintenance releases for repos with merged dependency PRs.

```bash
versionconductor release --orgs myorg --dry-run
versionconductor release --orgs myorg --since 2026-01-01
```

### graph (v0.2.0)

Build and analyze dependency graph across organizations.

```bash
# Build dependency graph
versionconductor graph build --orgs grokify,mycompany

# Show all modules that depend on a given module
versionconductor graph dependents github.com/grokify/mogo

# Show all dependencies of a module
versionconductor graph dependencies github.com/grokify/pipelineconductor

# Show correct upgrade order (topological sort)
versionconductor graph order --orgs grokify

# Find modules using outdated version of a dependency
versionconductor graph stale github.com/grokify/gogithub --min-version v0.7.0

# Visualize graph (outputs DOT format for Graphviz)
versionconductor graph visualize --orgs grokify --format dot > deps.dot
```

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub personal access token |
| `VERSIONCONDUCTOR_ORGS` | Default organizations to scan |
| `VERSIONCONDUCTOR_PROFILE` | Default merge profile |

### Configuration File

```yaml
# ~/.versionconductor.yaml
github_token: ${GITHUB_TOKEN}
orgs:
  - myorg
  - otherorg
profile: balanced
merge_strategy: squash
dry_run: false
```

## Dependencies

| Module | Version | Purpose |
|--------|---------|---------|
| `github.com/google/go-github/v82` | v82.0.0 | GitHub API client |
| `github.com/grokify/gogithub` | v0.7.0 | GitHub API helpers (auth, PR, release, tag) |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `github.com/spf13/viper` | v1.21.0 | Configuration management |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing |

## Safety Features

### 1. Dry-Run Default

All write operations (review, merge, release) default to dry-run mode. Use `--execute` flag to perform actual operations.

### 2. Policy Required

No auto-merge without explicit policy configuration. Built-in profiles provide sensible defaults.

### 3. Audit Logging

All operations are logged with:
- Timestamp
- Repository
- PR number
- Action taken
- Policy evaluation result

### 4. Rate Limiting

Respects GitHub API rate limits:
- Reads `X-RateLimit-*` headers
- Implements exponential backoff
- Pauses when approaching limits

### 5. Branch Protection Compliance

Respects repository branch protection rules:
- Required reviews
- Required status checks
- Admin enforcement

## Comparison with PipelineConductor

| Aspect | PipelineConductor | VersionConductor |
|--------|-------------------|------------------|
| Primary action | Read/Report | Read + Write (merge, release) |
| Focus | CI/CD compliance | Dependency lifecycle |
| Policies evaluate | Workflows, branch protection | PRs, test status, versions |
| Output actions | Compliance report | Merges, releases |
| Risk level | Low (read-only) | Higher (mutates repos) |
| Shared code | gogithub | gogithub |

## Testing Strategy

### Unit Tests

- Policy evaluation logic
- Semver parsing and bumping
- PR title parsing
- Report generation
- Graph building and traversal (v0.2.0)
- Topological sort (v0.2.0)
- go.mod parsing (v0.2.0)

### Integration Tests

- GitHub API interactions (with mocks)
- End-to-end command execution

### Manual Testing

- Real GitHub organization with test repos
- Verify dry-run vs execute behavior
- Test branch protection compliance

## Future Considerations

### v0.2.0

- **Dependency Graph**: Build and analyze inter-repo dependencies
- Topological sort for upgrade ordering
- Walk-up-graph to find dependents
- Stale module detection
- Cedar policy testing framework
- Policy loading from repository
- Enhanced changelog generation

### v0.3.0

- Slack/Teams notifications
- Webhook support
- Metrics/observability
- Graph caching and incremental updates

### v0.4.0

- Web dashboard
- Trend analysis
- Multi-tenant support
