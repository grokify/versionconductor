# Installation

## Go Install (Recommended)

```bash
go install github.com/plexusone/versionconductor/cmd/versionconductor@latest
```

This installs the latest version to `$GOPATH/bin/versionconductor`.

## Specific Version

```bash
go install github.com/plexusone/versionconductor/cmd/versionconductor@v0.3.0
```

## From Source

```bash
git clone https://github.com/plexusone/versionconductor
cd versionconductor
go build -o versionconductor ./cmd/versionconductor
```

## Verify Installation

```bash
versionconductor --version
```

## Requirements

- Go 1.22 or later
- GitHub token with appropriate permissions (see [Token Permissions](permissions.md))
