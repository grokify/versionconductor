// Auto-merge patch dependency updates
//
// This policy automatically merges patch version updates (e.g., 1.2.3 -> 1.2.4)
// when all CI checks pass and the PR has been open for at least 1 hour.

permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    // Only dependency PRs from known bots
    context.pr.isDependency == true &&

    // All CI checks must pass
    context.ci.allPassed == true &&

    // PR must be at least 1 hour old (time for CI to complete)
    context.pr.ageHours >= 1 &&

    // Only patch updates (no breaking changes expected)
    context.dependency.isPatch == true &&

    // PR must be mergeable (no conflicts)
    context.pr.mergeable == true &&

    // Cannot be a draft PR
    context.pr.draft == false
};

// Also allow auto-review for patch updates
permit(
    principal,
    action == Action::"review",
    resource
)
when {
    context.pr.isDependency == true &&
    context.ci.allPassed == true &&
    context.dependency.isPatch == true
};
