package model

import "time"

// ScanResult contains the results of scanning for dependency PRs.
type ScanResult struct {
	Timestamp    time.Time     `json:"timestamp"`
	Orgs         []string      `json:"orgs"`
	ReposScanned int           `json:"reposScanned"`
	PRsFound     int           `json:"prsFound"`
	PRs          []PullRequest `json:"prs"`
	Errors       []ScanError   `json:"errors,omitempty"`
}

// ScanError represents an error encountered during scanning.
type ScanError struct {
	Repo    string `json:"repo"`
	Message string `json:"message"`
}

// MergeResult contains the results of a merge operation.
type MergeResult struct {
	Timestamp    time.Time   `json:"timestamp"`
	DryRun       bool        `json:"dryRun"`
	Merged       []MergedPR  `json:"merged,omitempty"`
	Skipped      []SkippedPR `json:"skipped,omitempty"`
	Failed       []FailedPR  `json:"failed,omitempty"`
	MergedCount  int         `json:"mergedCount"`
	SkippedCount int         `json:"skippedCount"`
	FailedCount  int         `json:"failedCount"`
}

// MergedPR represents a successfully merged PR.
type MergedPR struct {
	PR       PullRequest `json:"pr"`
	MergedBy string      `json:"mergedBy"`
	SHA      string      `json:"sha"`
}

// SkippedPR represents a PR that was skipped during merge.
type SkippedPR struct {
	PR     PullRequest `json:"pr"`
	Reason string      `json:"reason"`
}

// FailedPR represents a PR that failed to merge.
type FailedPR struct {
	PR    PullRequest `json:"pr"`
	Error string      `json:"error"`
}

// ReviewResult contains the results of reviewing PRs.
type ReviewResult struct {
	Timestamp     time.Time     `json:"timestamp"`
	DryRun        bool          `json:"dryRun"`
	Approved      []PullRequest `json:"approved,omitempty"`
	Denied        []DeniedPR    `json:"denied,omitempty"`
	ApprovedCount int           `json:"approvedCount"`
	DeniedCount   int           `json:"deniedCount"`
}

// DeniedPR represents a PR that was denied review approval.
type DeniedPR struct {
	PR     PullRequest `json:"pr"`
	Reason string      `json:"reason"`
}

// ReleaseResult contains the results of creating releases.
type ReleaseResult struct {
	Timestamp    time.Time        `json:"timestamp"`
	DryRun       bool             `json:"dryRun"`
	Created      []CreatedRelease `json:"created,omitempty"`
	Skipped      []SkippedRelease `json:"skipped,omitempty"`
	Failed       []FailedRelease  `json:"failed,omitempty"`
	CreatedCount int              `json:"createdCount"`
	SkippedCount int              `json:"skippedCount"`
	FailedCount  int              `json:"failedCount"`
}

// CreatedRelease represents a successfully created release.
type CreatedRelease struct {
	Repo            RepoRef `json:"repo"`
	Version         string  `json:"version"`
	PreviousVersion string  `json:"previousVersion"`
	ReleaseURL      string  `json:"releaseUrl"`
	PRsMerged       int     `json:"prsMerged"`
}

// SkippedRelease represents a repository that was skipped for release.
type SkippedRelease struct {
	Repo   RepoRef `json:"repo"`
	Reason string  `json:"reason"`
}

// FailedRelease represents a failed release attempt.
type FailedRelease struct {
	Repo  RepoRef `json:"repo"`
	Error string  `json:"error"`
}
