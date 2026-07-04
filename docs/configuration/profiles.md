# Merge Profiles

Merge profiles define the criteria for auto-merging dependency PRs. VersionConductor includes four built-in profiles.

## Profile Comparison

| Profile | Min Age | Patch | Minor | Major | Max PRs/Run |
|---------|---------|-------|-------|-------|-------------|
| `aggressive` | 0 | Auto | Auto | Auto | Unlimited |
| `balanced` | 24h | Auto | Auto | Manual | 10 |
| `conservative` | 48h | Auto | Manual | Manual | 5 |
| `quarantine` | 5 days | Auto | Auto | Manual | 1 |

## Profile Details

### aggressive

Merge all passing dependency PRs immediately.

```yaml
name: aggressive
minAgeHours: 0
autoMergePatch: true
autoMergeMinor: true
autoMergeMajor: true
requireApproval: false
maxPRsPerRun: 0  # No limit
```

!!! warning "Use with Caution"
    This profile has no safety delays. Only use for low-risk internal projects.

### balanced

Wait 24 hours, auto-merge patch and minor updates.

```yaml
name: balanced
minAgeHours: 24
autoMergePatch: true
autoMergeMinor: true
autoMergeMajor: false
requireApproval: false
maxPRsPerRun: 10
```

Good for:

- Active projects with good test coverage
- Teams that review PRs daily

### conservative

Wait 48 hours, auto-merge only patch updates.

```yaml
name: conservative
minAgeHours: 48
autoMergePatch: true
autoMergeMinor: false
autoMergeMajor: false
requireApproval: true
maxPRsPerRun: 5
```

Good for:

- Production-critical applications
- Projects with complex dependencies

### quarantine (Recommended)

5-day waiting period with one PR per run.

```yaml
name: quarantine
minAgeHours: 120  # 5 days
minAgeDays: 5
autoMergePatch: true
autoMergeMinor: true
autoMergeMajor: false
requireApproval: false
maxPRsPerRun: 1
```

This is the **recommended profile** for Go projects because:

- Allows time for security issues to surface
- One PR per run avoids merge conflicts
- Aligns with community best practices (N+5 rule)

## Using Profiles

### CLI

```bash
versionconductor review --orgs myorg --profile quarantine
versionconductor merge --orgs myorg --profile quarantine --execute
```

### Config File

```yaml
merge:
  profile: quarantine
```

### GitHub Action

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      profile: quarantine
```

## Profile Settings

| Setting | Type | Description |
|---------|------|-------------|
| `name` | String | Profile identifier |
| `description` | String | Human-readable description |
| `minAgeHours` | Integer | Minimum PR age in hours |
| `minAgeDays` | Integer | Minimum PR age in days |
| `maxAgeHours` | Integer | Maximum PR age (0 = no limit) |
| `autoMergePatch` | Boolean | Auto-merge patch updates |
| `autoMergeMinor` | Boolean | Auto-merge minor updates |
| `autoMergeMajor` | Boolean | Auto-merge major updates |
| `requireAllChecks` | Boolean | Require all CI checks to pass |
| `allowPendingChecks` | Boolean | Allow pending checks |
| `mergeStrategy` | String | Merge method (squash, merge, rebase) |
| `deleteBranch` | Boolean | Delete branch after merge |
| `requireApproval` | Boolean | Require PR approval before merge |
| `maxPRsPerRun` | Integer | Max PRs to merge per run (0 = no limit) |

## Custom Profiles

Define custom profiles in your config file:

```yaml
profiles:
  security-critical:
    name: security-critical
    description: Strict policy for security-sensitive repos
    minAgeDays: 7
    autoMergePatch: true
    autoMergeMinor: false
    autoMergeMajor: false
    requireAllChecks: true
    maxPRsPerRun: 1
```

Use with:

```bash
versionconductor merge --profile security-critical
```
