// State Policy Bundle
//
// Controls PR state requirements for auto-merge.
// PRs must be in a mergeable state.

// ============================================================================
// PERMIT: Allow mergeable PRs
// ============================================================================

@id("allow-mergeable-pr")
@action("AUTO_MERGE")
@description("PRs that are mergeable and not drafts can proceed")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.mergeable == true &&
    context.pr.draft == false &&
    context.pr.hasConflicts == false
};

// ============================================================================
// FORBID: Block draft PRs
// ============================================================================

@id("forbid-draft-pr")
@action("MANUAL_REVIEW")
@description("Draft PRs are not ready for merge")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.draft == true
};

// ============================================================================
// FORBID: Block PRs with conflicts
// ============================================================================

@id("forbid-conflicting-pr")
@action("MANUAL_REVIEW")
@description("PRs with merge conflicts need manual resolution")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.hasConflicts == true
};

// ============================================================================
// FORBID: Block non-mergeable PRs
// ============================================================================

@id("forbid-unmergeable-pr")
@action("MANUAL_REVIEW")
@description("PRs that are not mergeable cannot be auto-merged")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.mergeable == false
};
