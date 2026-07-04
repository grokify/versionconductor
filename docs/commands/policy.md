# policy

The `policy` command provides tools for managing and evaluating Cedar policies.

## Subcommands

### policy evaluate

Evaluate Cedar policies against a pull request and get a decision.

```bash
versionconductor policy evaluate --repo owner/repo --pr 123
```

## Synopsis

```
versionconductor policy evaluate [flags]
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--repo` | Repository (owner/repo format) | |
| `--pr` | Pull request number | |
| `--policies` | Path to Cedar policies directory or file | Built-in |
| `--profile` | Merge profile to use | `balanced` |
| `--action` | Action to evaluate: merge, review, release | `merge` |
| `--stdin` | Read policy context from stdin as JSON | `false` |
| `--context-file` | Read policy context from JSON file | |
| `--comment` | Post decision as a comment on the PR | `false` |
| `--update-comment` | Update existing comment instead of creating new | `true` |
| `--format` | Output format: table, json | `table` |
| `--verbose` | Enable verbose output | `false` |

## Examples

### Evaluate a GitHub PR

```bash
versionconductor policy evaluate \
  --repo plexusone/myproject \
  --pr 123 \
  --profile quarantine
```

### Use Custom Policies

```bash
versionconductor policy evaluate \
  --repo plexusone/myproject \
  --pr 123 \
  --policies ./my-policies/
```

### JSON Output for Scripts

```bash
versionconductor policy evaluate \
  --repo plexusone/myproject \
  --pr 123 \
  --format json
```

Output:

```json
{
  "allowed": true,
  "action": "merge",
  "outcome": "AUTO_MERGE",
  "requiredActions": ["approve", "enable_auto_merge"],
  "reasons": ["Patch updates are safe to auto-merge"],
  "policies": ["dependency.cedar:0"],
  "evidence": {
    "author": "dependabot[bot]",
    "prAge": "6 days",
    "updateType": "patch",
    "filesChanged": "go.mod only",
    "ciStatus": "all passed"
  }
}
```

### Post Decision Comment

Post or update a comment on the PR with the decision audit trail:

```bash
versionconductor policy evaluate \
  --repo plexusone/myproject \
  --pr 123 \
  --comment
```

### Evaluate from stdin

Useful for testing policies without GitHub access:

```bash
echo '{
  "repo": {"owner": "test", "name": "test"},
  "pr": {
    "number": 1,
    "author": "dependabot[bot]",
    "ageDays": 6,
    "mergeable": true,
    "onlyGoModFiles": true
  },
  "dependency": {
    "updateType": "patch",
    "isPatch": true
  },
  "ci": {"allPassed": true},
  "goMod": {}
}' | versionconductor policy evaluate --stdin --policies ./policies
```

## Decision Outcomes

The policy engine returns one of these decision outcomes:

| Outcome | Description |
|---------|-------------|
| `AUTO_MERGE` | PR can be approved and auto-merged |
| `AUTO_APPROVE` | PR can be approved but not auto-merged |
| `QUEUE_FOR_MERGE` | PR should be queued for merge when ready |
| `MANUAL_REVIEW` | PR requires human review |
| `SECURITY_TEAM_REVIEW` | PR requires security team review |
| `REJECT` | PR should not be merged |

## Required Actions

Based on the outcome, these actions may be required:

| Action | Description |
|--------|-------------|
| `approve` | Approve the PR |
| `enable_auto_merge` | Enable GitHub auto-merge |
| `add_to_queue` | Add to merge queue |
| `comment` | Post a comment explaining the decision |
| `add_label` | Add a label to the PR |
| `request_review` | Request review from specified team |

## See Also

- [Policy Overview](../policies/overview.md)
- [Writing Policies](../policies/writing.md)
- [Context Reference](../policies/context.md)
