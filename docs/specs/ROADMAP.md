# VersionConductor Roadmap

This document tracks planned improvements and their implementation status.

## Priority: High

### 1. Unit Tests for Core Packages

**Status:** Not Started

Add comprehensive unit tests for packages with 0% coverage:

| Package | Current | Target | Notes |
|---------|---------|--------|-------|
| `internal/collector` | 0% | 80% | Mock GitHub client, test transformations |
| `internal/merger` | 0% | 80% | Mock merge operations, test strategies |
| `internal/releaser` | 0% | 80% | Mock tag/release creation, test semver |
| `internal/report` | 0% | 80% | Test all output formats |
| `pkg/model` | 0% | 90% | Test helper methods, parsing |

**Implementation:**

- Create `internal/collector/mock_test.go` with GitHub client mock
- Create `internal/merger/mock_test.go` with merger interface mock
- Create `internal/releaser/mock_test.go` with releaser interface mock
- Add table-driven tests for all public functions

### 2. Rate Limit Handling

**Status:** Complete

Add GitHub API rate limit awareness with exponential backoff:

- Check `X-RateLimit-Remaining` header on responses
- Implement exponential backoff with jitter
- Add `--rate-limit-buffer` flag (default: 100 requests)
- Log warnings when approaching rate limit
- Support primary and secondary rate limits

**Files to modify:**

- `internal/collector/github.go` - Add rate limit checking
- `internal/collector/ratelimit.go` - New file for rate limit logic
- `cmd/versionconductor/cmd/root.go` - Add global flags

## Priority: Medium

### 3. Structured Logging with slog

**Status:** Complete

Replace ad-hoc logging with structured slog logging via omniobserve:

**Dependency:**

```go
import "github.com/plexusone/omniobserve/observops"
```

**Implementation:**

- Initialize observops provider in root command
- Pass `*slog.Logger` via context to all packages
- Add trace correlation (trace_id, span_id) to logs
- Log levels: DEBUG (verbose), INFO (normal), WARN (issues), ERROR (failures)

**Structured log fields:**

| Field | Description |
|-------|-------------|
| `org` | GitHub organization |
| `repo` | Repository name |
| `pr` | PR number |
| `author` | PR author |
| `action` | Operation being performed |
| `decision` | Policy decision outcome |
| `duration_ms` | Operation duration |

### 4. Configuration Schema Documentation

**Status:** Complete

Document `.versionconductor.yaml` configuration file:

- Create JSON Schema for config validation
- Document all configuration options in `docs/configuration/file.md`
- Add example configurations for common use cases
- Validate config on load with helpful error messages

**Config structure:**

```yaml
# .versionconductor.yaml
orgs:
  - plexusone
  - myorg

profiles:
  default: quarantine

authors:
  allowed:
    - dependabot[bot]
    - renovate[bot]

observability:
  enabled: true
  provider: otlp
  endpoint: localhost:4317
```

### 5. GitHub Action Integration Tests

**Status:** Complete

Add integration tests for the reusable workflow:

- Test with act (local GitHub Actions runner)
- Mock GitHub API responses
- Test all authentication paths (App, token, GITHUB_TOKEN)
- Test policy evaluation scenarios
- Add workflow_dispatch test cases

## Priority: Low

### 6. OpenTelemetry Observability via OmniObserve

**Status:** Not Started

Add production observability using omniobserve's observops package:

**Supported backends:**

| Provider | Protocol | Configuration |
|----------|----------|---------------|
| OTLP | gRPC/HTTP | `--otel-endpoint` |
| New Relic | OTLP/gRPC | `--newrelic-key`, `--newrelic-region` |
| Datadog | OTLP/gRPC | `--datadog-site` (via DD Agent) |
| Dynatrace | OTLP/HTTP | `--dynatrace-endpoint`, `--dynatrace-token` |

**Implementation:**

```go
import (
    "github.com/plexusone/omniobserve/observops"
    _ "github.com/plexusone/omniobserve/observops/otlp"
    _ "github.com/plexusone/omniobserve/observops/newrelic"
    _ "github.com/plexusone/omniobserve/observops/datadog"
)

// Initialize provider
provider, err := observops.Open("otlp",
    observops.WithEndpoint(cfg.OtelEndpoint),
    observops.WithServiceName("versionconductor"),
)
defer provider.Shutdown(ctx)

// Create spans for operations
ctx, span := provider.Tracer().Start(ctx, "EvaluatePR")
defer span.End()

// Metrics
counter, _ := provider.Meter().Counter("prs_evaluated_total")
counter.Add(ctx, 1, observops.WithAttributes(
    observops.Attribute("decision", decision.Outcome.String()),
))
```

**Traces:**

| Span | Attributes |
|------|------------|
| `ScanOrganization` | org, repo_count |
| `EvaluatePR` | repo, pr, author, decision |
| `MergePR` | repo, pr, strategy, success |
| `CreateRelease` | repo, version, pr_count |
| `PolicyEvaluate` | profile, gates_passed, gates_failed |

**Metrics:**

| Metric | Type | Labels |
|--------|------|--------|
| `prs_scanned_total` | Counter | org, bot |
| `prs_evaluated_total` | Counter | decision, profile |
| `prs_merged_total` | Counter | org, strategy |
| `releases_created_total` | Counter | org |
| `policy_evaluation_duration_seconds` | Histogram | profile |
| `github_api_requests_total` | Counter | endpoint, status |
| `github_api_rate_limit_remaining` | Gauge | - |

**Configuration flags:**

```
--otel-enabled          Enable OpenTelemetry (default: false)
--otel-provider         Provider: otlp, newrelic, datadog, dynatrace
--otel-endpoint         OTLP endpoint (default: localhost:4317)
--otel-service-name     Service name (default: versionconductor)
--newrelic-key          New Relic license key
--newrelic-region       New Relic region: us, eu (default: us)
```

### 7. GitHub API Response Caching

**Status:** Not Started

Add caching for GitHub API responses in collector:

- Cache PR details, check runs, repository metadata
- Configurable TTL (default: 5 minutes)
- In-memory cache for CLI, Redis option for long-running
- Cache invalidation on write operations
- Add `--cache-ttl` and `--no-cache` flags

**Files to modify:**

- `internal/collector/cache.go` - New cache implementation
- `internal/collector/github.go` - Wrap API calls with cache

## Implementation Order

1. **Phase 1 - Foundation** (High Priority)
   - [x] Unit tests for collector (43.8% coverage)
   - [x] Unit tests for merger (4.3% coverage)
   - [x] Unit tests for releaser (73.8% coverage)
   - [x] Rate limit handling

2. **Phase 2 - Observability** (Medium Priority)
   - [x] slog structured logging via omniobserve (78.8% coverage)
   - [x] Configuration schema documentation
   - [x] GitHub Action integration tests

3. **Phase 3 - Production Readiness** (Low Priority)
   - [ ] OpenTelemetry tracing and metrics
   - [ ] API response caching

## Changelog

| Date | Change |
|------|--------|
| 2026-07-04 | Initial roadmap created from project review |
