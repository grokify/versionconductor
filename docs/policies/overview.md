# Cedar Policy Overview

VersionConductor uses [Cedar](https://www.cedarpolicy.com/), an open-source policy language developed by AWS, for fine-grained authorization decisions.

## Why Cedar?

Cedar provides several advantages for dependency management policies:

- **Declarative** - Policies are readable and auditable
- **Analyzable** - Cedar policies can be formally verified
- **Separable** - Policies can be managed independently from code
- **Expressive** - Complex conditions can be expressed clearly

## How It Works

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  PR Data    │────>│   Policy    │────>│  Decision   │
│  (Context)  │     │   Engine    │     │  + Outcome  │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    │   Cedar   │
                    │  Policies │
                    └───────────┘
```

1. VersionConductor collects PR metadata (author, files, CI status, etc.)
2. The policy engine builds a Cedar context from this data
3. Cedar policies are evaluated against the context
4. A decision is returned with outcome, reasons, and required actions

## Decision Outcomes

Each policy evaluation returns a decision outcome:

| Outcome | Description |
|---------|-------------|
| `AUTO_MERGE` | PR can be approved and auto-merged |
| `AUTO_APPROVE` | PR can be approved but not auto-merged |
| `QUEUE_FOR_MERGE` | PR should wait (e.g., for age requirement) |
| `MANUAL_REVIEW` | PR requires human review |
| `SECURITY_TEAM_REVIEW` | PR requires security team review |
| `REJECT` | PR should not be merged |

Policies use the `@action` annotation to specify their outcome:

```cedar
@id("forbid-replace-directive")
@action("SECURITY_TEAM_REVIEW")
forbid(principal, action == Action::"merge", resource)
when { context.goMod.hasReplaceChange == true };
```

## Policy Structure

Cedar policies have two types:

### Permit Policies

Allow an action when conditions are met:

```cedar
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isPatch == true &&
    context.ci.allPassed == true
};
```

### Forbid Policies

Block an action regardless of permit policies:

```cedar
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isMajor == true
};
```

!!! note "Forbid Takes Precedence"
    If any `forbid` policy matches, the action is denied even if `permit`
    policies also match. This provides a safe way to create blocklist rules.

## Built-in Policies

VersionConductor includes policy bundles organized by concern:

```
policies/go/
├── authors.cedar      # Bot allowlist
├── dependency.cedar   # Patch/minor/major rules
├── directives.cedar   # go.mod directive checks
├── files.cedar        # File gate (only go.mod/go.sum)
├── security.cedar     # CI requirements
├── state.cedar        # Mergeable state checks
└── timing.cedar       # Age/quarantine requirements
```

Each bundle focuses on one aspect of the decision:

| Bundle | Purpose |
|--------|---------|
| `authors.cedar` | Only allow trusted bots (dependabot, renovate) |
| `dependency.cedar` | Allow patch/minor, block major updates |
| `directives.cedar` | Block dangerous go.mod directives |
| `files.cedar` | Only allow go.mod/go.sum changes |
| `security.cedar` | Require CI checks to pass |
| `state.cedar` | Require PR to be mergeable, not draft |
| `timing.cedar` | Enforce N+5 quarantine period |

See [Policy Gates](gates.md) for details on each gate.

## Using Custom Policies

Load policies from a file or directory:

```bash
versionconductor review --orgs myorg --policy ./my-policies/
```

Or specify in config:

```yaml
policy:
  path: ./policies/
```

## Next Steps

- [Policy Gates](gates.md) - Understand each safety gate
- [Writing Policies](writing.md) - Create your own policies
- [Context Reference](context.md) - Available context fields
