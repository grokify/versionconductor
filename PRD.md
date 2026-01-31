# VersionConductor - Product Requirements Document

## Executive Summary

VersionConductor is a CLI tool for automated dependency PR management and maintenance releases across multiple GitHub repositories. It scans for Renovate and Dependabot PRs, applies policy-based auto-merge decisions, and creates maintenance releases when dependencies are updated.

## Problem Statement

Organizations with many repositories face significant overhead managing dependency updates:

1. **PR Fatigue**: Renovate and Dependabot generate hundreds of PRs weekly across repositories
2. **Manual Review Burden**: Each PR requires human review, even for safe patch updates
3. **Inconsistent Policies**: Different teams apply different merge criteria
4. **Release Lag**: Dependencies get merged but releases aren't cut, leaving users on outdated versions
5. **Security Delays**: Critical security patches sit in PR queues waiting for review

## Goals

1. **Reduce Manual Effort**: Automate review and merge of low-risk dependency updates
2. **Enforce Consistent Policies**: Cedar policy-as-code for org-wide merge criteria
3. **Accelerate Security Patches**: Fast-track critical updates with appropriate policies
4. **Automate Maintenance Releases**: Create patch releases when dependencies are updated
5. **Maintain Visibility**: Provide clear reporting on dependency update status

## Non-Goals

1. Not a replacement for Renovate/Dependabot (complements them)
2. Not a general-purpose PR management tool (focused on dependency updates)
3. Not a security vulnerability scanner (works with existing scanners)
4. Not a breaking change detector (relies on semver and tests)

## User Personas

### Platform Engineer

- Manages CI/CD for 100+ repositories
- Wants to reduce time spent on dependency PR reviews
- Needs org-wide policy enforcement
- Values automation with safety guardrails

### Security Engineer

- Needs rapid deployment of security patches
- Wants visibility into dependency update status
- Requires audit trail for compliance
- Values policy-based automation

### Open Source Maintainer

- Maintains multiple related projects
- Wants consistent dependency versions across ecosystem
- Needs automated maintenance releases
- Values time savings on routine tasks

## Features

### P0: Must Have (v0.1.0)

#### Dependency PR Discovery

- Scan repositories across multiple GitHub organizations
- Detect Renovate PRs (author: `renovate[bot]`, `renovate-bot`)
- Detect Dependabot PRs (author: `dependabot[bot]`)
- Extract dependency metadata: name, from/to versions, update type
- Filter by update type (major, minor, patch)

#### Policy Engine

- Cedar policy-as-code for merge decisions
- Built-in merge profiles: aggressive, balanced, conservative
- Policy context includes: PR age, test status, update type, ecosystem
- Dry-run mode for policy evaluation

#### PR Operations

- Auto-review: Add approval review based on policy
- Auto-merge: Merge approved PRs with configurable strategy
- Support squash, merge, and rebase strategies
- Respect branch protection rules

#### Reporting

- JSON output for automation
- Markdown output for human review
- Summary statistics by org, repo, update type

### P1: Should Have (v0.2.0)

#### Maintenance Releases

- Detect repos with merged dependency PRs since last release
- Bump patch version (v1.2.3 → v1.2.4)
- Generate changelog from merged PRs
- Create GitHub release with tag

#### Advanced Policies

- Policy testing and validation
- Custom policy loading from repository
- Policy inheritance and composition

### P2: Nice to Have (v0.3.0)

#### Notifications

- Slack integration for merge/release actions
- Email digests for dependency update status
- Webhook support for custom integrations

#### Dashboard

- Web UI for dependency update status
- Trend analysis over time
- Policy compliance reporting

## Success Metrics

| Metric | Target |
|--------|--------|
| Time to merge safe patches | < 24 hours (vs days) |
| Manual review reduction | 80% for patch updates |
| Release lag after dependency merge | < 1 week |
| Policy compliance rate | > 95% |

## Technical Requirements

See [TRD.md](TRD.md) for technical implementation details.

## Market Context

See [MRD.md](MRD.md) for market analysis and competitive landscape.

## Timeline

| Phase | Features | Target |
|-------|----------|--------|
| v0.1.0 | Scan, Review, Merge, Basic Reporting | Q1 2026 |
| v0.2.0 | Maintenance Releases, Policy Testing | Q2 2026 |
| v0.3.0 | Notifications, Dashboard | Q3 2026 |

## Related Projects

- [PipelineConductor](https://github.com/grokify/pipelineconductor): CI/CD pipeline compliance scanning
- [gogithub](https://github.com/grokify/gogithub): GitHub API helper library
