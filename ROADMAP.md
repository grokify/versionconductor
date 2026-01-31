# VersionConductor Roadmap

This document outlines the development roadmap for VersionConductor, prioritized by user value and implementation complexity.

## Version Overview

| Version | Theme | Target |
|---------|-------|--------|
| v0.1.x | Foundation & Stability | Q1 2026 |
| v0.2.0 | Dependency Graph | Q2 2026 |
| v0.3.0 | Orchestration & Scheduling | Q2 2026 |
| v0.4.0 | Multi-Language & Integrations | Q3 2026 |
| v1.0.0 | Production Ready | Q4 2026 |

---

## v0.1.x - Foundation & Stability

**Theme:** Make the core features production-ready with proper testing.

### v0.1.1 - Testing Foundation

| Priority | Item | Description |
|----------|------|-------------|
| P0 | Unit tests for semver | Test version parsing, comparison, bumping |
| P0 | Unit tests for PR detection | Test Renovate/Dependabot author detection |
| P0 | Unit tests for title parsing | Test dependency extraction from PR titles |
| P1 | Unit tests for report generation | Test JSON/Markdown output |
| P1 | Unit tests for policy profiles | Test built-in profile loading |
| P2 | Integration test framework | Mock GitHub API for e2e tests |

### v0.1.2 - Reliability

| Priority | Item | Description |
|----------|------|-------------|
| P0 | Rate limit handling | Respect X-RateLimit headers, implement backoff |
| P0 | Error handling improvements | Better error messages, recovery from transient failures |
| P1 | Retry logic | Retry failed API calls with exponential backoff |
| P1 | Progress reporting | Show progress when scanning many repos |
| P2 | Caching | Cache repo metadata to reduce API calls |

### v0.1.3 - Polish

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Confirmation prompts | Require confirmation for large batch operations |
| P1 | Verbose logging | Add --verbose flag for debugging |
| P2 | Config file validation | Validate config on load, helpful error messages |
| P2 | Shell completions | Bash/Zsh/Fish completions for CLI |

---

## v0.2.0 - Dependency Graph

**Theme:** Build and analyze dependency relationships across repos.

### Core Graph Implementation

| Priority | Item | Description |
|----------|------|-------------|
| P0 | go.mod parser | Parse module path, require, replace, exclude |
| P0 | Graph data structure | Directed graph with nodes (modules) and edges (dependencies) |
| P0 | Graph builder | Fetch go.mod from repos, build graph |
| P0 | Cycle detection | Detect and report cycles (shouldn't exist but handle gracefully) |
| P1 | Topological sort | Compute correct upgrade order |
| P1 | Dependents query | Find all modules that depend on X |
| P1 | Dependencies query | Find all modules that X depends on |

### Graph CLI Commands

| Priority | Item | Description |
|----------|------|-------------|
| P0 | `graph build` | Build graph from specified orgs |
| P0 | `graph dependents` | List dependents of a module |
| P1 | `graph dependencies` | List dependencies of a module |
| P1 | `graph order` | Show topological upgrade order |
| P1 | `graph stale` | Find modules using outdated versions |
| P2 | `graph visualize` | Output DOT format for Graphviz |

### Graph Caching

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Local cache | Cache graph to disk, invalidate on TTL |
| P2 | Incremental updates | Update only changed repos |
| P2 | Cache warming | Background refresh of cache |

### Testing

| Priority | Item | Description |
|----------|------|-------------|
| P0 | go.mod parser tests | Test various go.mod formats |
| P0 | Graph algorithm tests | Test topo sort, cycle detection |
| P1 | Integration tests | Test graph building with mocked API |

---

## v0.3.0 - Orchestration & Scheduling

**Theme:** Automate graph-wide upgrades on a schedule.

### Upgrade Orchestration

| Priority | Item | Description |
|----------|------|-------------|
| P0 | `upgrade` command | Upgrade graph in topological order |
| P0 | Upgrade dry-run | Preview what would be upgraded |
| P0 | Release creation | Create patch release after successful merge |
| P1 | Wait for CI | Poll for CI status before releasing |
| P1 | Upgrade report | Summary of all actions taken |
| P2 | Rollback info | Include info to revert if needed |

### Scheduling

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Cron expression support | Parse cron expressions for scheduling |
| P1 | `upgrade schedule` | Output cron job / systemd timer config |
| P2 | GitHub Action | Reusable workflow for scheduled upgrades |
| P2 | Status persistence | Track last upgrade run, next scheduled |

### Release Propagation

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Wait for PR creation | Poll for Renovate/Dependabot to create PRs |
| P2 | Trigger Renovate | Use Renovate API to trigger immediate check |
| P2 | Webhook receiver | Receive webhooks for faster propagation |

### Safety

| Priority | Item | Description |
|----------|------|-------------|
| P0 | Batch limits | Limit number of merges/releases per run |
| P0 | Failure handling | Stop on first failure vs continue |
| P1 | Conflict detection | Detect if upgrading A will conflict with B |
| P1 | Audit log | Detailed log of all actions for compliance |

---

## v0.4.0 - Multi-Language & Integrations

**Theme:** Expand beyond Go, integrate with more tools.

### Multi-Language Support

| Priority | Item | Description |
|----------|------|-------------|
| P1 | npm/package.json | Parse Node.js dependencies |
| P1 | Python/pyproject.toml | Parse Python dependencies |
| P2 | Rust/Cargo.toml | Parse Rust dependencies |
| P2 | Language detection | Auto-detect primary language of repo |

### Notifications

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Slack integration | Post upgrade summaries to Slack |
| P1 | Email digest | Send email summary of upgrades |
| P2 | Microsoft Teams | Post to Teams channels |
| P2 | Webhook output | POST results to custom endpoint |

### Integrations

| Priority | Item | Description |
|----------|------|-------------|
| P1 | GitHub Action | Published action for easy CI use |
| P2 | GitLab support | Support GitLab API in addition to GitHub |
| P2 | Bitbucket support | Support Bitbucket API |

---

## v1.0.0 - Production Ready

**Theme:** Battle-tested, documented, ready for enterprise use.

### Stability

| Priority | Item | Description |
|----------|------|-------------|
| P0 | Comprehensive test coverage | >80% coverage on critical paths |
| P0 | Production validation | Run on real orgs for 3+ months |
| P0 | Performance optimization | Handle 1000+ repos efficiently |
| P1 | Security audit | Review for credential handling, injection |

### Documentation

| Priority | Item | Description |
|----------|------|-------------|
| P0 | Complete user guide | Full documentation site |
| P0 | Troubleshooting guide | Common issues and solutions |
| P1 | Video tutorials | Demo videos for key workflows |
| P1 | Best practices guide | Recommended policies, schedules |

### Enterprise Features

| Priority | Item | Description |
|----------|------|-------------|
| P1 | GitHub App mode | Use GitHub App instead of PAT |
| P1 | Multi-token support | Different tokens per org |
| P2 | SAML/SSO integration | Enterprise authentication |
| P2 | Metrics/observability | Prometheus metrics, OpenTelemetry |

### Community

| Priority | Item | Description |
|----------|------|-------------|
| P1 | Policy library | Community-contributed Cedar policies |
| P1 | Plugin system | Extensibility for custom logic |
| P2 | Web dashboard | Optional web UI for visibility |

---

## Priority Definitions

| Priority | Meaning |
|----------|---------|
| P0 | Must have for the release. Blocks release if not done. |
| P1 | Should have. High value, do if time permits. |
| P2 | Nice to have. Can defer to next release. |

---

## Current Status

### Implemented (v0.1.0)

- [x] CLI framework (Cobra/Viper)
- [x] GitHub collector for PR discovery
- [x] Renovate/Dependabot detection
- [x] Policy engine skeleton
- [x] Built-in merge profiles
- [x] PR merge with strategy selection
- [x] Release creation with semver bumping
- [x] JSON/Markdown reports

### Not Yet Implemented

- [ ] Unit tests
- [ ] Integration tests
- [ ] Cedar policy evaluation (wired up)
- [ ] go.mod parsing
- [ ] Dependency graph
- [ ] Topological sort
- [ ] Scheduled upgrades
- [ ] Rate limit handling
- [ ] Caching

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to VersionConductor.

Priorities are set by maintainers based on user feedback. Open an issue to discuss roadmap items or propose new features.

---

## Related Projects

- [PipelineConductor](https://github.com/grokify/pipelineconductor) - CI/CD pipeline compliance scanning
- [gogithub](https://github.com/grokify/gogithub) - GitHub API helper library
