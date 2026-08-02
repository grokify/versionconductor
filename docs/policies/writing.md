# Writing Cedar Policies

This guide shows how to write custom Cedar policies for VersionConductor.

## Policy Basics

### Permit Policy

```cedar
permit(
    principal,           // The bot or user
    action,              // The action (merge, review)
    resource             // The PR
)
when {
    // Conditions that must be true
};
```

### Forbid Policy

```cedar
forbid(
    principal,
    action,
    resource
)
when {
    // Conditions that trigger denial
};
```

## Common Patterns

### Allow Patch Updates Only

```cedar
@id("patch-only")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isPatch == true &&
    context.ci.allPassed == true &&
    context.pr.mergeable == true
};

@id("forbid-non-patch")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isPatch == false
};
```

### Allow Specific Dependencies to Skip Age Check

```cedar
@id("trusted-fast-merge")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.name == "golang.org/x/crypto" &&
    context.ci.allPassed == true &&
    context.pr.onlyGoModFiles == true
};
```

### Different Rules by Repository

```cedar
@id("frontend-rules")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.repo.name == "web-frontend" &&
    context.ci.allPassed == true
    // Frontend repos might have different rules
};

@id("backend-rules")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.repo.name == "api-server" &&
    context.pr.ageDays >= 7 &&  // Stricter for backend
    context.ci.allPassed == true
};
```

### Block PRs to Archived Repos

```cedar
@id("block-archived")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.repo.archived == true
};
```

### Require Longer Quarantine for Private Repos

```cedar
@id("private-repo-quarantine")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.repo.private == true &&
    context.pr.ageDays >= 10 &&  // 10 days for private repos
    context.ci.allPassed == true
};
```

## Operators

Cedar supports these operators:

| Operator | Description |
|----------|-------------|
| `==` | Equals |
| `!=` | Not equals |
| `<`, `>`, `<=`, `>=` | Comparison |
| `&&` | Logical AND |
| `\|\|` | Logical OR |
| `!` | Logical NOT |
| `in` | Membership |
| `like` | Pattern matching |

## Policy Annotations

Use annotations to identify policies and specify outcomes:

```cedar
@id("my-policy-name")
@action("AUTO_MERGE")
@description("Human-readable description")
permit(
    principal,
    action,
    resource
)
when { ... };
```

### Available Annotations

| Annotation | Description |
|------------|-------------|
| `@id` | Unique identifier for the policy |
| `@action` | Decision outcome when policy matches |
| `@description` | Human-readable explanation |

### Decision Outcomes

The `@action` annotation specifies what happens when the policy matches:

| Outcome | Description | Use Case |
|---------|-------------|----------|
| `AUTO_MERGE` | Approve and enable auto-merge | Safe dependency updates |
| `AUTO_APPROVE` | Approve but don't auto-merge | Updates that need manual merge |
| `QUEUE_FOR_MERGE` | Queue for later merge | PRs not yet ready (age, etc.) |
| `MANUAL_REVIEW` | Request human review | Minor concerns |
| `SECURITY_TEAM_REVIEW` | Request security team | Dangerous changes |
| `REJECT` | Block the merge | Policy violations |

### Example with Outcomes

```cedar
@id("allow-patch")
@action("AUTO_MERGE")
@description("Patch updates are safe to auto-merge")
permit(principal, action == Action::"merge", resource)
when { context.dependency.isPatch == true };

@id("forbid-replace")
@action("SECURITY_TEAM_REVIEW")
@description("Replace directives need security review")
forbid(principal, action == Action::"merge", resource)
when { context.goMod.hasReplaceChange == true };

@id("forbid-young")
@action("QUEUE_FOR_MERGE")
@description("PRs must wait 5 days")
forbid(principal, action == Action::"merge", resource)
when { context.pr.ageDays < 5 };
```

## Loading Custom Policies

Cedar policies are evaluated via `policy evaluate` (`review`/`merge` use built-in merge profiles instead):

```bash
versionconductor policy evaluate --repo myorg/service-api --pr 123 --policies ./my-policies/
```

### Directory Structure

Organize policies by concern for maintainability:

```
policies/
├── go/
│   ├── authors.cedar      # Bot allowlist
│   ├── dependency.cedar   # Version rules
│   ├── directives.cedar   # go.mod checks
│   ├── files.cedar        # File gate
│   ├── security.cedar     # CI requirements
│   ├── state.cedar        # Mergeable checks
│   └── timing.cedar       # Age requirements
└── examples/
    └── custom.cedar       # Custom rules
```

All `.cedar` files in the directory (and subdirectories) are loaded.

## Testing Policies

### Using the CLI

Test policies against a real PR:

```bash
versionconductor policy evaluate \
  --repo myorg/myrepo \
  --pr 123 \
  --policies ./my-policies/ \
  --verbose
```

Or test with JSON input (no GitHub access needed):

```bash
echo '{
  "pr": {"author": "dependabot[bot]", "ageDays": 3},
  "dependency": {"isPatch": true},
  "ci": {"allPassed": true}
}' | versionconductor policy evaluate --stdin --policies ./my-policies/
```

### Writing Go Tests

```go
func TestCustomPolicy(t *testing.T) {
    engine, err := policy.NewCedarEngine("./my-policies/", "quarantine")
    if err != nil {
        t.Fatal(err)
    }

    ctx := &model.PolicyContext{
        PR: model.PRContext{
            Author: "dependabot[bot]",
            AgeDays: 3,
        },
        // ... other context
    }

    decision, err := engine.CanMerge(context.Background(), ctx)
    if decision.Allowed {
        t.Error("Expected denial for 3-day-old PR")
    }
    if decision.Outcome != model.DecisionQueueForMerge {
        t.Errorf("Expected QUEUE_FOR_MERGE, got %s", decision.Outcome)
    }
}
```

## Debugging

Enable verbose logging to see policy evaluation:

```bash
versionconductor review --orgs myorg --verbose
```

The diagnostic output shows:

- Which policies matched
- Which policies denied
- Evaluation errors

## Best Practices

1. **Use forbid for blocklists** - Forbid policies override permits
2. **Add @id annotations** - Makes debugging easier
3. **Keep policies focused** - One rule per policy
4. **Test edge cases** - Especially boundary conditions
5. **Version control policies** - Treat them like code
