// Authors Policy Bundle
//
// Controls which actors are allowed to have their PRs auto-merged.
// Only trusted dependency bots should be in this list.

// ============================================================================
// PERMIT: Allow PRs from Dependabot
// ============================================================================

@id("allow-dependabot")
@action("AUTO_MERGE")
@description("Allow PRs authored by GitHub Dependabot")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.author == "dependabot[bot]" &&
    context.pr.isDependency == true
};

// ============================================================================
// PERMIT: Allow PRs from Renovate
// ============================================================================

@id("allow-renovate")
@action("AUTO_MERGE")
@description("Allow PRs authored by Renovate Bot")
permit(
    principal,
    action == Action::"merge",
    resource
)
when {
    context.pr.author == "renovate[bot]" &&
    context.pr.isDependency == true
};

// ============================================================================
// PERMIT: Allow PRs from Renovate for review action
// ============================================================================

@id("allow-dependabot-review")
@action("AUTO_APPROVE")
@description("Allow review of Dependabot PRs")
permit(
    principal,
    action == Action::"review",
    resource
)
when {
    context.pr.author == "dependabot[bot]" &&
    context.pr.isDependency == true
};

@id("allow-renovate-review")
@action("AUTO_APPROVE")
@description("Allow review of Renovate PRs")
permit(
    principal,
    action == Action::"review",
    resource
)
when {
    context.pr.author == "renovate[bot]" &&
    context.pr.isDependency == true
};
