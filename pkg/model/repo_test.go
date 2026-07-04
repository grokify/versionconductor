package model

import "testing"

func TestRepoRef_FullName(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		repo  string
		want  string
	}{
		{"standard", "owner", "repo", "owner/repo"},
		{"org repo", "plexusone", "versionconductor", "plexusone/versionconductor"},
		{"empty owner", "", "repo", "/repo"},
		{"empty repo", "owner", "", "owner/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RepoRef{Owner: tt.owner, Name: tt.repo}
			got := r.FullName()
			if got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRepoRef(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		wantOwner string
		wantName  string
	}{
		{"standard", "owner/repo", "owner", "repo"},
		{"org repo", "plexusone/versionconductor", "plexusone", "versionconductor"},
		{"no slash", "repo", "", "repo"},
		{"multiple slashes", "owner/repo/extra", "owner", "repo/extra"},
		{"empty", "", "", ""},
		{"just slash", "/", "", ""},
		{"leading slash", "/repo", "", "repo"},
		{"trailing slash", "owner/", "owner", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRepoRef(tt.fullName)
			if got.Owner != tt.wantOwner {
				t.Errorf("ParseRepoRef(%q).Owner = %q, want %q", tt.fullName, got.Owner, tt.wantOwner)
			}
			if got.Name != tt.wantName {
				t.Errorf("ParseRepoRef(%q).Name = %q, want %q", tt.fullName, got.Name, tt.wantName)
			}
		})
	}
}

func TestParseRepoRef_RoundTrip(t *testing.T) {
	tests := []string{
		"owner/repo",
		"plexusone/versionconductor",
		"google/go-github",
	}

	for _, fullName := range tests {
		t.Run(fullName, func(t *testing.T) {
			ref := ParseRepoRef(fullName)
			got := ref.FullName()
			if got != fullName {
				t.Errorf("RoundTrip failed: ParseRepoRef(%q).FullName() = %q", fullName, got)
			}
		})
	}
}
