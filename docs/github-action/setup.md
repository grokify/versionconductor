# GitHub Action Setup

VersionConductor provides a reusable GitHub Action workflow for automating Go dependency PR management.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Organization (.github repo)                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  .github/workflows/go-dependency-automerge.yaml         │    │
│  │  (Reusable workflow with policy logic)                  │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
┌───────────────────▼───┐   ┌───────────▼───────────────┐
│       repo-a          │   │         repo-b            │
│  ┌─────────────────┐  │   │  ┌─────────────────┐      │
│  │ Wrapper workflow│  │   │  │ Wrapper workflow│      │
│  │ (calls central) │  │   │  │ (calls central) │      │
│  └─────────────────┘  │   │  └─────────────────┘      │
└───────────────────────┘   └───────────────────────────┘
```

## Quick Start

### 1. Add Wrapper Workflow to Your Repo

Create `.github/workflows/go-dependency-automerge.yaml`:

```yaml
name: Go Dependency Auto-Merge

on:
  schedule:
    - cron: "7,22,37,52 * * * *"  # Every 15 minutes
  workflow_dispatch:
    inputs:
      pr:
        description: 'PR number to evaluate immediately'
        required: false
        type: string
      dry-run:
        description: 'Evaluate only, do not merge'
        required: false
        type: boolean
        default: false

permissions:
  contents: read
  pull-requests: write
  checks: read
  statuses: read

jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      pr: ${{ inputs.pr || '' }}
      profile: 'quarantine'
      min-age-days: 5
      dry-run: ${{ inputs.dry-run || false }}
    secrets: inherit
```

### 2. Configure Branch Protection

Ensure your repository has:

- Required status checks enabled
- Auto-merge enabled (Settings → General → Allow auto-merge)

### 3. Test with Workflow Dispatch

1. Go to Actions tab in your repository
2. Select "Go Dependency Auto-Merge"
3. Click "Run workflow"
4. Optionally enter a PR number to test specific PR
5. Enable "dry-run" to test without actually merging

## How It Works

1. **Schedule or Dispatch** - Workflow runs every 15 minutes or manually
2. **Find PRs** - Lists open dependency PRs from allowed authors
3. **Select Oldest** - Picks the oldest eligible PR (one per run)
4. **Evaluate** - Checks all policy gates
5. **Approve** - Approves PR if all gates pass
6. **Enable Auto-Merge** - Sets PR to merge when checks pass

## Configuration Options

| Input | Default | Description |
|-------|---------|-------------|
| `pr` | - | Specific PR to evaluate (skips age check) |
| `profile` | `quarantine` | Merge profile |
| `min-age-days` | `5` | Minimum PR age in days |
| `allowed-authors` | `dependabot[bot],renovate[bot]` | Allowed PR authors |
| `dry-run` | `false` | Evaluate only |
| `go-version` | `1.26.x` | Go version for building |
| `versionconductor-version` | `latest` | VersionConductor version |

## Next Steps

- [Reusable Workflow](reusable.md) - Details on the central workflow
- [Examples](examples.md) - Common configuration patterns
