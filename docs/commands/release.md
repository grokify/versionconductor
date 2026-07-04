# release

Create maintenance releases for repositories with merged dependency PRs.

## Usage

```bash
versionconductor release [flags]
```

## Examples

```bash
# Dry-run (default)
versionconductor release --orgs myorg

# Create releases
versionconductor release --orgs myorg --execute

# Only PRs merged since a date
versionconductor release --orgs myorg --since 2025-01-01 --execute

# Create as drafts for review
versionconductor release --orgs myorg --draft --execute

# Custom version prefix
versionconductor release --orgs myorg --prefix "" --execute  # No prefix
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--orgs` | `-o` | Organizations to scan |
| `--repos` | `-r` | Specific repos (owner/name) |
| `--execute` | | Actually create releases (default: dry-run) |
| `--since` | | Only PRs merged since date (YYYY-MM-DD) |
| `--draft` | | Create as draft releases |
| `--prefix` | | Version prefix (default: "v") |
| `--generate-notes` | | Auto-generate release notes (default: true) |
| `--format` | `-f` | Output format |

## Version Bumping

VersionConductor automatically determines the version bump:

| Change Type | Version Bump |
|-------------|--------------|
| Dependency patches | Patch (1.2.3 → 1.2.4) |
| Dependency minors | Patch (1.2.3 → 1.2.4) |
| Multiple updates | Patch (1.2.3 → 1.2.4) |

Dependency updates always result in patch releases. Use manual releases for minor/major version bumps.

## Release Notes

Auto-generated release notes include:

- List of merged dependency PRs
- Link to each PR
- Dependency name and version change
