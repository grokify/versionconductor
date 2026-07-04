# Creating the GitHub App

This guide walks through creating a GitHub App for the `versionconductor[bot]` identity.

## Why Use a GitHub App?

| Feature | GITHUB_TOKEN | GitHub App |
|---------|--------------|------------|
| Identity | `github-actions[bot]` | `versionconductor[bot]` |
| Audit trail | Generic | Clear bot identity |
| Token lifetime | Job duration | 1 hour (renewable) |
| Permissions | Repository-scoped | Fine-grained per installation |
| Cross-repo | No | Yes (within installation) |

## Step 1: Create the App

Navigate to your organization's app settings:

```
https://github.com/organizations/YOUR_ORG/settings/apps/new
```

Or: **GitHub → Your Org → Settings → Developer settings → GitHub Apps → New GitHub App**

### Basic Information

| Field | Value |
|-------|-------|
| **GitHub App name** | `versionconductor` |
| **Description** | Automated dependency PR management |
| **Homepage URL** | `https://github.com/YOUR_ORG/versionconductor` |

### Webhook

Uncheck **Active** - webhooks are not needed for this use case.

### Permissions

Configure these **Repository permissions**:

| Permission | Access | Purpose |
|------------|--------|---------|
| **Contents** | Read | Read go.mod/go.sum files |
| **Metadata** | Read | Access repository metadata |
| **Pull requests** | Read and write | Approve and merge PRs |
| **Checks** | Read | Verify CI status |
| **Commit statuses** | Read | Check commit statuses |

Leave all other permissions as "No access".

### Installation Scope

Select: **Only on this account**

This restricts the app to your organization only.

## Step 2: Note the App ID

After clicking "Create GitHub App", you'll see the app settings page.

Note the **App ID** displayed near the top - you'll need this for the workflow.

```
App ID: 123456
```

## Step 3: Generate Private Key

1. Scroll down to the **Private keys** section
2. Click **Generate a private key**
3. A `.pem` file will download automatically
4. Store this file securely - it cannot be recovered if lost

!!! warning "Private Key Security"
    The private key grants full access as the app. Store it securely and never commit it to version control.

## Step 4: Install the App

1. In the app settings, click **Install App** in the left sidebar
2. Select your organization
3. Choose installation scope:
   - **All repositories** - recommended for org-wide dependency management
   - **Only select repositories** - for limited rollout
4. Click **Install**

## Step 5: Add Secrets

Add the App ID and private key as secrets for your workflows.

### Option A: Web UI

1. Go to your organization's secrets page:
   ```
   https://github.com/organizations/YOUR_ORG/settings/secrets/actions
   ```
2. Click **New organization secret**
3. Name: `VERSIONCONDUCTOR_APP_ID`, Value: your App ID (e.g., `123456`)
4. Visibility: **All repositories** (or select specific repos)
5. Click **Add secret**
6. Repeat for `VERSIONCONDUCTOR_PRIVATE_KEY` (paste full `.pem` file contents)

### Option B: GitHub CLI

**Organization-Level Secrets (Recommended)**

Makes secrets available to all repositories in the organization:

```bash
# Set App ID
gh secret set VERSIONCONDUCTOR_APP_ID \
  --body "123456" \
  --org YOUR_ORG \
  --visibility all

# Set private key from file
gh secret set VERSIONCONDUCTOR_PRIVATE_KEY \
  --org YOUR_ORG \
  --visibility all \
  < /path/to/versionconductor.pem
```

**Repository-Level Secrets**

For individual repositories:

```bash
# Set App ID
gh secret set VERSIONCONDUCTOR_APP_ID \
  --body "123456" \
  --repo YOUR_ORG/your-repo

# Set private key from file
gh secret set VERSIONCONDUCTOR_PRIVATE_KEY \
  --repo YOUR_ORG/your-repo \
  < /path/to/versionconductor.pem
```

## Step 6: Configure Workflow

Update your workflow to use the GitHub App credentials:

```yaml
jobs:
  auto-merge:
    uses: plexusone/.github/.github/workflows/go-dependency-automerge.yaml@main
    with:
      profile: 'quarantine'
      min-age-days: 5
    secrets:
      app-id: ${{ secrets.VERSIONCONDUCTOR_APP_ID }}
      app-private-key: ${{ secrets.VERSIONCONDUCTOR_PRIVATE_KEY }}
```

## Step 7: Verify Setup

Trigger a test run to verify the app is working:

```bash
gh workflow run go-dependency-automerge.yaml \
  --repo YOUR_ORG/your-repo \
  -f dry-run=true
```

Check the workflow logs. You should see:

```
Using GitHub App token (versionconductor[bot])
```

PR approvals will now show as coming from `versionconductor[bot]`.

## Troubleshooting

### "Resource not accessible by integration"

The app doesn't have the required permissions. Check:

1. App has "Pull requests: Read and write" permission
2. App is installed on the repository
3. Secrets are correctly configured

### "Could not create installation token"

The private key or app ID is incorrect. Verify:

1. App ID matches the one shown in app settings
2. Private key is the complete `.pem` file contents
3. No extra whitespace in secrets

### App not appearing in PR approvals

The workflow may be falling back to `GITHUB_TOKEN`. Check:

1. Both `app-id` and `app-private-key` secrets are set
2. Secrets are available to the workflow (check visibility settings)
3. Workflow logs show "Using GitHub App token"

## Rotating the Private Key

To rotate the private key:

1. Go to app settings → Private keys
2. Click **Generate a private key** (creates new key)
3. Update the `VERSIONCONDUCTOR_PRIVATE_KEY` secret
4. Verify workflows still work
5. Delete the old key from app settings

## Revoking Access

To revoke the app's access to a repository:

1. Go to app settings → Install App
2. Click the gear icon next to your organization
3. Under "Repository access", deselect the repository
4. Click **Save**

To completely uninstall:

1. Go to organization settings → Installed GitHub Apps
2. Click **Configure** next to versionconductor
3. Click **Uninstall** at the bottom

## Next Steps

- [Setup Guide](setup.md) - Configure the workflow in your repositories
- [Examples](examples.md) - Common configuration patterns
