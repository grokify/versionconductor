# VersionConductor - Market Requirements Document

## Executive Summary

VersionConductor addresses the growing challenge of dependency update management at scale. As organizations adopt automated dependency update tools like Renovate and Dependabot, they face a new problem: managing the flood of PRs these tools generate. No widely-adopted open-source solution exists that combines multi-repo PR scanning, policy-based auto-merge, and automated maintenance releases.

## Market Opportunity

### The Problem Space

Organizations using Renovate or Dependabot face a paradox: automation creates more work.

**Key Statistics:**

- Large enterprises manage 500-5,000+ repositories
- Renovate/Dependabot can generate 50-200+ PRs weekly per organization
- 70-80% of dependency updates are low-risk patch updates
- Manual review of each PR takes 5-15 minutes
- Security patches often wait days in PR queues

### The Dependency Update Lifecycle Gap

| Stage | Tools Available | Gap |
|-------|-----------------|-----|
| Detection | Renovate, Dependabot | ✅ Well-served |
| PR Creation | Renovate, Dependabot | ✅ Well-served |
| Review & Merge | **Manual process** | ❌ **Unaddressed** |
| Release Creation | **Manual process** | ❌ **Unaddressed** |
| Upgrade Ordering | **No tooling** | ❌ **Unaddressed** |
| Cascade Tracking | **No tooling** | ❌ **Unaddressed** |

**VersionConductor fills the Review & Merge, Release Creation, Upgrade Ordering, and Cascade Tracking gaps.**

### The Multi-Repo Dependency Challenge

Organizations maintaining 100s of repos across multiple GitHub accounts face an additional challenge: **interdependencies between their own modules**.

**Example Scenario:**

```
mogo (foundational library)
  ↑
gogithub (depends on mogo)
  ↑
pipelineconductor (depends on gogithub, mogo)
  ↑
versionconductor (depends on gogithub, mogo)
```

When `mogo` releases a new version:
1. Which repos need to be updated? (gogithub, pipelineconductor, versionconductor)
2. What order should they be updated? (gogithub first, then the others)
3. After gogithub is updated, do pipelineconductor and versionconductor need another update?

**No existing tool answers these questions across multiple GitHub orgs.**

## Competitive Landscape

### Existing Solutions

#### 1. Renovate Auto-Merge

Renovate has built-in auto-merge capabilities.

**What it solves:**

- Can auto-merge based on update type
- Configurable per-repository

**Limitations:**

- Configuration scattered across repositories
- No central policy management
- No cross-repo visibility
- No maintenance release automation
- No orchestrated graph-wide upgrades
- No automatic release creation after merge

#### 2. Dependabot Auto-Merge Actions

GitHub Actions workflows that auto-merge Dependabot PRs.

**What it solves:**

- Automates merge for specific criteria
- Works within GitHub Actions

**Limitations:**

- Requires workflow in each repository
- No policy-as-code
- No central management
- Limited to Dependabot PRs

#### 3. Mergify

Commercial tool for PR automation.

**What it solves:**

- Powerful rule engine for PR merging
- Works with any PR, not just dependencies

**Limitations:**

- Not open source
- Per-repository configuration
- No maintenance release automation
- Cost scales with repository count

#### 4. Kodiak

Open-source auto-merge bot.

**What it solves:**

- Auto-merge based on labels and status
- GitHub App model

**Limitations:**

- Per-repository configuration
- No dependency-specific features
- No release automation
- No policy-as-code

### Enterprise Practice Summary

| Practice | What It Solves | Limitations |
|----------|----------------|-------------|
| Renovate auto-merge | Per-repo automation | No central policy, scattered config |
| Dependabot Actions | GitHub-native automation | Per-repo workflows, no visibility |
| Mergify | Powerful rule engine | Commercial, no release automation |
| Kodiak | Label-based merging | No dependency focus, no releases |
| Manual review | Human judgment | Doesn't scale, PR fatigue |
| Ignore updates | Reduces noise | Security risk, technical debt |

### What Doesn't Exist Today

No open-source framework that:

- ✔ Scans dependency PRs across multiple organizations
- ✔ Applies central policy-as-code for merge decisions
- ✔ Provides org-wide visibility into dependency status
- ✔ Orchestrates scheduled graph-wide upgrades in correct order
- ✔ Automatically creates releases after merging, propagating to dependents
- ✔ Automatically creates maintenance releases
- ✔ Works with both Renovate and Dependabot

**This combination doesn't exist publicly today.**

## Customer Pain Points

### Pain Point 1: PR Fatigue

> "We enabled Renovate and now we're drowning in PRs. The automation created more work, not less."
> — Platform Engineer

**Impact:**

- Developers ignore dependency PRs
- PRs sit open for weeks
- Security patches delayed

### Pain Point 2: Inconsistent Merge Policies

> "Some teams merge everything automatically, others require manual review for every patch. We have no consistency."
> — Security Engineer

**Impact:**

- Security posture varies by team
- Compliance reporting difficult
- No enforced standards

### Pain Point 3: Release Lag

> "Dependencies get merged but we forget to cut releases. Users are stuck on old versions with known issues."
> — Open Source Maintainer

**Impact:**

- Users don't get fixes
- Support burden increases
- Project appears unmaintained

### Pain Point 4: No Central Visibility

> "I have no idea how many dependency PRs are open across our 200 repos. Could be 10, could be 500."
> — Engineering Manager

**Impact:**

- Can't measure progress
- Can't identify bottlenecks
- Can't report to leadership

### Pain Point 5: Security Patch Delays

> "A critical CVE fix sat in a Dependabot PR for 3 days because no one noticed it among all the other PRs."
> — Security Engineer

**Impact:**

- Extended vulnerability window
- Compliance violations
- Potential breach

### Pain Point 6: Unknown Upgrade Order

> "We have 150 Go modules across 3 GitHub orgs. When we update our core library, we have no idea which repos to update first or which ones depend on it."
> — Platform Engineer

**Impact:**

- Random upgrade order causes churn
- Dependent modules updated before dependencies
- Multiple rounds of updates required
- Wasted CI cycles

### Pain Point 7: Cascade Blindness

> "We released a new version of our utility library but forgot which of our 80 services use it. Three months later, half are still on the old version."
> — Engineering Manager

**Impact:**

- Inconsistent versions across ecosystem
- Security fixes not propagated
- Bug fixes not reaching dependents
- Manual tracking in spreadsheets

## Value Proposition

### How VersionConductor Addresses Pain Points

| Pain Point | How VersionConductor Helps |
|------------|---------------------------|
| PR fatigue | Auto-merge safe updates, reduce manual review |
| Inconsistent policies | Central Cedar policy-as-code |
| Release lag | Automated maintenance releases |
| No visibility | Cross-org scanning and reporting |
| Security delays | Policy-based fast-track for patches |
| Unknown upgrade order | Dependency graph with topological sort |
| Cascade blindness | Walk-up-graph to find all dependents |

### Unique Differentiators

1. **Dependency Graph**: Build and analyze dependency relationships across orgs
2. **Scheduled Graph Upgrades**: Upgrade entire dependency graph in correct order on a schedule
3. **Release Propagation**: Automatically release updated modules, triggering cascading updates
4. **Policy-as-Code**: Cedar policies are testable, versionable, auditable
2. **Cross-Org Scanning**: Single view of dependency PRs across organizations
3. **Maintenance Releases**: Automatic patch releases when dependencies update
4. **Open Source**: No vendor lock-in, community-driven
5. **Bot-Agnostic**: Works with Renovate, Dependabot, or any dependency bot

## Target Market Segments

### Primary: Mid-to-Large Engineering Organizations

- 50-500+ repositories
- Using Renovate or Dependabot
- Platform engineering function
- Compliance requirements

### Secondary: Open Source Maintainers

- Maintain multiple related projects
- Want automated releases
- Limited time for manual work

### Tertiary: Security Teams

- Need visibility into dependency status
- Want policy enforcement
- Require audit trails

## Market Timing

### Why Now?

1. **Renovate/Dependabot Adoption**: These tools are now standard
2. **PR Volume Growth**: More repos = more PRs = more pain
3. **Security Pressure**: Supply chain security is top priority
4. **Policy-as-Code Maturity**: Cedar provides solid foundation
5. **Platform Engineering Rise**: Teams exist to solve this problem

## Success Metrics

### Adoption Metrics

| Metric | Year 1 Target |
|--------|---------------|
| GitHub stars | 300+ |
| Organizations using | 15+ |
| Repos managed | 3,000+ |

### Customer Value Metrics

| Metric | Target |
|--------|--------|
| PR review time reduction | 80% for patches |
| Time to merge security patches | < 24 hours |
| Manual release effort | 90% reduction |

## Conclusion

VersionConductor addresses a real gap in the dependency management lifecycle. Organizations have adopted tools to detect and create dependency update PRs, but the review, merge, and release stages remain largely manual. By providing policy-based automation for these stages, VersionConductor completes the dependency update automation story.
