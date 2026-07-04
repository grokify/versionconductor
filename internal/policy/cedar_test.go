package policy

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/plexusone/versionconductor/pkg/model"
)

func TestCedarEngine_LoadPolicies(t *testing.T) {
	// Find the policies directory relative to this test
	policiesDir := filepath.Join("..", "..", "policies")

	engine, err := NewCedarEngine(policiesDir, "quarantine")
	if err != nil {
		t.Fatalf("NewCedarEngine failed: %v", err)
	}

	count := engine.PolicyCount()
	if count == 0 {
		t.Error("Expected policies to be loaded, got 0")
	}

	t.Logf("Loaded %d policies", count)
}

func TestCedarEngine_EvaluateMerge(t *testing.T) {
	policiesDir := filepath.Join("..", "..", "policies")

	engine, err := NewCedarEngine(policiesDir, "quarantine")
	if err != nil {
		t.Fatalf("NewCedarEngine failed: %v", err)
	}

	tests := []struct {
		name     string
		ctx      *model.PolicyContext
		expected bool
	}{
		{
			name: "should permit patch update from dependabot after 5 days",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         123,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144, // 6 days
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expected: true,
		},
		{
			name: "should deny major update",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         124,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 2.0.0",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144,
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "2.0.0",
					UpdateType:  "major",
					IsMajor:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expected: false,
		},
		{
			name: "should deny when replace directive changes",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         125,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144,
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasReplaceChange:       true,
					HasDirectiveChanges:    true,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expected: false,
		},
		{
			name: "should deny PR younger than 5 days",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         126,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       48, // 2 days
					AgeDays:        2,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expected: false,
		},
		{
			name: "should deny when files other than go.mod/go.sum changed",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         127,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144,
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: false, // Other files changed!
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := engine.CanMerge(context.Background(), tt.ctx)
			if err != nil {
				t.Fatalf("CanMerge failed: %v", err)
			}

			if decision.Allowed != tt.expected {
				t.Errorf("Expected Allowed=%v, got %v. Reasons: %v, Policies: %v",
					tt.expected, decision.Allowed, decision.Reasons, decision.Policies)
			}
		})
	}
}

func TestCedarEngine_DecisionFields(t *testing.T) {
	policiesDir := filepath.Join("..", "..", "policies")

	engine, err := NewCedarEngine(policiesDir, "quarantine")
	if err != nil {
		t.Fatalf("NewCedarEngine failed: %v", err)
	}

	tests := []struct {
		name            string
		ctx             *model.PolicyContext
		expectedOutcome model.DecisionOutcome
		checkActions    bool
		checkEvidence   bool
	}{
		{
			name: "approved PR should have AUTO_MERGE outcome",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         123,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144,
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expectedOutcome: model.DecisionAutoMerge,
			checkActions:    true,
			checkEvidence:   true,
		},
		{
			name: "replace directive should trigger SECURITY_TEAM_REVIEW",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         125,
					Title:          "Bump with replace",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       144,
					AgeDays:        6,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasReplaceChange:       true,
					HasDirectiveChanges:    true,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expectedOutcome: model.DecisionSecurityReview,
			checkActions:    true,
			checkEvidence:   true,
		},
		{
			name: "young PR should trigger QUEUE_FOR_MERGE",
			ctx: &model.PolicyContext{
				Repo: model.RepoContext{
					Owner:    "plexusone",
					Name:     "versionconductor",
					FullName: "plexusone/versionconductor",
				},
				PR: model.PRContext{
					Number:         126,
					Title:          "Bump github.com/foo/bar from 1.2.3 to 1.2.4",
					Author:         "dependabot[bot]",
					IsDependency:   true,
					DependBot:      "dependabot",
					AgeHours:       48,
					AgeDays:        2,
					Mergeable:      true,
					Draft:          false,
					OnlyGoModFiles: true,
				},
				Dependency: model.DependencyContext{
					Name:        "github.com/foo/bar",
					Ecosystem:   "go",
					FromVersion: "1.2.3",
					ToVersion:   "1.2.4",
					UpdateType:  "patch",
					IsPatch:     true,
				},
				CI: model.CIContext{
					AllPassed: true,
				},
				GoMod: model.GoModContext{
					HasDirectiveChanges:    false,
					HasNewDirectDependency: false,
				},
				Profile: model.ProfileContext{
					Name:        "quarantine",
					MinAgeHours: 120,
					MinAgeDays:  5,
				},
			},
			expectedOutcome: model.DecisionQueueForMerge,
			checkActions:    true,
			checkEvidence:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := engine.CanMerge(context.Background(), tt.ctx)
			if err != nil {
				t.Fatalf("CanMerge failed: %v", err)
			}

			// Check outcome
			if decision.Outcome != tt.expectedOutcome {
				t.Errorf("Expected Outcome=%v, got %v", tt.expectedOutcome, decision.Outcome)
			}

			// Check that RequiredActions is populated
			if tt.checkActions && len(decision.RequiredActions) == 0 {
				t.Error("Expected RequiredActions to be populated")
			}

			// Check that Evidence is populated
			if tt.checkEvidence {
				if decision.Evidence == nil {
					t.Error("Expected Evidence to be populated")
				} else {
					if decision.Evidence.Author == "" {
						t.Error("Expected Evidence.Author to be populated")
					}
					if decision.Evidence.UpdateType == "" {
						t.Error("Expected Evidence.UpdateType to be populated")
					}
				}
			}

			t.Logf("Decision: Outcome=%s, Actions=%v, Policies=%v",
				decision.Outcome, decision.RequiredActions, decision.Policies)
		})
	}
}
