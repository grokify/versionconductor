# graph

Analyze dependency relationships across repositories.

All `graph` subcommands (except `build`) load a graph via cache if available, or build one live from GitHub. All require `--orgs`.

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `build` | Build dependency graph |
| `dependents <module>` | Show modules that depend on a module |
| `dependencies <module>` | Show dependencies of a module |
| `order` | Show upgrade order for managed modules |
| `stale <module> --min-version <version>` | Find modules using outdated versions of a dependency |
| `stats` | Show graph statistics |
| `visualize` | Generate graph visualization |

## build

Build the dependency graph from repositories:

```bash
versionconductor graph build --orgs myorg

# Multiple orgs, filter by language
versionconductor graph build --orgs grokify,agentplexus --languages go

# Write graph JSON to a file
versionconductor graph build --orgs myorg --format json --output graph.json
```

## dependents

Show which modules depend on a given module (the module is a positional argument, not a flag):

```bash
versionconductor graph dependents github.com/grokify/mogo --orgs myorg
```

## dependencies

Show dependencies of a module:

```bash
versionconductor graph dependencies github.com/grokify/gogithub --orgs myorg
```

## order

Show the topological order to upgrade managed modules in:

```bash
versionconductor graph order --orgs myorg

# Filter by organization
versionconductor graph order --orgs myorg --org github.com/grokify
```

## stale

Find managed modules using an outdated version of a dependency. Both the module (positional argument) and `--min-version` are required:

```bash
versionconductor graph stale github.com/grokify/gogithub --min-version v0.7.0 --orgs myorg
```

## stats

Show statistics about the dependency graph:

```bash
versionconductor graph stats --orgs myorg
```

Output:

```
Total modules: 25
Managed modules: 15
External dependencies: 150
Edges: 200
Max depth: 5
```

## visualize

Generate a visual representation of the graph. Use `--viz-format` (not the global `--format`, which controls table/json/markdown/csv output elsewhere):

```bash
# DOT format (Graphviz), the default
versionconductor graph visualize --orgs myorg > deps.dot
dot -Tpng deps.dot -o deps.png

# Mermaid format
versionconductor graph visualize --orgs myorg --viz-format mermaid > deps.md

# Include external (unmanaged) dependencies, change layout direction
versionconductor graph visualize --orgs myorg --show-external --direction LR
```

## Flags

### Global (apply to all subcommands)

| Flag | Description |
|------|-------------|
| `--orgs` | Organizations to include (required) |
| `--repos` | Specific repos |
| `--token` | GitHub token (or set `GITHUB_TOKEN`) |
| `--format` | Output format: table, json, markdown, csv |

### Caching (apply to all `graph` subcommands except `build`)

| Flag | Default | Description |
|------|---------|-------------|
| `--cache` | `true` | Enable caching of API responses |
| `--cache-dir` | system temp | Cache directory |
| `--cache-ttl` | `1h` | Cache TTL duration |
| `--no-cache` | `false` | Disable caching |

### `build`

| Flag | Default | Description |
|------|---------|-------------|
| `--languages` | `go` | Languages to scan: go, typescript, swift |
| `--output` | stdout | Output file for graph JSON |

### `order`

| Flag | Description |
|------|-------------|
| `--org` | Filter by organization |

### `stale`

| Flag | Description |
|------|-------------|
| `--min-version` | Minimum required version (required) |

### `visualize`

| Flag | Default | Description |
|------|---------|-------------|
| `--viz-format` | `dot` | Output format: dot, mermaid |
| `--show-external` | `false` | Include external dependencies |
| `--show-versions` | `true` | Show version labels on edges |
| `--cluster` | `true` | Cluster nodes by organization |
| `--direction` | `TB` | Layout direction: TB, LR, BT, RL |
