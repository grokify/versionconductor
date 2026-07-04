# Workflow Examples

Common configuration patterns for the Go dependency auto-merge workflow.

## Standard Setup (Recommended)

5-day quarantine, one PR per run:

```yaml
name: Go Dependency Auto-Merge

on:
  schedule:
    - cron: "7,22,37,52 * * * *"
  workflow_dispatch:
    inputs:
      pr:
        description: 'PR number to evaluate immediately'
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

## Faster for Internal Projects

24-hour wait, more aggressive:

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      profile: 'balanced'
      min-age-days: 1
    secrets: inherit
```

## Self-Hosted Renovate

Add your custom bot to allowed authors:

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      allowed-authors: 'dependabot[bot],renovate[bot],my-org-renovate[bot]'
    secrets: inherit
```

## Daily Schedule Instead of 15-Minute

Less frequent for low-volume repos:

```yaml
on:
  schedule:
    - cron: "0 9 * * *"  # Daily at 9 AM UTC
  workflow_dispatch:
    inputs:
      pr:
        required: false
        type: string
```

## With Custom Token

Using a PAT or GitHub App token:

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      profile: 'quarantine'
    secrets:
      token: ${{ secrets.DEPENDENCY_BOT_TOKEN }}
```

## Pinned Version

Pin to specific versionconductor release:

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      versionconductor-version: 'v0.3.0'
    secrets: inherit
```

## Testing Setup

Always dry-run, manual dispatch only:

```yaml
on:
  workflow_dispatch:
    inputs:
      pr:
        description: 'PR number to test'
        required: true
        type: string

jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      pr: ${{ inputs.pr }}
      dry-run: true
    secrets: inherit
```

## Multi-Repo Rollout

To add the workflow to multiple repos:

```bash
# Create workflow file
cat > .github/workflows/go-dependency-automerge.yaml << 'EOF'
name: Go Dependency Auto-Merge

on:
  schedule:
    - cron: "7,22,37,52 * * * *"
  workflow_dispatch:
    inputs:
      pr:
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
    secrets: inherit
EOF

# Copy to each repo
for repo in repo1 repo2 repo3; do
  cp .github/workflows/go-dependency-automerge.yaml ../$repo/.github/workflows/
done
```

## Monitoring

Check workflow runs:

```bash
# List recent runs
gh run list --workflow=go-dependency-automerge.yaml

# View specific run
gh run view <run-id>

# Watch logs
gh run watch <run-id>
```

Check merged PRs:

```bash
# List recently merged dependency PRs
gh pr list --state merged --author dependabot[bot] --limit 10
```
