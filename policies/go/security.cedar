// Security Policy Bundle
//
// Controls CI and security check requirements for auto-merge.
// All required checks must pass before auto-merge.

// ============================================================================
// PERMIT: Allow PRs with all checks passing
// ============================================================================

@id("allow-passing-ci")
@action("AUTO_MERGE")
@description("PRs with all CI checks passing are safe to merge")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.ci.allPassed == true
};

// ============================================================================
// FORBID: Block PRs with failing checks
// ============================================================================

@id("forbid-failing-ci")
@action("REJECT")
@description("PRs with failing CI checks cannot be auto-merged")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.ci.anyFailed == true
};

// ============================================================================
// FORBID: Block PRs with pending checks (optional strictness)
// ============================================================================

@id("forbid-pending-ci")
@action("QUEUE_FOR_MERGE")
@description("PRs with pending CI checks should wait")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.ci.anyPending == true &&
    context.ci.allPassed == false
};
