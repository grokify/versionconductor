// Dependency Policy Bundle
//
// Controls which types of dependency updates can be auto-merged.
// Major updates require manual review.

// ============================================================================
// PERMIT: Allow patch updates
// ============================================================================

@id("allow-patch-updates")
@action("AUTO_MERGE")
@description("Patch updates (1.2.3 -> 1.2.4) are safe to auto-merge")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isPatch == true &&
    context.dependency.ecosystem == "go"
};

// ============================================================================
// PERMIT: Allow minor updates
// ============================================================================

@id("allow-minor-updates")
@action("AUTO_MERGE")
@description("Minor updates (1.2.3 -> 1.3.0) are generally safe to auto-merge")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isMinor == true &&
    context.dependency.ecosystem == "go"
};

// ============================================================================
// FORBID: Block major updates
// ============================================================================

@id("forbid-major-updates")
@action("MANUAL_REVIEW")
@description("Major updates may contain breaking changes and require manual review")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.dependency.isMajor == true
};
