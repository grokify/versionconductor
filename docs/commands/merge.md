# merge

Merge approved dependency PRs.

## Usage

```bash
versionconductor merge [flags]
```

## Examples

```bash
# Dry-run (default)
versionconductor merge --orgs myorg

# Actually merge
versionconductor merge --orgs myorg --execute

# Use squash merge
versionconductor merge --orgs myorg --strategy squash --execute

# Limit merges per run
versionconductor merge --orgs myorg --max-prs 5 --execute

# Keep branches
versionconductor merge --orgs myorg --delete-branch=false --execute

# Wait for pending checks to complete before merging
versionconductor merge --orgs myorg --wait-for-checks --checks-timeout 600 --execute
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--orgs` | | Organizations to scan |
| `--repos` | | Specific repos (owner/name) |
| `--profile` | `balanced` | Merge profile: aggressive, balanced, conservative |
| `--execute` | `false` | Actually merge PRs (default: dry-run) |
| `--strategy` | `squash` | Merge strategy (squash, merge, rebase) |
| `--delete-branch` | `true` | Delete branch after merge |
| `--max-prs` | `0` | Maximum PRs to merge (0 = no limit) |
| `--wait-for-checks` | `false` | Wait for pending checks to complete |
| `--checks-timeout` | `300` | Timeout in seconds for waiting on checks |
| `--update-type` | | Filter by update type: major, minor, patch |
| `--bot` | | Filter by dependency bot: renovate, dependabot |
| `--format` | `table` | Output format |

## Merge Strategies

| Strategy | Description |
|----------|-------------|
| `squash` | Squash all commits into one (default) |
| `merge` | Create a merge commit |
| `rebase` | Rebase commits onto base branch |

## Safety Features

- **Dry-run by default** - Use `--execute` to actually merge
- **Profile evaluation** - PRs must pass the selected merge profile's age, update-type, and CI-status gates
- **Rate limiting** - Respects `--max-prs` limit
- **Conflict check** - Only merges PRs without conflicts

For Cedar-based policy evaluation (quarantine gates, decision outcomes, PR comments) instead of profiles, use [`policy evaluate`](policy.md).
