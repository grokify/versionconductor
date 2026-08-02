# CLI Options

## Global Flags

These flags are available on all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--token` | GitHub token | `$GITHUB_TOKEN` |
| `--orgs` | Organizations to scan | - |
| `--repos` | Specific repos (owner/name) | - |
| `--format` | Output format | `table` |
| `--dry-run` | Show what would happen without making changes | `false` |
| `--verbose` | Enable verbose logging | `false` |
| `--config` | Config file path | `$HOME/.versionconductor.yaml` |

None of these flags have single-letter shorthands (e.g. there is no `-o` for `--orgs`) — always use the full `--flag` form.

### Observability flags

| Flag | Description | Default |
|------|-------------|---------|
| `--otel-enabled` | Enable OpenTelemetry observability | `false` |
| `--otel-provider` | Observability provider: otlp, newrelic, datadog | `otlp` |
| `--otel-endpoint` | OTLP endpoint (e.g., localhost:4317) | - |
| `--otel-api-key` | API key for cloud providers | - |

These can also be set via `observability:` in the config file or `VERSIONCONDUCTOR_OTEL_*` environment variables — see [Configuration File](file.md).

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
| `--min-age` | Minimum PR age in hours | `0` |
| `--max-age` | Maximum PR age in hours (0 = no limit) | `0` |
| `--include-archived` | Include archived repos | `false` |
| `--include-private` | Include private repos | `true` |
| `--output` | Output file (default: stdout) | - |

### review

```bash
versionconductor review [flags]
```

Decides approvals using the built-in merge profile (age, update type, CI status) — it does not evaluate Cedar policies. For Cedar-based evaluation, use [`policy evaluate`](#policy-evaluate) instead.

| Flag | Description | Default |
|------|-------------|---------|
| `--profile` | Review profile: aggressive, balanced, conservative, quarantine | `balanced` |
| `--execute` | Actually add reviews | `false` |
| `--update-type` | Filter by update type: major, minor, patch | - |
| `--bot` | Filter by dependency bot: renovate, dependabot | - |
| `--review-body` | Custom review body message | - |

### merge

```bash
versionconductor merge [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--profile` | Merge profile: aggressive, balanced, conservative | `balanced` |
| `--execute` | Actually merge PRs | `false` |
| `--strategy` | Merge strategy: merge, squash, rebase | `squash` |
| `--delete-branch` | Delete branch after merge | `true` |
| `--max-prs` | Max PRs to merge (0 = no limit) | `0` |
| `--wait-for-checks` | Wait for pending checks to complete | `false` |
| `--checks-timeout` | Timeout in seconds for waiting on checks | `300` |
| `--update-type` | Filter by update type: major, minor, patch | - |
| `--bot` | Filter by dependency bot: renovate, dependabot | - |

### release

```bash
versionconductor release [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--execute` | Actually create releases | `false` |
| `--draft` | Create releases as drafts | `false` |
| `--prerelease` | Mark releases as prereleases | `false` |
| `--generate-notes` | Use GitHub's auto-generated release notes | `true` |
| `--since` | Only consider PRs merged since this date (YYYY-MM-DD) | - |
| `--min-prs` | Minimum number of merged PRs to trigger a release | `1` |
| `--max-releases` | Maximum number of releases to create (0 = no limit) | `0` |
| `--prefix` | Version prefix | `v` |

### graph

```bash
versionconductor graph <subcommand> [flags]
```

| Subcommand | Description |
|------------|-------------|
| `build` | Build dependency graph |
| `dependents <module>` | Show dependents of a module |
| `dependencies <module>` | Show dependencies of a module |
| `order` | Show upgrade order for managed modules |
| `stale <module> --min-version <version>` | Find modules using outdated versions |
| `stats` | Show graph statistics |
| `visualize` | Generate graph visualization |

Each subcommand has its own flags (e.g. `visualize` uses `--viz-format`, not the global `--format`), plus shared caching flags (`--cache`, `--cache-dir`, `--cache-ttl`, `--no-cache`). See [graph](../commands/graph.md) for the full flag reference.

### policy evaluate

```bash
versionconductor policy evaluate [flags]
```

Evaluates Cedar policies against a PR directly (independent of `review`/`merge`).

| Flag | Description | Default |
|------|-------------|---------|
| `--repo` | Repository (owner/repo format) | - |
| `--pr` | Pull request number | `0` |
| `--policies` | Path to Cedar policies directory or file | - |
| `--profile` | Merge profile: aggressive, balanced, conservative, quarantine | `balanced` |
| `--stdin` | Read policy context from stdin as JSON instead of the GitHub API | `false` |
| `--context-file` | Read policy context from a JSON file instead of the GitHub API | - |
| `--action` | Action to evaluate: merge, review, release | `merge` |
| `--comment` | Post the decision as a comment on the PR | `false` |
| `--update-comment` | Update an existing comment instead of creating a new one | `true` |

See [policy](../commands/policy.md) for details.

## Examples

```bash
# Scan with filters
versionconductor scan --orgs myorg --bot dependabot --update-type patch

# Review with a profile
versionconductor review --orgs myorg --profile quarantine --execute

# Evaluate Cedar policies directly and post the decision as a PR comment
versionconductor policy evaluate --repo myorg/service-api --pr 123 --profile quarantine --comment

# Merge with squash strategy
versionconductor merge --orgs myorg --strategy squash --execute

# Create draft releases
versionconductor release --orgs myorg --draft --execute

# Visualize dependencies as a Mermaid diagram
versionconductor graph visualize --orgs myorg --viz-format mermaid
```
