# CLI Options

## Global Flags

These flags are available on all commands:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--token` | `-t` | GitHub token | `$GITHUB_TOKEN` |
| `--orgs` | `-o` | Organizations to scan | - |
| `--repos` | `-r` | Specific repos (owner/name) | - |
| `--format` | `-f` | Output format | `table` |
| `--verbose` | `-v` | Enable verbose logging | `false` |
| `--config` | `-c` | Config file path | `.versionconductor.yaml` |

## Output Formats

| Format | Description |
|--------|-------------|
| `table` | Human-readable ASCII table |
| `json` | JSON for programmatic use |
| `markdown` | Markdown for reports |
| `csv` | CSV for spreadsheets |

## Command-Specific Flags

### scan

```bash
versionconductor scan [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--bot` | Filter by bot (dependabot, renovate) | - |
| `--update-type` | Filter by type (patch, minor, major) | - |
| `--include-archived` | Include archived repos | `false` |
| `--include-private` | Include private repos | `true` |

### review

```bash
versionconductor review [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--profile` | Merge profile to use | `balanced` |
| `--policy` | Path to Cedar policy file/dir | - |
| `--execute` | Actually approve PRs | `false` |
| `--max-prs` | Max PRs to review | `0` (unlimited) |

### merge

```bash
versionconductor merge [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--profile` | Merge profile to use | `balanced` |
| `--policy` | Path to Cedar policy file/dir | - |
| `--execute` | Actually merge PRs | `false` |
| `--strategy` | Merge strategy | `squash` |
| `--delete-branch` | Delete branch after merge | `true` |
| `--max-prs` | Max PRs to merge | `0` (unlimited) |

### release

```bash
versionconductor release [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--execute` | Actually create releases | `false` |
| `--since` | Only PRs merged since date | - |
| `--draft` | Create as draft releases | `false` |
| `--prefix` | Version prefix | `v` |
| `--generate-notes` | Auto-generate release notes | `true` |

### graph

```bash
versionconductor graph <subcommand> [flags]
```

Subcommands:

| Subcommand | Description |
|------------|-------------|
| `build` | Build dependency graph |
| `dependents` | Show dependents of a module |
| `dependencies` | Show dependencies of a module |
| `order` | Show upgrade order |
| `stale` | Find stale dependencies |
| `stats` | Show graph statistics |
| `visualize` | Generate graph visualization |

## Examples

```bash
# Scan with filters
versionconductor scan --orgs myorg --bot dependabot --update-type patch

# Review with Cedar policies
versionconductor review --orgs myorg --policy ./policies/ --profile quarantine

# Merge with squash strategy
versionconductor merge --orgs myorg --strategy squash --execute

# Create draft releases
versionconductor release --orgs myorg --draft --execute

# Visualize dependencies
versionconductor graph visualize --orgs myorg --format mermaid
```
