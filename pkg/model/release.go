package model

import "time"

// Release represents a GitHub release.
type Release struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tagName"`
	Name        string    `json:"name"`
	Body        string    `json:"body,omitempty"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"createdAt"`
	PublishedAt time.Time `json:"publishedAt"`
	HTMLURL     string    `json:"htmlUrl"`
	Repo        RepoRef   `json:"repo"`
}

// Tag represents a Git tag.
type Tag struct {
	Name string  `json:"name"`
	SHA  string  `json:"sha"`
	Repo RepoRef `json:"repo"`
}

// ReleaseRequest contains the information needed to create a new release.
type ReleaseRequest struct {
	Repo            RepoRef `json:"repo"`
	TagName         string  `json:"tagName"`
	TargetCommitish string  `json:"targetCommitish,omitempty"` // Branch or commit SHA
	Name            string  `json:"name"`
	Body            string  `json:"body"`
	Draft           bool    `json:"draft"`
	Prerelease      bool    `json:"prerelease"`
	GenerateNotes   bool    `json:"generateNotes"`
}

// ReleaseCandidate represents a repository that may need a new release.
type ReleaseCandidate struct {
	Repo             Repo          `json:"repo"`
	CurrentVersion   string        `json:"currentVersion"`
	ProposedVersion  string        `json:"proposedVersion"`
	MergedPRs        []PullRequest `json:"mergedPRs"`
	MergedPRCount    int           `json:"mergedPRCount"`
	LastReleaseAt    *time.Time    `json:"lastReleaseAt,omitempty"`
	DaysSinceRelease int           `json:"daysSinceRelease,omitempty"`
	ShouldRelease    bool          `json:"shouldRelease"`
	ReleaseReason    string        `json:"releaseReason,omitempty"`
}
