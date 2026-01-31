package releaser

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Prefix     string // "v" or empty
}

// Parse parses a version string into a Version struct.
func Parse(v string) (*Version, error) {
	ver := &Version{}

	// Check for 'v' prefix
	if strings.HasPrefix(v, "v") {
		ver.Prefix = "v"
		v = strings.TrimPrefix(v, "v")
	}

	// Split on '+' for build metadata
	if idx := strings.Index(v, "+"); idx >= 0 {
		ver.Build = v[idx+1:]
		v = v[:idx]
	}

	// Split on '-' for prerelease
	if idx := strings.Index(v, "-"); idx >= 0 {
		ver.Prerelease = v[idx+1:]
		v = v[:idx]
	}

	// Parse major.minor.patch
	parts := strings.Split(v, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	var err error

	ver.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	if len(parts) >= 2 {
		ver.Minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	if len(parts) >= 3 {
		ver.Patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return ver, nil
}

// String returns the version as a string.
func (v *Version) String() string {
	s := fmt.Sprintf("%s%d.%d.%d", v.Prefix, v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// BumpMajor increments the major version and resets minor and patch.
func (v *Version) BumpMajor() *Version {
	return &Version{
		Major:  v.Major + 1,
		Minor:  0,
		Patch:  0,
		Prefix: v.Prefix,
	}
}

// BumpMinor increments the minor version and resets patch.
func (v *Version) BumpMinor() *Version {
	return &Version{
		Major:  v.Major,
		Minor:  v.Minor + 1,
		Patch:  0,
		Prefix: v.Prefix,
	}
}

// BumpPatch increments the patch version.
func (v *Version) BumpPatch() *Version {
	return &Version{
		Major:  v.Major,
		Minor:  v.Minor,
		Patch:  v.Patch + 1,
		Prefix: v.Prefix,
	}
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Prerelease versions have lower precedence
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}

	return strings.Compare(v.Prerelease, other.Prerelease)
}

// IsSemver checks if a string is a valid semver tag.
func IsSemver(s string) bool {
	pattern := `^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// FindLatestVersion finds the highest semver version from a list of tags.
func FindLatestVersion(tags []string) string {
	var versions []*Version

	for _, tag := range tags {
		if !IsSemver(tag) {
			continue
		}
		v, err := Parse(tag)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		return ""
	}

	// Sort versions descending
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Compare(versions[j]) > 0
	})

	return versions[0].String()
}

// NextPatchVersion returns the next patch version from the current version string.
func NextPatchVersion(current string) (string, error) {
	v, err := Parse(current)
	if err != nil {
		return "", err
	}
	return v.BumpPatch().String(), nil
}

// NextMinorVersion returns the next minor version from the current version string.
func NextMinorVersion(current string) (string, error) {
	v, err := Parse(current)
	if err != nil {
		return "", err
	}
	return v.BumpMinor().String(), nil
}

// NextMajorVersion returns the next major version from the current version string.
func NextMajorVersion(current string) (string, error) {
	v, err := Parse(current)
	if err != nil {
		return "", err
	}
	return v.BumpMajor().String(), nil
}
