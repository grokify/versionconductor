package model

import "time"

// Repo represents a GitHub repository.
type Repo struct {
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	FullName      string    `json:"fullName"`
	Description   string    `json:"description,omitempty"`
	DefaultBranch string    `json:"defaultBranch"`
	Private       bool      `json:"private"`
	Archived      bool      `json:"archived"`
	Language      string    `json:"language,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
	HTMLURL       string    `json:"htmlUrl"`
}

// RepoFilter defines criteria for filtering repositories.
type RepoFilter struct {
	IncludeArchived  bool     `json:"includeArchived"`
	IncludePrivate   bool     `json:"includePrivate"`
	IncludeForks     bool     `json:"includeForks"`
	Languages        []string `json:"languages,omitempty"`
	Topics           []string `json:"topics,omitempty"`
	ExcludeRepos     []string `json:"excludeRepos,omitempty"`
	MinStars         int      `json:"minStars,omitempty"`
	HasOpenDependPRs bool     `json:"hasOpenDependPRs"`
}

// RepoRef is a lightweight reference to a repository.
type RepoRef struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// FullName returns the full repository name in owner/repo format.
func (r RepoRef) FullName() string {
	return r.Owner + "/" + r.Name
}

// ParseRepoRef parses a full name like "owner/repo" into a RepoRef.
func ParseRepoRef(fullName string) RepoRef {
	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			return RepoRef{
				Owner: fullName[:i],
				Name:  fullName[i+1:],
			}
		}
	}
	return RepoRef{Name: fullName}
}
