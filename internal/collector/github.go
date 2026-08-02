package collector

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/checks"
	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/gogithub/pr"
	"github.com/grokify/gogithub/release"
	"github.com/grokify/gogithub/tag"

	"github.com/plexusone/versionconductor/pkg/model"
)

// GitHubCollector implements Collector for GitHub repositories.
type GitHubCollector struct {
	client clientv1.Client
}

// NewGitHubCollector creates a new GitHub collector.
func NewGitHubCollector(token string) (*GitHubCollector, error) {
	ctx := context.Background()
	client, err := clientv1.NewClient(ctx, token)
	if err != nil {
		return nil, err
	}
	return &GitHubCollector{
		client: client,
	}, nil
}

// ListRepos returns repositories matching the filter criteria.
func (c *GitHubCollector) ListRepos(ctx context.Context, orgs []string, filter model.RepoFilter) ([]model.Repo, error) {
	var repos []model.Repo

	for _, org := range orgs {
		ghRepos, err := c.client.ListOrgRepos(ctx, org)
		if err != nil {
			return nil, err
		}

		for _, r := range ghRepos {
			repo := convertRepo(r)

			// Apply filters
			if repo.Archived && !filter.IncludeArchived {
				continue
			}
			if repo.Private && !filter.IncludePrivate {
				continue
			}
			if r.Fork && !filter.IncludeForks {
				continue
			}
			if isExcluded(repo.FullName, filter.ExcludeRepos) {
				continue
			}

			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// ListDependencyPRs returns open dependency PRs for a repository.
func (c *GitHubCollector) ListDependencyPRs(ctx context.Context, repo model.RepoRef) ([]model.PullRequest, error) {
	var prs []model.PullRequest

	opts := &clientv1.ListPullRequestsOptions{State: "open"}

	ghPRs, err := pr.ListPRs(ctx, c.client, repo.Owner, repo.Name, opts)
	if err != nil {
		return nil, err
	}

	for _, ghPR := range ghPRs {
		mpr := convertPR(ghPR, repo)

		// Check if this is a dependency PR
		mpr.DependBot = model.DetectDependBot(mpr.Author)
		if mpr.DependBot != model.DependBotUnknown {
			mpr.IsDependency = true
			mpr.Dependency = parseDependencyFromTitle(mpr.Title)
			prs = append(prs, mpr)
		}
	}

	return prs, nil
}

// GetPRDetails returns detailed information about a specific PR.
func (c *GitHubCollector) GetPRDetails(ctx context.Context, repo model.RepoRef, prNumber int) (*model.PullRequest, error) {
	ghPR, err := pr.GetPR(ctx, c.client, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	mpr := convertPR(ghPR, repo)
	mpr.DependBot = model.DetectDependBot(mpr.Author)
	if mpr.DependBot != model.DependBotUnknown {
		mpr.IsDependency = true
		mpr.Dependency = parseDependencyFromTitle(mpr.Title)
	}

	// Get mergeable status
	if ghPR.Mergeable != nil {
		mpr.Mergeable = *ghPR.Mergeable
	}

	return &mpr, nil
}

// GetPRChecks returns the CI check runs for a PR.
func (c *GitHubCollector) GetPRChecks(ctx context.Context, repo model.RepoRef, prNumber int) ([]model.CheckRun, error) {
	ghChecks, err := checks.ListCheckRunsForPR(ctx, c.client, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	var result []model.CheckRun
	for _, cr := range ghChecks {
		result = append(result, model.CheckRun{
			Name:       cr.Name,
			Status:     cr.Status,
			Conclusion: cr.Conclusion,
		})
	}

	return result, nil
}

// GetLatestRelease returns the most recent release for a repository.
func (c *GitHubCollector) GetLatestRelease(ctx context.Context, repo model.RepoRef) (*model.Release, error) {
	ghRelease, err := release.GetLatestRelease(ctx, c.client, repo.Owner, repo.Name)
	if err != nil {
		// Check for 404 (no releases)
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}
		return nil, err
	}

	return &model.Release{
		ID:          ghRelease.ID,
		TagName:     ghRelease.TagName,
		Name:        ghRelease.Name,
		Body:        ghRelease.Body,
		Draft:       ghRelease.Draft,
		Prerelease:  ghRelease.Prerelease,
		CreatedAt:   ghRelease.CreatedAt,
		PublishedAt: derefTime(ghRelease.PublishedAt),
		HTMLURL:     ghRelease.HTMLURL,
		Repo:        repo,
	}, nil
}

// ListTags returns all tags for a repository.
func (c *GitHubCollector) ListTags(ctx context.Context, repo model.RepoRef) ([]model.Tag, error) {
	ghTags, err := tag.ListTags(ctx, c.client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	var tags []model.Tag
	for _, t := range ghTags {
		tags = append(tags, model.Tag{
			Name: t.Name,
			SHA:  t.SHA,
			Repo: repo,
		})
	}

	return tags, nil
}

// GetMergedPRsSinceTag returns PRs merged since the given tag.
//
// Note: unlike go-github's raw pagination, clientv1.ListPullRequests always
// fetches every matching PR before returning (it has no early-exit hook), so
// this can make more API calls than strictly necessary for repos with a very
// long closed-PR history. Results are still filtered correctly by since.
func (c *GitHubCollector) GetMergedPRsSinceTag(ctx context.Context, repo model.RepoRef, tagName string) ([]model.PullRequest, error) {
	// Get the tag's commit date
	tagSHA, err := tag.GetTagSHA(ctx, c.client, repo.Owner, repo.Name, tagName)
	if err != nil {
		return nil, err
	}

	commit, err := c.client.GetCommit(ctx, repo.Owner, repo.Name, tagSHA)
	if err != nil {
		return nil, err
	}

	since := commit.Committer.Date

	opts := &clientv1.ListPullRequestsOptions{
		State:     "closed",
		Sort:      "updated",
		Direction: "desc",
	}

	ghPRs, err := pr.ListPRs(ctx, c.client, repo.Owner, repo.Name, opts)
	if err != nil {
		return nil, err
	}

	var prs []model.PullRequest
	for _, ghPR := range ghPRs {
		if ghPR.MergedAt == nil {
			continue
		}
		if ghPR.MergedAt.Before(since) {
			continue
		}

		mpr := convertPR(ghPR, repo)
		mpr.DependBot = model.DetectDependBot(mpr.Author)
		if mpr.DependBot != model.DependBotUnknown {
			mpr.IsDependency = true
			mpr.Dependency = parseDependencyFromTitle(mpr.Title)
		}
		prs = append(prs, mpr)
	}

	return prs, nil
}

// convertRepo converts a GitHub repository to our model.
func convertRepo(r *gogithub.Repository) model.Repo {
	owner := ""
	if r.Owner != nil {
		owner = r.Owner.Login
	}

	return model.Repo{
		Owner:         owner,
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
		Archived:      r.Archived,
		Language:      r.Language,
		Topics:        r.Topics,
		UpdatedAt:     r.UpdatedAt,
		HTMLURL:       r.HTMLURL,
	}
}

// convertPR converts a GitHub pull request to our model.
func convertPR(ghPR *gogithub.PullRequest, repo model.RepoRef) model.PullRequest {
	var labels []string
	for _, l := range ghPR.Labels {
		labels = append(labels, l.Name)
	}

	author := ""
	if ghPR.User != nil {
		author = ghPR.User.Login
	}

	mpr := model.PullRequest{
		Number:    ghPR.Number,
		Title:     ghPR.Title,
		Body:      ghPR.Body,
		State:     ghPR.State,
		Author:    author,
		HTMLURL:   ghPR.HTMLURL,
		Draft:     ghPR.Draft,
		Labels:    labels,
		CreatedAt: ghPR.CreatedAt,
		UpdatedAt: ghPR.UpdatedAt,
		Repo:      repo,
	}

	if ghPR.MergedAt != nil {
		t := *ghPR.MergedAt
		mpr.MergedAt = &t
	}

	return mpr
}

// derefTime returns the zero value if t is nil, else *t.
func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// parseDependencyFromTitle extracts dependency information from a PR title.
func parseDependencyFromTitle(title string) model.Dependency {
	dep := model.Dependency{}

	// Try to extract version numbers
	versionRe := regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)
	versions := versionRe.FindAllString(title, 2)

	if len(versions) >= 2 {
		dep.FromVersion = versions[0]
		dep.ToVersion = versions[1]
		dep.UpdateType = determineUpdateType(dep.FromVersion, dep.ToVersion)
	} else if len(versions) == 1 {
		dep.ToVersion = versions[0]
	}

	// Try to extract dependency name
	patterns := []string{
		`(?:update|bump|upgrade)\s+(?:dependency\s+)?(\S+)`,
		`deps(?:\([^)]+\))?:\s*(?:update|bump|upgrade)\s+(\S+)`,
		`(\S+)\s+from\s+v?\d`,
	}

	lower := strings.ToLower(title)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(lower); len(matches) > 1 {
			dep.Name = matches[1]
			break
		}
	}

	// Detect ecosystem from dependency name
	dep.Ecosystem = detectEcosystem(dep.Name)

	return dep
}

// determineUpdateType determines the semantic version update type.
func determineUpdateType(from, to string) model.UpdateType {
	fromParts := parseVersion(from)
	toParts := parseVersion(to)

	if len(fromParts) < 3 || len(toParts) < 3 {
		return model.UpdateTypeUnknown
	}

	if toParts[0] > fromParts[0] {
		return model.UpdateTypeMajor
	}
	if toParts[1] > fromParts[1] {
		return model.UpdateTypeMinor
	}
	if toParts[2] > fromParts[2] {
		return model.UpdateTypePatch
	}

	return model.UpdateTypeUnknown
}

// parseVersion parses a version string into numeric parts.
func parseVersion(v string) []int {
	// Remove leading 'v'
	v = strings.TrimPrefix(v, "v")

	parts := strings.Split(v, ".")
	result := make([]int, len(parts))

	for i, p := range parts {
		// Parse only numeric prefix
		var num int
		for _, ch := range p {
			if ch >= '0' && ch <= '9' {
				num = num*10 + int(ch-'0')
			} else {
				break
			}
		}
		result[i] = num
	}

	return result
}

// detectEcosystem attempts to detect the package ecosystem from the dependency name.
func detectEcosystem(name string) string {
	switch {
	case strings.HasPrefix(name, "github.com/"):
		return "go"
	case strings.HasPrefix(name, "golang.org/"):
		return "go"
	case strings.HasPrefix(name, "@"):
		return "npm"
	case strings.Contains(name, "/") && !strings.Contains(name, "."):
		return "npm"
	default:
		return ""
	}
}

// isExcluded checks if a repo is in the exclude list.
func isExcluded(fullName string, excludeList []string) bool {
	for _, ex := range excludeList {
		if fullName == ex {
			return true
		}
	}
	return false
}

// TestsPassed checks if all check runs passed.
func TestsPassed(checkRuns []model.CheckRun) bool {
	if len(checkRuns) == 0 {
		return false
	}

	for _, c := range checkRuns {
		if !c.IsSuccess() {
			return false
		}
	}
	return true
}

// PRComment represents a comment on a pull request.
type PRComment struct {
	ID        int64
	Body      string
	Author    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreatePRComment creates a new comment on a pull request.
func (c *GitHubCollector) CreatePRComment(ctx context.Context, repo model.RepoRef, prNumber int, body string) (*PRComment, error) {
	created, err := c.client.CreateIssueComment(ctx, repo.Owner, repo.Name, prNumber, body)
	if err != nil {
		return nil, err
	}

	return issueCommentToPRComment(created), nil
}

// UpdatePRComment updates an existing comment on a pull request.
func (c *GitHubCollector) UpdatePRComment(ctx context.Context, repo model.RepoRef, commentID int64, body string) (*PRComment, error) {
	updated, err := c.client.EditIssueComment(ctx, repo.Owner, repo.Name, commentID, body)
	if err != nil {
		return nil, err
	}

	return issueCommentToPRComment(updated), nil
}

// FindBotCommentByMarker finds an existing comment containing a specific marker.
// This is used to find comments created by the bot for updating.
func (c *GitHubCollector) FindBotCommentByMarker(ctx context.Context, repo model.RepoRef, prNumber int, marker string) (*PRComment, error) {
	comments, err := c.client.ListIssueComments(ctx, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	for _, comment := range comments {
		if strings.Contains(comment.Body, marker) {
			return issueCommentToPRComment(comment), nil
		}
	}

	return nil, nil // Not found
}

func issueCommentToPRComment(c *gogithub.IssueComment) *PRComment {
	author := ""
	if c.User != nil {
		author = c.User.Login
	}
	return &PRComment{
		ID:        c.ID,
		Body:      c.Body,
		Author:    author,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// WaitForChecks polls until all checks complete or timeout.
func (c *GitHubCollector) WaitForChecks(ctx context.Context, repo model.RepoRef, prNumber int, timeout time.Duration) ([]model.CheckRun, error) {
	// Get PR to get head SHA
	ghPR, err := pr.GetPR(ctx, c.client, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	sha := ""
	if ghPR.Head != nil {
		sha = ghPR.Head.SHA
	}
	pollInterval := 30 * time.Second

	ghChecks, _, err := checks.WaitForChecks(ctx, c.client, repo.Owner, repo.Name, sha, timeout, pollInterval)
	if err != nil {
		return nil, err
	}

	var result []model.CheckRun
	for _, cr := range ghChecks {
		result = append(result, model.CheckRun{
			Name:       cr.Name,
			Status:     cr.Status,
			Conclusion: cr.Conclusion,
		})
	}

	return result, nil
}
