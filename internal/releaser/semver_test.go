package releaser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPre    string
		wantBuild  string
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "simple",
			input:      "1.2.3",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantPrefix: "",
		},
		{
			name:       "with v prefix",
			input:      "v1.2.3",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantPrefix: "v",
		},
		{
			name:       "with prerelease",
			input:      "v1.2.3-alpha.1",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantPre:    "alpha.1",
			wantPrefix: "v",
		},
		{
			name:       "with build",
			input:      "v1.2.3+build.123",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantBuild:  "build.123",
			wantPrefix: "v",
		},
		{
			name:       "with prerelease and build",
			input:      "v1.2.3-beta+build",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantPre:    "beta",
			wantBuild:  "build",
			wantPrefix: "v",
		},
		{
			name:       "major only",
			input:      "1",
			wantMajor:  1,
			wantPrefix: "",
		},
		{
			name:       "major.minor only",
			input:      "1.2",
			wantMajor:  1,
			wantMinor:  2,
			wantPrefix: "",
		},
		{
			name:       "multi-digit",
			input:      "v10.20.30",
			wantMajor:  10,
			wantMinor:  20,
			wantPatch:  30,
			wantPrefix: "v",
		},
		{
			name:    "invalid major",
			input:   "vX.2.3",
			wantErr: true,
		},
		{
			name:    "invalid minor",
			input:   "v1.X.3",
			wantErr: true,
		},
		{
			name:    "invalid patch",
			input:   "v1.2.X",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Major != tt.wantMajor {
				t.Errorf("Major = %d, want %d", got.Major, tt.wantMajor)
			}
			if got.Minor != tt.wantMinor {
				t.Errorf("Minor = %d, want %d", got.Minor, tt.wantMinor)
			}
			if got.Patch != tt.wantPatch {
				t.Errorf("Patch = %d, want %d", got.Patch, tt.wantPatch)
			}
			if got.Prerelease != tt.wantPre {
				t.Errorf("Prerelease = %q, want %q", got.Prerelease, tt.wantPre)
			}
			if got.Build != tt.wantBuild {
				t.Errorf("Build = %q, want %q", got.Build, tt.wantBuild)
			}
			if got.Prefix != tt.wantPrefix {
				t.Errorf("Prefix = %q, want %q", got.Prefix, tt.wantPrefix)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "simple",
			version: Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "with prefix",
			version: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v"},
			want:    "v1.2.3",
		},
		{
			name:    "with prerelease",
			version: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v", Prerelease: "alpha"},
			want:    "v1.2.3-alpha",
		},
		{
			name:    "with build",
			version: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v", Build: "123"},
			want:    "v1.2.3+123",
		},
		{
			name:    "with prerelease and build",
			version: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v", Prerelease: "beta", Build: "456"},
			want:    "v1.2.3-beta+456",
		},
		{
			name:    "zeros",
			version: Version{Major: 0, Minor: 0, Patch: 0, Prefix: "v"},
			want:    "v0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_BumpMajor(t *testing.T) {
	tests := []struct {
		name  string
		input Version
		want  string
	}{
		{
			name:  "simple",
			input: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v"},
			want:  "v2.0.0",
		},
		{
			name:  "from zero",
			input: Version{Major: 0, Minor: 5, Patch: 10, Prefix: "v"},
			want:  "v1.0.0",
		},
		{
			name:  "no prefix",
			input: Version{Major: 1, Minor: 2, Patch: 3},
			want:  "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.BumpMajor().String()
			if got != tt.want {
				t.Errorf("BumpMajor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_BumpMinor(t *testing.T) {
	tests := []struct {
		name  string
		input Version
		want  string
	}{
		{
			name:  "simple",
			input: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v"},
			want:  "v1.3.0",
		},
		{
			name:  "from zero",
			input: Version{Major: 0, Minor: 0, Patch: 5, Prefix: "v"},
			want:  "v0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.BumpMinor().String()
			if got != tt.want {
				t.Errorf("BumpMinor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_BumpPatch(t *testing.T) {
	tests := []struct {
		name  string
		input Version
		want  string
	}{
		{
			name:  "simple",
			input: Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v"},
			want:  "v1.2.4",
		},
		{
			name:  "from zero",
			input: Version{Major: 1, Minor: 0, Patch: 0, Prefix: "v"},
			want:  "v1.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.BumpPatch().String()
			if got != tt.want {
				t.Errorf("BumpPatch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want int
	}{
		{
			name: "equal",
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "major greater",
			v1:   Version{Major: 2, Minor: 0, Patch: 0},
			v2:   Version{Major: 1, Minor: 9, Patch: 9},
			want: 1,
		},
		{
			name: "major less",
			v1:   Version{Major: 1, Minor: 9, Patch: 9},
			v2:   Version{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "minor greater",
			v1:   Version{Major: 1, Minor: 3, Patch: 0},
			v2:   Version{Major: 1, Minor: 2, Patch: 9},
			want: 1,
		},
		{
			name: "minor less",
			v1:   Version{Major: 1, Minor: 2, Patch: 9},
			v2:   Version{Major: 1, Minor: 3, Patch: 0},
			want: -1,
		},
		{
			name: "patch greater",
			v1:   Version{Major: 1, Minor: 2, Patch: 4},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			name: "patch less",
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 4},
			want: -1,
		},
		{
			name: "prerelease less than release",
			v1:   Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
			v2:   Version{Major: 1, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "release greater than prerelease",
			v1:   Version{Major: 1, Minor: 0, Patch: 0},
			v2:   Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
			want: 1,
		},
		{
			name: "prerelease comparison",
			v1:   Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
			v2:   Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta"},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Compare(&tt.v2)
			if got != tt.want {
				t.Errorf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsSemver(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple", "1.2.3", true},
		{"with v prefix", "v1.2.3", true},
		{"with prerelease", "v1.2.3-alpha", true},
		{"with build", "v1.2.3+build", true},
		{"with prerelease and build", "v1.2.3-alpha+build", true},
		{"complex prerelease", "v1.2.3-alpha.1.beta.2", true},
		{"major only", "1", false},
		{"major.minor", "1.2", false},
		{"invalid format", "v1.2.3.4", false},
		{"non-numeric", "vX.Y.Z", false},
		{"empty", "", false},
		{"just v", "v", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSemver(tt.input)
			if got != tt.want {
				t.Errorf("IsSemver(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindLatestVersion(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "simple list",
			tags: []string{"v1.0.0", "v1.1.0", "v1.2.0"},
			want: "v1.2.0",
		},
		{
			name: "unsorted",
			tags: []string{"v1.2.0", "v0.9.0", "v2.0.0", "v1.5.0"},
			want: "v2.0.0",
		},
		{
			name: "with non-semver tags",
			tags: []string{"latest", "v1.0.0", "release-2023", "v2.0.0"},
			want: "v2.0.0",
		},
		{
			name: "prerelease same version lower",
			tags: []string{"v2.0.0-alpha", "v2.0.0"},
			want: "v2.0.0", // Release > prerelease of same version
		},
		{
			name: "empty list",
			tags: []string{},
			want: "",
		},
		{
			name: "no semver tags",
			tags: []string{"latest", "main", "release"},
			want: "",
		},
		{
			name: "mixed prefixes",
			tags: []string{"1.0.0", "v2.0.0", "v1.5.0"},
			want: "v2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLatestVersion(tt.tags)
			if got != tt.want {
				t.Errorf("FindLatestVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextPatchVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
		wantErr bool
	}{
		{"simple", "v1.2.3", "v1.2.4", false},
		{"no prefix", "1.2.3", "1.2.4", false},
		{"from zero", "v1.0.0", "v1.0.1", false},
		{"invalid", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextPatchVersion(tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextPatchVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NextPatchVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextMinorVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
		wantErr bool
	}{
		{"simple", "v1.2.3", "v1.3.0", false},
		{"no prefix", "1.2.3", "1.3.0", false},
		{"from zero", "v1.0.0", "v1.1.0", false},
		{"invalid", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextMinorVersion(tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextMinorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NextMinorVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextMajorVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
		wantErr bool
	}{
		{"simple", "v1.2.3", "v2.0.0", false},
		{"no prefix", "1.2.3", "2.0.0", false},
		{"from zero", "v0.9.9", "v1.0.0", false},
		{"invalid", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextMajorVersion(tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextMajorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NextMajorVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParse_RoundTrip(t *testing.T) {
	versions := []string{
		"v1.2.3",
		"1.2.3",
		"v0.0.0",
		"v10.20.30",
		"v1.2.3-alpha",
		"v1.2.3+build",
		"v1.2.3-alpha+build",
	}

	for _, v := range versions {
		t.Run(v, func(t *testing.T) {
			parsed, err := Parse(v)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", v, err)
			}
			got := parsed.String()
			if got != v {
				t.Errorf("RoundTrip failed: Parse(%q).String() = %q", v, got)
			}
		})
	}
}
