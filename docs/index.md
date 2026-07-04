# VersionConductor

Automated dependency PR management and maintenance releases for GitHub repositories.

Part of the DevOpsOrchestra suite alongside [PipelineConductor](https://github.com/grokify/pipelineconductor).

## Features

- **Scan** - Find Renovate/Dependabot PRs across organizations
- **Review** - Auto-approve dependency PRs based on Cedar policies
- **Merge** - Auto-merge approved PRs with configurable strategies
- **Release** - Create maintenance releases when dependencies are updated
- **Graph** - Analyze dependency relationships across repositories

## Why VersionConductor?

Managing dependency updates across many repositories is time-consuming:

1. **Manual review fatigue** - Reviewing 100+ Dependabot PRs weekly is tedious
2. **Inconsistent policies** - Different repos have different merge criteria
3. **Supply chain risk** - Auto-merging without gates is dangerous
4. **Coordination overhead** - Dependencies often need to be upgraded in order

VersionConductor solves this with:

- **Cedar policy-as-code** - Define org-wide merge criteria as auditable policies
- **Safety gates** - Only merge PRs that pass all configured checks
- **GitHub Action integration** - Automate the entire workflow
- **Dependency graph analysis** - Understand upgrade order across repos

## Quick Example

```bash
# Install
go install github.com/plexusone/versionconductor/cmd/versionconductor@latest

# Set token
export GITHUB_TOKEN=ghp_your_token

# Scan for dependency PRs
versionconductor scan --orgs myorg

# Review with quarantine profile (5-day waiting period)
versionconductor review --orgs myorg --profile quarantine --execute
```

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
└─────────────────────────────────────────────────────────────────┘
```

## Next Steps

- [Installation](getting-started/installation.md) - Install VersionConductor
- [Quick Start](getting-started/quickstart.md) - Get up and running
- [Cedar Policies](policies/overview.md) - Learn about policy-driven automation
- [GitHub Action](github-action/setup.md) - Automate with GitHub Actions
