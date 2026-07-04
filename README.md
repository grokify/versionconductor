# VersionConductor

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/versionconductor/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/versionconductor/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/versionconductor/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/versionconductor/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/versionconductor/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/versionconductor/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/versionconductor
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/versionconductor
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.github.io/versionconductor
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fversionconductor
 [loc-svg]: https://tokei.rs/b1/github/plexusone/versionconductor
 [repo-url]: https://github.com/plexusone/versionconductor
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/versionconductor/blob/main/LICENSE

Automated dependency PR management and maintenance releases for GitHub repositories.

Part of the DevOpsOrchestra suite alongside [PipelineConductor](https://github.com/grokify/pipelineconductor).

## Features

- **Scan** - Find Renovate/Dependabot PRs across organizations
- **Review** - Auto-approve dependency PRs based on Cedar policies
- **Merge** - Auto-merge approved PRs with configurable strategies
- **Release** - Create maintenance releases when dependencies are updated
- **Graph** - Analyze dependency relationships across repositories
- **GitHub Action** - Automate everything with a reusable workflow

## Installation

```bash
go install github.com/plexusone/versionconductor/cmd/versionconductor@latest
```

## Quick Start

```bash
# Set token
export GITHUB_TOKEN=ghp_your_token

# Scan for dependency PRs
versionconductor scan --orgs myorg

# Review with 5-day quarantine policy
versionconductor review --orgs myorg --profile quarantine --execute

# Merge approved PRs
versionconductor merge --orgs myorg --execute
```

## GitHub Action

Add automated dependency management to any repo:

```yaml
# .github/workflows/go-dependency-automerge.yaml
name: Go Dependency Auto-Merge

on:
  schedule:
    - cron: "7,22,37,52 * * * *"  # Every 15 minutes
  workflow_dispatch:
    inputs:
      pr:
        description: 'PR number to evaluate'
        required: false
        type: string

permissions:
  contents: read
  pull-requests: write
  checks: read
  statuses: read

jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      profile: 'quarantine'
      min-age-days: 5
    secrets: inherit
```

See [GitHub Action Setup](https://plexusone.github.io/versionconductor/github-action/setup/) for details.

## Cedar Policies

VersionConductor uses [Cedar](https://www.cedarpolicy.com/) for policy-driven automation.

The default Go dependency policy enforces:

| Gate | Description |
|------|-------------|
| Author | Only `dependabot[bot]` or `renovate[bot]` |
| Files | Only `go.mod` and `go.sum` changed |
| Directives | No `replace`/`exclude`/`toolchain` changes |
| Version | Patch or minor updates only (no major) |
| Age | 5-day quarantine period |
| CI | All checks must pass |

Example Cedar policy:

```cedar
@id("allow-patch-updates")
@action("AUTO_MERGE")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.author == "dependabot[bot]" &&
    context.pr.onlyGoModFiles == true &&
    context.goMod.hasDirectiveChanges == false &&
    context.dependency.isPatch == true &&
    context.pr.ageDays >= 5 &&
    context.ci.allPassed == true
};
```

The `@action` annotation specifies the decision outcome when the policy matches.

See [Cedar Policies](https://plexusone.github.io/versionconductor/policies/overview/) for full documentation.

## Merge Profiles

| Profile | Min Age | Patch | Minor | Major |
|---------|---------|-------|-------|-------|
| `aggressive` | 0 | Auto | Auto | Auto |
| `balanced` | 24h | Auto | Auto | Manual |
| `conservative` | 48h | Auto | Manual | Manual |
| `quarantine` | 5 days | Auto | Auto | Manual |

```bash
versionconductor review --orgs myorg --profile quarantine --execute
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | List open dependency PRs |
| `review` | Auto-approve PRs based on policy |
| `merge` | Merge approved PRs |
| `release` | Create maintenance releases |
| `graph` | Dependency graph analysis |
| `policy evaluate` | Evaluate policies against a PR |

All write commands are **dry-run by default**. Use `--execute` to perform actions.

### Policy Evaluation

Test policy evaluation locally or in CI:

```bash
# Evaluate a PR against policies
versionconductor policy evaluate --repo owner/repo --pr 123 --profile quarantine

# Post decision as PR comment
versionconductor policy evaluate --repo owner/repo --pr 123 --comment
```

Decisions include outcomes like `AUTO_MERGE`, `QUEUE_FOR_MERGE`, `MANUAL_REVIEW`, or `SECURITY_TEAM_REVIEW`.

## Documentation

Full documentation available at [plexusone.github.io/versionconductor](https://plexusone.github.io/versionconductor)

- [Getting Started](https://plexusone.github.io/versionconductor/getting-started/quickstart/)
- [Cedar Policies](https://plexusone.github.io/versionconductor/policies/overview/)
- [GitHub Action](https://plexusone.github.io/versionconductor/github-action/setup/)
- [Configuration](https://plexusone.github.io/versionconductor/configuration/profiles/)

## Development

```bash
# Clone
git clone https://github.com/plexusone/versionconductor
cd versionconductor

# Build
go build ./cmd/versionconductor

# Test
go test -v ./...

# Lint
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.
