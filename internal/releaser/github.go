package releaser

import (
	"context"
	"fmt"
	"time"

	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/gogithub/release"
	"github.com/grokify/gogithub/tag"

	"github.com/plexusone/versionconductor/pkg/model"
)

// GitHubReleaser implements Releaser for GitHub.
type GitHubReleaser struct {
	client clientv1.Client
}

// NewGitHubReleaser creates a new GitHub releaser.
func NewGitHubReleaser(token string) (*GitHubReleaser, error) {
	ctx := context.Background()
	client, err := clientv1.NewClient(ctx, token)
	if err != nil {
		return nil, err
	}
	return &GitHubReleaser{
		client: client,
	}, nil
}

// CreateRelease creates a new release for a repository.
func (r *GitHubReleaser) CreateRelease(ctx context.Context, req *model.ReleaseRequest) (*model.Release, error) {
	input := &clientv1.CreateReleaseInput{
		TagName:              req.TagName,
		TargetCommitish:      req.TargetCommitish,
		Name:                 req.Name,
		Body:                 req.Body,
		Draft:                req.Draft,
		Prerelease:           req.Prerelease,
		GenerateReleaseNotes: req.GenerateNotes,
	}

	created, err := release.CreateRelease(ctx, r.client, req.Repo.Owner, req.Repo.Name, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create release: %w", err)
	}

	publishedAt := time.Time{}
	if created.PublishedAt != nil {
		publishedAt = *created.PublishedAt
	}

	return &model.Release{
		ID:          created.ID,
		TagName:     created.TagName,
		Name:        created.Name,
		Body:        created.Body,
		Draft:       created.Draft,
		Prerelease:  created.Prerelease,
		CreatedAt:   created.CreatedAt,
		PublishedAt: publishedAt,
		HTMLURL:     created.HTMLURL,
		Repo:        req.Repo,
	}, nil
}

// CreateTag creates a new tag for a repository.
func (r *GitHubReleaser) CreateTag(ctx context.Context, repo model.RepoRef, tagName, sha, message string) error {
	return tag.CreateTag(ctx, r.client, repo.Owner, repo.Name, tagName, sha, message)
}

// GetLatestTag returns the most recent semver tag.
func (r *GitHubReleaser) GetLatestTag(ctx context.Context, repo model.RepoRef) (string, error) {
	tagNames, err := tag.GetTagNames(ctx, r.client, repo.Owner, repo.Name)
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %w", err)
	}

	latest := FindLatestVersion(tagNames)
	if latest == "" {
		return "", fmt.Errorf("no semver tags found")
	}

	return latest, nil
}

// GetTagSHA returns the SHA for a given tag.
func (r *GitHubReleaser) GetTagSHA(ctx context.Context, repo model.RepoRef, tagName string) (string, error) {
	return tag.GetTagSHA(ctx, r.client, repo.Owner, repo.Name, tagName)
}

// GetDefaultBranchSHA returns the SHA of the default branch HEAD.
func (r *GitHubReleaser) GetDefaultBranchSHA(ctx context.Context, repo model.RepoRef, branch string) (string, error) {
	sha, err := r.client.GetBranchSHA(ctx, repo.Owner, repo.Name, branch)
	if err != nil {
		return "", fmt.Errorf("failed to get branch ref: %w", err)
	}

	return sha, nil
}
