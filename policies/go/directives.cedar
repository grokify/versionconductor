// Directives Policy Bundle
//
// Controls go.mod directive changes that are allowed in auto-merged PRs.
// Dangerous directives like replace, exclude, toolchain require manual review.

// ============================================================================
// PERMIT: Allow PRs without directive changes
// ============================================================================

@id("allow-no-directive-changes")
@action("AUTO_MERGE")
@description("PRs without go.mod directive changes are safe")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasDirectiveChanges == false &&
    context.goMod.hasNewDirectDependency == false
};

// ============================================================================
// FORBID: Block replace directive changes
// ============================================================================

@id("forbid-replace-directive")
@action("SECURITY_TEAM_REVIEW")
@description("Replace directives can redirect to malicious forks")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasReplaceChange == true
};

// ============================================================================
// FORBID: Block exclude directive changes
// ============================================================================

@id("forbid-exclude-directive")
@action("MANUAL_REVIEW")
@description("Exclude directives can cause unexpected behavior")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasExcludeChange == true
};

// ============================================================================
// FORBID: Block retract directive changes
// ============================================================================

@id("forbid-retract-directive")
@action("MANUAL_REVIEW")
@description("Retract directives indicate known issues with versions")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasRetractChange == true
};

// ============================================================================
// FORBID: Block toolchain changes
// ============================================================================

@id("forbid-toolchain-change")
@action("MANUAL_REVIEW")
@description("Toolchain changes can affect build behavior")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasToolchainChange == true
};

// ============================================================================
// FORBID: Block Go version changes
// ============================================================================

@id("forbid-go-version-change")
@action("MANUAL_REVIEW")
@description("Go version changes can affect compatibility")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasGoVersionChange == true
};

// ============================================================================
// FORBID: Block new direct dependencies
// ============================================================================

@id("forbid-new-dependency")
@action("SECURITY_TEAM_REVIEW")
@description("New dependencies require security and license review")
forbid(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.goMod.hasNewDirectDependency == true
};
