# Configuration File

VersionConductor can be configured via a YAML file.

## JSON Schema

A JSON Schema is available for IDE auto-completion and validation:

```yaml
# .versionconductor.yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/plexusone/versionconductor/main/configs/schema/versionconductor.schema.json

orgs:
  - myorg
```

The schema is located at `configs/schema/versionconductor.schema.json`.

## File Locations

Configuration is loaded from (in order of precedence):

1. `--config` flag
2. `.versionconductor.yaml` in current directory
3. `.versionconductor.yaml` in home directory

## Full Example

```yaml
# Organizations to scan
orgs:
  - myorg
  - anotherorg

# Or specific repositories
repos:
  - myorg/repo1
  - myorg/repo2

# GitHub token (can use environment variable)
token: ${GITHUB_TOKEN}

# Merge settings
merge:
  profile: quarantine
  strategy: squash
  delete-branch: true

# Policy settings
policy:
  path: ./policies/

# Release settings
release:
  generate-notes: true
  prefix: v
  draft: false

# Filtering
filter:
  include-archived: false
  include-private: true
  include-forks: false
  exclude-repos:
    - myorg/legacy-repo
    - myorg/archived-project

# Output
output:
  format: table
  verbose: false

# Graph settings (for dependency analysis)
graph:
  cache:
    enabled: true
    directory: .cache/versionconductor
    ttl: 24h
```

## Section Reference

### orgs / repos

Specify targets to scan:

```yaml
orgs:
  - plexusone
  - grokify

# Or specific repos
repos:
  - plexusone/versionconductor
  - grokify/mogo
```

### token

GitHub token for API access:

```yaml
# Direct value (not recommended)
token: ghp_xxxxx

# From environment variable
token: ${GITHUB_TOKEN}

# Or use default (GITHUB_TOKEN env var)
# Just omit the token field
```

### merge

Merge behavior settings:

```yaml
merge:
  profile: quarantine      # aggressive, balanced, conservative, quarantine
  strategy: squash         # squash, merge, rebase
  delete-branch: true      # Delete branch after merge
  max-prs: 5               # Max PRs per run
```

### policy

Cedar policy configuration:

```yaml
policy:
  path: ./policies/        # Path to .cedar files
```

### release

Release creation settings:

```yaml
release:
  generate-notes: true     # Auto-generate release notes
  prefix: v                # Version prefix (v1.0.0)
  draft: false             # Create as draft
```

### filter

Repository filtering:

```yaml
filter:
  include-archived: false
  include-private: true
  include-forks: false
  exclude-repos:
    - myorg/skip-this-repo
```

### output

Output settings:

```yaml
output:
  format: table            # table, json, markdown, csv
  verbose: false           # Enable verbose logging
```

### graph

Dependency graph settings:

```yaml
graph:
  cache:
    enabled: true
    directory: .cache/versionconductor
    ttl: 24h
```

### observability

OpenTelemetry observability settings:

```yaml
observability:
  enabled: true                    # Enable telemetry collection
  provider: otlp                   # otlp, newrelic, datadog, dynatrace
  endpoint: localhost:4317         # OTLP endpoint
  api-key: ${NEWRELIC_LICENSE_KEY} # For cloud providers
```

**Supported providers:**

| Provider | Endpoint Format | API Key |
|----------|-----------------|---------|
| `otlp` | `host:port` (gRPC) | Optional |
| `newrelic` | Auto-configured | Required (`NEWRELIC_LICENSE_KEY`) |
| `datadog` | Via DD Agent | Optional |
| `dynatrace` | `https://{id}.live.dynatrace.com/api/v2/otlp` | Required |

## Environment Variables

All settings can be overridden with environment variables:

| Setting | Environment Variable |
|---------|---------------------|
| `token` | `GITHUB_TOKEN` |
| `merge.profile` | `VERSIONCONDUCTOR_PROFILE` |
| `output.format` | `VERSIONCONDUCTOR_FORMAT` |
| `observability.enabled` | `VERSIONCONDUCTOR_OTEL_ENABLED` |
| `observability.provider` | `VERSIONCONDUCTOR_OTEL_PROVIDER` |
| `observability.endpoint` | `VERSIONCONDUCTOR_OTEL_ENDPOINT` |
| `observability.api-key` | `VERSIONCONDUCTOR_OTEL_API_KEY` |

## Precedence

Settings are applied in this order (later overrides earlier):

1. Default values
2. Config file
3. Environment variables
4. CLI flags
