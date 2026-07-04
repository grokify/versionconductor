package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/plexusone/versionconductor/pkg/model"
)

// CommentMarker is used to identify VersionConductor comments for updates.
const CommentMarker = "<!-- versionconductor-decision -->"

// FormatDecisionComment formats a PolicyDecision as a markdown comment for PR audit trail.
func FormatDecisionComment(decision *model.PolicyDecision, profileName string) string {
	var sb strings.Builder

	// Hidden marker for identifying our comments
	sb.WriteString(CommentMarker)
	sb.WriteString("\n\n")

	// Header with outcome
	sb.WriteString("## VersionConductor Policy Decision\n\n")

	// Status emoji and outcome
	if decision.Allowed {
		sb.WriteString(fmt.Sprintf("**Decision:** :white_check_mark: **%s**\n\n", decision.Outcome))
	} else {
		sb.WriteString(fmt.Sprintf("**Decision:** :x: **%s**\n\n", decision.Outcome))
	}

	// Required actions
	if len(decision.RequiredActions) > 0 {
		sb.WriteString("### Required Actions\n\n")
		for _, action := range decision.RequiredActions {
			sb.WriteString(fmt.Sprintf("- `%s`\n", action))
		}
		sb.WriteString("\n")
	}

	// Reasons
	if len(decision.Reasons) > 0 {
		sb.WriteString("### Evaluation Results\n\n")
		for _, reason := range decision.Reasons {
			sb.WriteString(fmt.Sprintf("- %s\n", reason))
		}
		sb.WriteString("\n")
	}

	// Evidence summary
	if decision.Evidence != nil {
		sb.WriteString("### Evidence Summary\n\n")
		sb.WriteString("| Property | Value |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Author | `%s` |\n", decision.Evidence.Author))
		sb.WriteString(fmt.Sprintf("| PR Age | %s |\n", decision.Evidence.PRAge))
		sb.WriteString(fmt.Sprintf("| Update Type | %s |\n", decision.Evidence.UpdateType))
		sb.WriteString(fmt.Sprintf("| Ecosystem | %s |\n", decision.Evidence.Ecosystem))
		sb.WriteString(fmt.Sprintf("| Files Changed | %s |\n", decision.Evidence.FilesChanged))
		sb.WriteString(fmt.Sprintf("| CI Status | %s |\n", decision.Evidence.CIStatus))
		sb.WriteString(fmt.Sprintf("| Directives | %s |\n", decision.Evidence.Directives))
		sb.WriteString("\n")
	}

	// Matching policies (collapsed)
	if len(decision.Policies) > 0 {
		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Matching Policies</summary>\n\n")
		for _, p := range decision.Policies {
			sb.WriteString(fmt.Sprintf("- `%s`\n", p))
		}
		sb.WriteString("\n</details>\n\n")
	}

	// Footer
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("*Evaluated at %s using profile `%s`*\n",
		time.Now().UTC().Format(time.RFC3339), profileName))
	sb.WriteString("*Powered by [VersionConductor](https://github.com/plexusone/versionconductor)*\n")

	return sb.String()
}

// FormatDecisionJSON formats a PolicyDecision as JSON for machine consumption.
// This is useful for GitHub Action outputs.
func FormatDecisionJSON(decision *model.PolicyDecision) (string, error) {
	// This is handled by json.Marshal in the command layer
	return decision.Summary(), nil
}

// OutcomeEmoji returns an appropriate emoji for the decision outcome.
func OutcomeEmoji(outcome model.DecisionOutcome) string {
	switch outcome {
	case model.DecisionAutoMerge:
		return ":rocket:"
	case model.DecisionAutoApprove:
		return ":white_check_mark:"
	case model.DecisionQueueForMerge:
		return ":hourglass:"
	case model.DecisionManualReview:
		return ":eyes:"
	case model.DecisionSecurityReview:
		return ":shield:"
	case model.DecisionReject:
		return ":no_entry:"
	default:
		return ":question:"
	}
}
