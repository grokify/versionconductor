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

# Only approve patch updates from a specific bot
versionconductor review --orgs myorg --update-type patch --bot dependabot --execute
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--orgs` | | Organizations to scan |
| `--repos` | | Specific repos (owner/name) |
| `--profile` | `balanced` | Review profile: aggressive, balanced, conservative, quarantine |
| `--execute` | `false` | Actually add reviews (default: dry-run) |
| `--update-type` | | Filter by update type: major, minor, patch |
| `--bot` | | Filter by dependency bot: renovate, dependabot |
| `--review-body` | | Custom review body message |
| `--format` | `table` | Output format |

## Profiles

| Profile | Min Age | Auto-Approve |
|---------|---------|--------------|
| aggressive | 0 | All updates |
| balanced | 24h | Patch + Minor |
| conservative | 48h | Patch only |
| quarantine | 5 days | Patch + Minor |

## Profile vs. Cedar Policy Evaluation

`review` decides approvals using the built-in merge profile above (age, update type, CI status) — it does not load Cedar policies. For Cedar-based policy evaluation (quarantine gates, decision outcomes, PR comments), use [`policy evaluate`](policy.md) instead.

See [Cedar Policies](../policies/overview.md) for details on the Cedar-based approach.
