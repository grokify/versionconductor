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

## Authentication Options

The workflow supports three authentication methods in priority order:

### Option 1: GitHub App (Recommended)

Use the `versionconductor[bot]` identity for clearer audit trails.

**Benefits:**

- Clear bot identity in PR approvals
- Better audit trail
- Short-lived tokens (1 hour)
- Fine-grained permissions per installation

**Setup:**

1. Create a GitHub App named `versionconductor` with permissions:
   - Pull requests: read/write
   - Contents: read
   - Checks: read
   - Commit statuses: read
   - Metadata: read

2. Install the app on your organization/repositories

3. Add secrets to your repository:
   - `VERSIONCONDUCTOR_APP_ID` - The App ID
   - `VERSIONCONDUCTOR_PRIVATE_KEY` - The private key (PEM format)

4. Configure the workflow:

```yaml
secrets:
  app-id: ${{ secrets.VERSIONCONDUCTOR_APP_ID }}
  app-private-key: ${{ secrets.VERSIONCONDUCTOR_PRIVATE_KEY }}
```

### Option 2: Custom Token

Use a Personal Access Token (PAT) with appropriate permissions.

```yaml
secrets:
  token: ${{ secrets.MY_GITHUB_TOKEN }}
```

### Option 3: Default GITHUB_TOKEN

No configuration needed. The workflow uses the repository's default token.

```yaml
secrets: inherit
```

!!! note "Token Limitations"
    The default `GITHUB_TOKEN` appears as `github-actions[bot]` and cannot
    trigger other workflows. For better audit trails, use the GitHub App option.

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
    secrets:
      # GitHub App (recommended)
      app-id: ${{ secrets.VERSIONCONDUCTOR_APP_ID }}
      app-private-key: ${{ secrets.VERSIONCONDUCTOR_PRIVATE_KEY }}
      # Or use: secrets: inherit
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
2. **Authenticate** - Uses GitHub App token or falls back to GITHUB_TOKEN
3. **Find PRs** - Lists open dependency PRs from allowed authors
4. **Select Oldest** - Picks the oldest eligible PR (one per run)
5. **Evaluate** - Checks all policy gates
6. **Approve** - Approves PR as `versionconductor[bot]` or `github-actions[bot]`
7. **Enable Auto-Merge** - Sets PR to merge when checks pass

## Configuration Options

| Input | Default | Description |
|-------|---------|-------------|
| `pr` | - | Specific PR to evaluate (skips age check) |
| `profile` | `quarantine` | Merge profile |
| `min-age-days` | `5` | Minimum PR age in days |
| `allowed-authors` | `dependabot[bot],renovate[bot]` | Allowed PR authors |
| `dry-run` | `false` | Evaluate only |
| `go-version` | `1.24.x` | Go version for building |
| `versionconductor-version` | `latest` | VersionConductor version |

## Secrets

| Secret | Required | Description |
|--------|----------|-------------|
| `app-id` | No | GitHub App ID for `versionconductor[bot]` |
| `app-private-key` | No | GitHub App private key |
| `token` | No | Custom GitHub token (fallback) |

## Next Steps

- [Reusable Workflow](reusable.md) - Details on the central workflow
- [Examples](examples.md) - Common configuration patterns
