# Context Reference

This page documents all context fields available in Cedar policies.

## Context Structure

```cedar
context {
    repo: RepoContext,
    pr: PRContext,
    dependency: DependencyContext,
    ci: CIContext,
    goMod: GoModContext,
    profile: ProfileContext
}
```

## repo (Repository Context)

| Field | Type | Description |
|-------|------|-------------|
| `repo.owner` | String | Repository owner (org or user) |
| `repo.name` | String | Repository name |
| `repo.fullName` | String | Full name (owner/name) |
| `repo.private` | Boolean | True if private repository |
| `repo.archived` | Boolean | True if archived |
| `repo.language` | String | Primary language |
| `repo.isMonorepo` | Boolean | True if monorepo detected |

### Example

```cedar
when {
    context.repo.owner == "myorg" &&
    context.repo.private == false
}
```

## pr (Pull Request Context)

| Field | Type | Description |
|-------|------|-------------|
| `pr.number` | Long | PR number |
| `pr.title` | String | PR title |
| `pr.author` | String | PR author username |
| `pr.isDependency` | Boolean | True if dependency PR |
| `pr.dependBot` | String | Bot name (dependabot, renovate) |
| `pr.ageHours` | Long | Hours since PR opened |
| `pr.ageDays` | Long | Days since PR opened |
| `pr.mergeable` | Boolean | True if mergeable |
| `pr.draft` | Boolean | True if draft PR |
| `pr.hasConflicts` | Boolean | True if has merge conflicts |
| `pr.onlyGoModFiles` | Boolean | True if only go.mod/go.sum changed |

### Example

```cedar
when {
    context.pr.author == "dependabot[bot]" &&
    context.pr.ageDays >= 5 &&
    context.pr.onlyGoModFiles == true
}
```

## dependency (Dependency Context)

| Field | Type | Description |
|-------|------|-------------|
| `dependency.name` | String | Dependency name/path |
| `dependency.ecosystem` | String | Package ecosystem (go, npm, etc.) |
| `dependency.fromVersion` | String | Current version |
| `dependency.toVersion` | String | Target version |
| `dependency.updateType` | String | Update type (major, minor, patch) |
| `dependency.isMajor` | Boolean | True if major update |
| `dependency.isMinor` | Boolean | True if minor update |
| `dependency.isPatch` | Boolean | True if patch update |

### Example

```cedar
when {
    context.dependency.ecosystem == "go" &&
    context.dependency.isPatch == true
}
```

## ci (CI Context)

| Field | Type | Description |
|-------|------|-------------|
| `ci.allPassed` | Boolean | True if all checks passed |
| `ci.anyFailed` | Boolean | True if any check failed |
| `ci.anyPending` | Boolean | True if any check pending |
| `ci.requiredPassed` | Boolean | True if required checks passed |

### Example

```cedar
when {
    context.ci.allPassed == true &&
    context.ci.anyFailed == false
}
```

## goMod (Go Module Context)

| Field | Type | Description |
|-------|------|-------------|
| `goMod.hasReplaceChange` | Boolean | True if replace directive changed |
| `goMod.hasExcludeChange` | Boolean | True if exclude directive changed |
| `goMod.hasRetractChange` | Boolean | True if retract directive changed |
| `goMod.hasToolchainChange` | Boolean | True if toolchain directive changed |
| `goMod.hasGoVersionChange` | Boolean | True if go version changed |
| `goMod.hasDirectiveChanges` | Boolean | True if any directive changed |
| `goMod.hasNewDirectDependency` | Boolean | True if new dependency added |

### Example

```cedar
when {
    context.goMod.hasDirectiveChanges == false &&
    context.goMod.hasNewDirectDependency == false
}
```

## profile (Profile Context)

| Field | Type | Description |
|-------|------|-------------|
| `profile.name` | String | Active profile name |
| `profile.minAgeHours` | Long | Minimum PR age in hours |
| `profile.minAgeDays` | Long | Minimum PR age in days |

### Example

```cedar
when {
    context.pr.ageDays >= context.profile.minAgeDays
}
```

## Full Example

```cedar
@id("go-dependency-automerge")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    // Author check
    (context.pr.author == "dependabot[bot]" ||
     context.pr.author == "renovate[bot]") &&

    // Dependency check
    context.pr.isDependency == true &&
    context.dependency.ecosystem == "go" &&

    // File check
    context.pr.onlyGoModFiles == true &&

    // Directive check
    context.goMod.hasDirectiveChanges == false &&
    context.goMod.hasNewDirectDependency == false &&

    // Version check
    (context.dependency.isPatch == true ||
     context.dependency.isMinor == true) &&

    // Age check
    context.pr.ageDays >= context.profile.minAgeDays &&

    // CI check
    context.ci.allPassed == true &&

    // State check
    context.pr.mergeable == true &&
    context.pr.draft == false &&
    context.pr.hasConflicts == false
};
```
