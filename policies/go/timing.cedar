// Timing Policy Bundle
//
// Controls the quarantine period (N+5) for dependency updates.
// PRs must be open for a minimum time before auto-merge.

// ============================================================================
// PERMIT: Allow PRs that meet age requirement
// ============================================================================

@id("allow-mature-pr")
@action("AUTO_MERGE")
@description("PRs that have been open long enough pass the quarantine period")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.ageDays >= context.profile.minAgeDays
};

// ============================================================================
// FORBID: Block PRs that are too young
// ============================================================================

@id("forbid-young-pr")
@action("QUEUE_FOR_MERGE")
@description("PRs must wait for quarantine period before auto-merge")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.ageDays < context.profile.minAgeDays
};
