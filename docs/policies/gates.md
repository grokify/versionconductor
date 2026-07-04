# Policy Gates

VersionConductor's default Go dependency policy enforces multiple safety gates. Each gate must pass for a PR to be auto-merged.

## Gate Summary

| Gate | Description | Why It Matters |
|------|-------------|----------------|
| Author | Only trusted bots | Prevents malicious PRs |
| File | Only go.mod/go.sum | Limits scope of changes |
| Directive | No replace/exclude/toolchain | Prevents dangerous configurations |
| Version | Patch/minor only | Major versions need manual review |
| New Dependency | No new dependencies | New deps need security review |
| Age | N+5 day quarantine | Time for community to find issues |
| CI | All checks pass | Ensures code quality |
| Mergeable | Not draft, no conflicts | PR is ready to merge |

## Author Gate

**Allowed authors:**

- `dependabot[bot]`
- `renovate[bot]`

```cedar
(context.pr.author == "dependabot[bot]" ||
 context.pr.author == "renovate[bot]")
```

!!! tip "Self-Hosted Renovate"
    If you use self-hosted Renovate, add your bot's username to the allowed
    authors in your custom policy.

## File Gate

**Allowed files:**

- `go.mod`
- `go.sum`

```cedar
context.pr.onlyGoModFiles == true
```

This prevents:

- Code changes hidden in dependency PRs
- Workflow file modifications
- Configuration changes

## Directive Gate

**Blocked changes:**

| Directive | Risk |
|-----------|------|
| `replace` | Can redirect to malicious forks |
| `exclude` | Can cause unexpected behavior |
| `retract` | Indicates known issues |
| `toolchain` | Can change build behavior |
| `go` version | Can affect compatibility |

```cedar
context.goMod.hasDirectiveChanges == false
```

## Version Gate

**Allowed:**

- Patch updates (1.2.3 → 1.2.4)
- Minor updates (1.2.3 → 1.3.0)

**Blocked:**

- Major updates (1.2.3 → 2.0.0)

```cedar
(context.dependency.isPatch == true || context.dependency.isMinor == true)
```

!!! warning "Major Updates"
    Major version updates often include breaking changes and should be
    reviewed manually. The forbid policy ensures they're never auto-merged.

## New Dependency Gate

New direct dependencies require manual review:

```cedar
context.goMod.hasNewDirectDependency == false
```

This ensures:

- Security review of new dependencies
- License compatibility check
- Maintenance status evaluation

## Age Gate (N+5 Quarantine)

PRs must be open for at least 5 days before auto-merge:

```cedar
context.pr.ageDays >= context.profile.minAgeDays
```

The quarantine period allows:

- Time for security vulnerabilities to be discovered
- Community to report regressions
- Maintainers to issue retractions if needed

!!! info "Manual Override"
    Use `workflow_dispatch` with a specific PR number to bypass the age
    check for urgent security updates.

## CI Gate

All CI checks must pass:

```cedar
context.ci.allPassed == true
```

This includes:

- Unit tests
- Integration tests
- Linting
- Security scanning (govulncheck, CodeQL)

## Mergeable Gate

The PR must be in a mergeable state:

```cedar
context.pr.mergeable == true &&
context.pr.draft == false &&
context.pr.hasConflicts == false
```

## Forbid Policies

Each gate has a corresponding `forbid` policy that explicitly blocks violations:

```cedar
// Example: Forbid major updates
@id("forbid-major-updates")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isMajor == true
};
```

Forbid policies ensure that even if a permit policy has a bug, dangerous PRs are still blocked.

## Customizing Gates

To customize gates, create your own policy file:

```cedar
// my-policy.cedar

// Allow major updates from specific dependencies
@id("allow-trusted-major")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isMajor == true &&
    context.dependency.name == "github.com/trusted/library"
};
```

See [Writing Policies](writing.md) for more examples.
