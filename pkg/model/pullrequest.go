package model

import (
	"strings"
	"time"
)

// PullRequest represents a GitHub pull request, with dependency-specific metadata.
type PullRequest struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Body         string     `json:"body,omitempty"`
	State        string     `json:"state"` // open, closed
	Author       string     `json:"author"`
	HTMLURL      string     `json:"htmlUrl"`
	IsDependency bool       `json:"isDependency"`
	DependBot    DependBot  `json:"dependBot,omitempty"`
	Dependency   Dependency `json:"dependency,omitempty"`
	TestsPassed  bool       `json:"testsPassed"`
	Mergeable    bool       `json:"mergeable"`
	MergeableStr string     `json:"mergeableState,omitempty"`
	Draft        bool       `json:"draft"`
	Labels       []string   `json:"labels,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	MergedAt     *time.Time `json:"mergedAt,omitempty"`
	Repo         RepoRef    `json:"repo"`
}

// DependBot identifies the dependency management bot.
type DependBot string

const (
	DependBotUnknown    DependBot = ""
	DependBotRenovate   DependBot = "renovate"
	DependBotDependabot DependBot = "dependabot"
)

// DetectDependBot determines which dependency bot created a PR based on author.
func DetectDependBot(author string) DependBot {
	lower := strings.ToLower(author)
	switch {
	case strings.Contains(lower, "renovate"):
		return DependBotRenovate
	case strings.Contains(lower, "dependabot"):
		return DependBotDependabot
	default:
		return DependBotUnknown
	}
}

// Dependency represents a dependency update in a PR.
type Dependency struct {
	Name        string     `json:"name"`
	Ecosystem   string     `json:"ecosystem"` // go, npm, pip, maven, etc.
	FromVersion string     `json:"fromVersion"`
	ToVersion   string     `json:"toVersion"`
	UpdateType  UpdateType `json:"updateType"` // major, minor, patch
}

// UpdateType represents the semantic version update type.
type UpdateType string

const (
	UpdateTypeMajor   UpdateType = "major"
	UpdateTypeMinor   UpdateType = "minor"
	UpdateTypePatch   UpdateType = "patch"
	UpdateTypeUnknown UpdateType = "unknown"
)

// AgeHours returns the age of the PR in hours.
func (pr *PullRequest) AgeHours() int {
	return int(time.Since(pr.CreatedAt).Hours())
}

// IsMerged returns true if the PR has been merged.
func (pr *PullRequest) IsMerged() bool {
	return pr.MergedAt != nil
}

// CheckRun represents a CI check run status.
type CheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`     // queued, in_progress, completed
	Conclusion string `json:"conclusion"` // success, failure, neutral, cancelled, skipped, timed_out, action_required
}

// IsSuccess returns true if the check run completed successfully.
func (c CheckRun) IsSuccess() bool {
	return c.Status == "completed" && c.Conclusion == "success"
}

// PRFilter defines criteria for filtering pull requests.
type PRFilter struct {
	State       string       `json:"state,omitempty"`       // open, closed, all
	DependBot   DependBot    `json:"dependBot,omitempty"`   // renovate, dependabot, or empty for all
	UpdateTypes []UpdateType `json:"updateTypes,omitempty"` // major, minor, patch
	MinAgeHours int          `json:"minAgeHours,omitempty"`
	MaxAgeHours int          `json:"maxAgeHours,omitempty"`
	TestsPassed *bool        `json:"testsPassed,omitempty"`
	Mergeable   *bool        `json:"mergeable,omitempty"`
}
