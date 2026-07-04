package collector

import (
	"testing"

	"github.com/plexusone/versionconductor/pkg/model"
)

func TestParseDependencyFromTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantName string
		wantFrom string
		wantTo   string
		wantType model.UpdateType
	}{
		{
			name:     "dependabot bump with versions",
			title:    "Bump github.com/spf13/cobra from 1.7.0 to 1.8.0",
			wantName: "github.com/spf13/cobra",
			wantFrom: "1.7.0",
			wantTo:   "1.8.0",
			wantType: model.UpdateTypeMinor,
		},
		{
			name:     "renovate update with versions",
			title:    "Update dependency github.com/google/go-github/v60 from v60.0.0 to v61.0.0",
			wantName: "github.com/google/go-github/v60",
			wantFrom: "v60.0.0",
			wantTo:   "v61.0.0",
			wantType: model.UpdateTypeMajor,
		},
		{
			name:     "patch update",
			title:    "Bump golang.org/x/oauth2 from 0.15.0 to 0.15.1",
			wantName: "golang.org/x/oauth2",
			wantFrom: "0.15.0",
			wantTo:   "0.15.1",
			wantType: model.UpdateTypePatch,
		},
		{
			name:     "deps scope format",
			title:    "deps(go): update github.com/spf13/viper to v1.18.0",
			wantName: "github.com/spf13/viper",
			wantFrom: "",
			wantTo:   "v1.18.0",
			wantType: "", // Only one version, so UpdateType not determined
		},
		{
			name:     "upgrade keyword",
			title:    "Upgrade github.com/stretchr/testify from 1.8.4 to 1.9.0",
			wantName: "github.com/stretchr/testify",
			wantFrom: "1.8.4",
			wantTo:   "1.9.0",
			wantType: model.UpdateTypeMinor,
		},
		{
			name:     "no versions",
			title:    "Update some-package",
			wantName: "some-package",
			wantFrom: "",
			wantTo:   "",
			wantType: "", // No versions, so UpdateType not determined
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDependencyFromTitle(tt.title)

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.FromVersion != tt.wantFrom {
				t.Errorf("FromVersion = %q, want %q", got.FromVersion, tt.wantFrom)
			}
			if got.ToVersion != tt.wantTo {
				t.Errorf("ToVersion = %q, want %q", got.ToVersion, tt.wantTo)
			}
			if got.UpdateType != tt.wantType {
				t.Errorf("UpdateType = %q, want %q", got.UpdateType, tt.wantType)
			}
		})
	}
}

func TestDetermineUpdateType(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want model.UpdateType
	}{
		{"major bump", "1.0.0", "2.0.0", model.UpdateTypeMajor},
		{"minor bump", "1.0.0", "1.1.0", model.UpdateTypeMinor},
		{"patch bump", "1.0.0", "1.0.1", model.UpdateTypePatch},
		{"major with v prefix", "v1.0.0", "v2.0.0", model.UpdateTypeMajor},
		{"minor with v prefix", "v1.2.3", "v1.3.0", model.UpdateTypeMinor},
		{"patch with v prefix", "v1.2.3", "v1.2.4", model.UpdateTypePatch},
		{"same version", "1.0.0", "1.0.0", model.UpdateTypeUnknown},
		{"invalid from", "abc", "1.0.0", model.UpdateTypeUnknown},
		{"invalid to", "1.0.0", "xyz", model.UpdateTypeUnknown},
		{"multi-digit major", "10.0.0", "11.0.0", model.UpdateTypeMajor},
		{"multi-digit minor", "1.10.0", "1.11.0", model.UpdateTypeMinor},
		{"multi-digit patch", "1.0.10", "1.0.11", model.UpdateTypePatch},
		{"prerelease suffix", "1.0.0-alpha", "1.0.1-beta", model.UpdateTypePatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineUpdateType(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("determineUpdateType(%q, %q) = %q, want %q", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    []int
	}{
		{"simple", "1.2.3", []int{1, 2, 3}},
		{"with v prefix", "v1.2.3", []int{1, 2, 3}},
		{"multi-digit", "10.20.30", []int{10, 20, 30}},
		{"with prerelease", "1.2.3-alpha", []int{1, 2, 3}},
		{"with build metadata", "1.2.3+build123", []int{1, 2, 3}},
		{"two parts", "1.2", []int{1, 2}},
		{"single part", "1", []int{1}},
		{"leading zeros stripped", "01.02.03", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVersion(tt.version)
			if len(got) != len(tt.want) {
				t.Errorf("parseVersion(%q) length = %d, want %d", tt.version, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseVersion(%q)[%d] = %d, want %d", tt.version, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDetectEcosystem(t *testing.T) {
	tests := []struct {
		name       string
		dependency string
		want       string
	}{
		{"go module github.com", "github.com/spf13/cobra", "go"},
		{"go module golang.org", "golang.org/x/oauth2", "go"},
		{"npm scoped package", "@types/node", "npm"},
		{"npm package with slash", "lodash/fp", "npm"},
		{"unknown", "some-package", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectEcosystem(tt.dependency)
			if got != tt.want {
				t.Errorf("detectEcosystem(%q) = %q, want %q", tt.dependency, got, tt.want)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name        string
		fullName    string
		excludeList []string
		want        bool
	}{
		{"in list", "owner/repo", []string{"owner/repo"}, true},
		{"not in list", "owner/repo", []string{"other/repo"}, false},
		{"empty list", "owner/repo", []string{}, false},
		{"multiple in list", "owner/repo", []string{"a/b", "owner/repo", "c/d"}, true},
		{"case sensitive", "Owner/Repo", []string{"owner/repo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExcluded(tt.fullName, tt.excludeList)
			if got != tt.want {
				t.Errorf("isExcluded(%q, %v) = %v, want %v", tt.fullName, tt.excludeList, got, tt.want)
			}
		})
	}
}

func TestTestsPassed(t *testing.T) {
	tests := []struct {
		name   string
		checks []model.CheckRun
		want   bool
	}{
		{
			name:   "empty checks",
			checks: []model.CheckRun{},
			want:   false,
		},
		{
			name: "all success",
			checks: []model.CheckRun{
				{Name: "test", Status: "completed", Conclusion: "success"},
				{Name: "lint", Status: "completed", Conclusion: "success"},
			},
			want: true,
		},
		{
			name: "one failure",
			checks: []model.CheckRun{
				{Name: "test", Status: "completed", Conclusion: "success"},
				{Name: "lint", Status: "completed", Conclusion: "failure"},
			},
			want: false,
		},
		{
			name: "in progress",
			checks: []model.CheckRun{
				{Name: "test", Status: "in_progress", Conclusion: ""},
			},
			want: false,
		},
		{
			name: "queued",
			checks: []model.CheckRun{
				{Name: "test", Status: "queued", Conclusion: ""},
			},
			want: false,
		},
		{
			name: "skipped counts as failure",
			checks: []model.CheckRun{
				{Name: "test", Status: "completed", Conclusion: "skipped"},
			},
			want: false,
		},
		{
			name: "neutral counts as failure",
			checks: []model.CheckRun{
				{Name: "test", Status: "completed", Conclusion: "neutral"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TestsPassed(tt.checks)
			if got != tt.want {
				t.Errorf("TestsPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}
