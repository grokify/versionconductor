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
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--orgs` | `-o` | Organizations to scan |
| `--repos` | `-r` | Specific repos (owner/name) |
| `--profile` | | Merge profile |
| `--policy` | | Path to Cedar policy file/directory |
| `--execute` | | Actually merge PRs (default: dry-run) |
| `--strategy` | | Merge strategy (squash, merge, rebase) |
| `--delete-branch` | | Delete branch after merge (default: true) |
| `--max-prs` | | Maximum PRs to merge |
| `--format` | `-f` | Output format |

## Merge Strategies

| Strategy | Description |
|----------|-------------|
| `squash` | Squash all commits into one (default) |
| `merge` | Create a merge commit |
| `rebase` | Rebase commits onto base branch |

## Safety Features

- **Dry-run by default** - Use `--execute` to actually merge
- **Policy evaluation** - PRs must pass all policy gates
- **Rate limiting** - Respects `--max-prs` limit
- **Conflict check** - Only merges PRs without conflicts
