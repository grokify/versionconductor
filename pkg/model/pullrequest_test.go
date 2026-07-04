package model

import (
	"testing"
	"time"
)

func TestDetectDependBot(t *testing.T) {
	tests := []struct {
		name   string
		author string
		want   DependBot
	}{
		{"renovate bot", "renovate[bot]", DependBotRenovate},
		{"dependabot bot", "dependabot[bot]", DependBotDependabot},
		{"renovate uppercase", "Renovate[bot]", DependBotRenovate},
		{"dependabot uppercase", "Dependabot[bot]", DependBotDependabot},
		{"renovate preview", "renovate-preview[bot]", DependBotRenovate},
		{"unknown user", "johnsmith", DependBotUnknown},
		{"empty", "", DependBotUnknown},
		{"github actions", "github-actions[bot]", DependBotUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDependBot(tt.author)
			if got != tt.want {
				t.Errorf("DetectDependBot(%q) = %q, want %q", tt.author, got, tt.want)
			}
		})
	}
}

func TestPullRequest_AgeHours(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		wantMin   int
		wantMax   int
	}{
		{
			name:      "just created",
			createdAt: now,
			wantMin:   0,
			wantMax:   0,
		},
		{
			name:      "1 hour ago",
			createdAt: now.Add(-1 * time.Hour),
			wantMin:   1,
			wantMax:   1,
		},
		{
			name:      "24 hours ago",
			createdAt: now.Add(-24 * time.Hour),
			wantMin:   24,
			wantMax:   24,
		},
		{
			name:      "5 days ago",
			createdAt: now.Add(-5 * 24 * time.Hour),
			wantMin:   120,
			wantMax:   120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{CreatedAt: tt.createdAt}
			got := pr.AgeHours()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("AgeHours() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestPullRequest_IsMerged(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		mergedAt *time.Time
		want     bool
	}{
		{
			name:     "not merged",
			mergedAt: nil,
			want:     false,
		},
		{
			name:     "merged",
			mergedAt: &now,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{MergedAt: tt.mergedAt}
			got := pr.IsMerged()
			if got != tt.want {
				t.Errorf("IsMerged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckRun_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		conclusion string
		want       bool
	}{
		{"completed success", "completed", "success", true},
		{"completed failure", "completed", "failure", false},
		{"completed neutral", "completed", "neutral", false},
		{"completed skipped", "completed", "skipped", false},
		{"completed cancelled", "completed", "cancelled", false},
		{"completed timed_out", "completed", "timed_out", false},
		{"completed action_required", "completed", "action_required", false},
		{"in_progress", "in_progress", "", false},
		{"queued", "queued", "", false},
		{"in_progress success", "in_progress", "success", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CheckRun{
				Status:     tt.status,
				Conclusion: tt.conclusion,
			}
			got := c.IsSuccess()
			if got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
