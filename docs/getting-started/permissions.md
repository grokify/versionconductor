# Token Permissions

VersionConductor requires a GitHub token to access the API. The permissions needed depend on which commands you use.

## Personal Access Token (Classic)

For classic PATs, use the `repo` scope for full functionality.

## Fine-Grained Personal Access Token (Recommended)

For fine-grained PATs, configure permissions based on your use case:

### Read-Only (scan only)

| Permission | Access | Purpose |
|------------|--------|---------|
| Pull requests | Read | List open PRs |
| Metadata | Read | Required baseline |
| Actions | Read | Check CI status |

### Full Functionality

| Permission | Access | Purpose |
|------------|--------|---------|
| Pull requests | Read & Write | List, review, merge PRs |
| Contents | Read & Write | Merge commits, delete branches |
| Metadata | Read | Required baseline |
| Actions | Read | Check CI status |
| Checks | Read | Get check run details |
| Releases | Read & Write | Create releases and tags |

## GitHub App Token

For organizations, consider using a GitHub App with these permissions:

| Permission | Access |
|------------|--------|
| Pull requests | Read & Write |
| Contents | Read & Write |
| Metadata | Read |
| Checks | Read |

## Environment Variable

Set your token as an environment variable:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

Or use a `.env` file with your preferred method of loading it.

## GitHub Actions

In GitHub Actions, use the built-in `GITHUB_TOKEN` or a custom secret:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

!!! warning "GITHUB_TOKEN Limitations"
    The default `GITHUB_TOKEN` in GitHub Actions cannot approve PRs in some
    organization configurations. You may need to use a PAT or GitHub App token.
