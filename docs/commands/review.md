# review

Auto-approve dependency PRs that meet policy criteria.

## Usage

```bash
versionconductor review [flags]
```

## Examples

```bash
# Dry-run (default)
versionconductor review --orgs myorg

# Actually approve
versionconductor review --orgs myorg --execute

# Use specific profile
versionconductor review --orgs myorg --profile quarantine --execute

# Use Cedar policies
versionconductor review --orgs myorg --policy ./policies/ --execute
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--orgs` | `-o` | Organizations to scan |
| `--repos` | `-r` | Specific repos (owner/name) |
| `--profile` | | Merge profile (default: balanced) |
| `--policy` | | Path to Cedar policy file/directory |
| `--execute` | | Actually approve PRs (default: dry-run) |
| `--max-prs` | | Maximum PRs to review |
| `--format` | `-f` | Output format |

## Profiles

| Profile | Min Age | Auto-Approve |
|---------|---------|--------------|
| aggressive | 0 | All updates |
| balanced | 24h | Patch + Minor |
| conservative | 48h | Patch only |
| quarantine | 5 days | Patch + Minor |

## Policy Evaluation

When `--policy` is specified, Cedar policies are loaded and evaluated. The PR is approved only if:

1. A `permit` policy matches
2. No `forbid` policy matches
3. All policy conditions are satisfied

See [Cedar Policies](../policies/overview.md) for details.
