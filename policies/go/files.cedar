// Files Policy Bundle
//
// Controls which file changes are allowed in auto-merged PRs.
// For Go dependencies, only go.mod and go.sum should change.

// ============================================================================
// PERMIT: Allow PRs that only change go.mod/go.sum
// ============================================================================

@id("allow-gomod-only")
@action("AUTO_MERGE")
@description("PRs that only modify go.mod and go.sum are safe dependency updates")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.onlyGoModFiles == true
};

// ============================================================================
// FORBID: Block PRs that change other files
// ============================================================================

@id("forbid-non-gomod-files")
@action("MANUAL_REVIEW")
@description("PRs that modify files outside go.mod/go.sum require manual review")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.onlyGoModFiles == false
};
