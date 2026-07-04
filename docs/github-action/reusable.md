# Reusable Workflow

The reusable workflow contains all the policy logic and can be called from any repository in your organization.

## Location

```
your-org/.github
  └── .github/workflows/go-dependency-automerge.yaml
```

## Workflow Definition

```yaml
name: Go Dependency Auto-Merge

on:
  workflow_call:
    inputs:
      pr:
        description: 'PR number to evaluate immediately (skips age check)'
        required: false
        type: string
        default: ''
      profile:
        description: 'Merge profile'
        required: false
        type: string
        default: 'quarantine'
      min-age-days:
        description: 'Minimum PR age in days'
        required: false
        type: number
        default: 5
      allowed-authors:
        description: 'Comma-separated list of allowed PR authors'
        required: false
        type: string
        default: 'dependabot[bot],renovate[bot]'
      dry-run:
        description: 'If true, only evaluate but do not merge'
        required: false
        type: boolean
        default: false
      go-version:
        description: 'Go version for building versionconductor'
        required: false
        type: string
        default: '1.26.x'
      versionconductor-version:
        description: 'Version of versionconductor to use'
        required: false
        type: string
        default: 'latest'
    secrets:
      token:
        description: 'GitHub token with PR permissions'
        required: false
```

## Required Permissions

The calling workflow must grant these permissions:

```yaml
permissions:
  contents: read
  pull-requests: write
  checks: read
  statuses: read
```

## Policy Checks

The workflow evaluates these gates:

1. **Author Check** - PR author must be in allowed list
2. **Age Check** - PR must be at least N days old (unless manual dispatch)
3. **File Check** - Only go.mod and go.sum changed
4. **Directive Check** - No replace/exclude/retract/toolchain changes
5. **Go Version Check** - No go version changes
6. **Merge State** - PR must be mergeable (not draft, no conflicts)

## Concurrency

The workflow uses concurrency groups to prevent parallel runs:

```yaml
concurrency:
  group: go-dependency-automerge-${{ github.repository }}
  cancel-in-progress: false
```

This ensures only one PR is processed at a time per repository.

## One PR Per Run

The workflow processes only the oldest eligible PR per run. This avoids:

- Merge conflicts between concurrent PRs
- CI invalidation from base branch changes
- Rate limiting issues

With 15-minute schedule, you can process up to 96 PRs per day per repo.

## Updating the Workflow

When you update the central workflow:

1. All repos using `@main` get the update automatically
2. Repos pinned to a specific version need manual update

For production, consider pinning to a release tag:

```yaml
uses: your-org/.github/.github/workflows/go-dependency-automerge.yaml@v1.0.0
```

## Customization

To customize for your organization:

1. Fork or copy the workflow to your `.github` repo
2. Modify inputs, gates, or logic as needed
3. Test with `dry-run: true`
4. Roll out to repositories gradually
