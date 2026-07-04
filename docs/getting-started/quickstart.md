# Quick Start

This guide walks you through using VersionConductor to manage dependency PRs.

## Prerequisites

1. [Install VersionConductor](installation.md)
2. Create a GitHub token with required [permissions](permissions.md)

## Step 1: Set Your Token

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

## Step 2: Scan for Dependency PRs

Find all open dependency PRs in your organization:

```bash
versionconductor scan --orgs myorg
```

Example output:

```
┌──────────────────────────┬────────┬─────────────────────────────────────┬───────────┬──────────┐
│ REPOSITORY               │ PR     │ TITLE                               │ BOT       │ AGE      │
├──────────────────────────┼────────┼─────────────────────────────────────┼───────────┼──────────┤
│ myorg/service-api        │ #123   │ Bump github.com/foo/bar v1.2.3→1.2.4│ dependabot│ 3 days   │
│ myorg/service-api        │ #124   │ Update golang.org/x/oauth2 v0.15.0  │ renovate  │ 5 days   │
│ myorg/web-frontend       │ #45    │ Bump lodash from 4.17.20 to 4.17.21 │ dependabot│ 7 days   │
└──────────────────────────┴────────┴─────────────────────────────────────┴───────────┴──────────┘

Found 3 dependency PRs across 2 repositories
```

## Step 3: Review PRs (Dry Run)

Preview which PRs would be approved:

```bash
versionconductor review --orgs myorg --profile quarantine
```

The `quarantine` profile requires:

- PR age ≥ 5 days
- Only go.mod/go.sum changed
- No directive changes (replace/exclude/toolchain)
- Patch or minor update only
- All CI checks passed

## Step 4: Approve PRs

When ready, add `--execute` to actually approve:

```bash
versionconductor review --orgs myorg --profile quarantine --execute
```

## Step 5: Merge PRs

Merge approved PRs:

```bash
versionconductor merge --orgs myorg --execute
```

## Step 6: Create Releases (Optional)

Create maintenance releases for repos with merged dependency PRs:

```bash
versionconductor release --orgs myorg --execute
```

## Next Steps

- Learn about [Cedar Policies](../policies/overview.md) for fine-grained control
- Set up [GitHub Action automation](../github-action/setup.md)
- Configure [merge profiles](../configuration/profiles.md)
