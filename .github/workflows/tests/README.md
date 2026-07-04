# GitHub Actions Integration Tests

This directory contains test fixtures for running GitHub Actions workflows locally.

## Prerequisites

Install [act](https://github.com/nektos/act):

```bash
# macOS
brew install act

# Linux
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

## Running Tests

### All Tests

```bash
# Run integration tests
act -W .github/workflows/integration-test.yaml

# With workflow_dispatch event
act workflow_dispatch -W .github/workflows/integration-test.yaml \
  -e .github/workflows/tests/workflow_dispatch.json
```

### Specific Job

```bash
# Run only build and test
act -W .github/workflows/integration-test.yaml -j test-build

# Run policy evaluation tests
act -W .github/workflows/integration-test.yaml -j test-policy-evaluation
```

### With Secrets

```bash
# Provide GitHub token
act -W .github/workflows/integration-test.yaml \
  -s GITHUB_TOKEN="$(gh auth token)"
```

## Test Suites

| Suite | Description |
|-------|-------------|
| `test-build` | Build and test the Go code |
| `test-auth-github-token` | Test GITHUB_TOKEN authentication |
| `test-policy-evaluation` | Test Cedar policy evaluation |
| `test-workflow-inputs` | Test workflow input validation |
| `test-merge-strategies` | Test merge strategy options |

## Event Files

| File | Purpose |
|------|---------|
| `workflow_dispatch.json` | Simulates manual workflow trigger |

## Troubleshooting

### act fails with permission errors

```bash
# Run with elevated permissions
act --privileged
```

### Missing GitHub context

```bash
# Provide repository context
act --env GITHUB_REPOSITORY=plexusone/versionconductor
```
