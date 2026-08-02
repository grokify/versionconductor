# scan

List open dependency PRs across repositories.

## Usage

```bash
versionconductor scan [flags]
```

## Examples

```bash
# Scan an organization
versionconductor scan --orgs myorg

# Scan specific repositories
versionconductor scan --repos owner/repo1,owner/repo2

# Filter by dependency bot
versionconductor scan --orgs myorg --bot renovate

# Filter by update type
versionconductor scan --orgs myorg --update-type patch,minor

# Output as JSON
versionconductor scan --orgs myorg --format json
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--orgs` | | Organizations to scan |
| `--repos` | | Specific repos (owner/name) |
| `--bot` | | Filter by bot (dependabot, renovate) |
| `--update-type` | | Filter by type (patch, minor, major) |
| `--format` | `table` | Output format (table, json, markdown, csv) |
| `--min-age` | `0` | Minimum PR age in hours |
| `--max-age` | `0` | Maximum PR age in hours (0 = no limit) |
| `--include-archived` | `false` | Include archived repos |
| `--include-private` | `true` | Include private repos |
| `--output` | stdout | Output file |

## Output

```
┌──────────────────────────┬────────┬─────────────────────────────────────┬───────────┬──────────┐
│ REPOSITORY               │ PR     │ TITLE                               │ BOT       │ AGE      │
├──────────────────────────┼────────┼─────────────────────────────────────┼───────────┼──────────┤
│ myorg/service-api        │ #123   │ Bump github.com/foo/bar v1.2.3→1.2.4│ dependabot│ 3 days   │
└──────────────────────────┴────────┴─────────────────────────────────────┴───────────┴──────────┘
```
