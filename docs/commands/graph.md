# graph

Analyze dependency relationships across repositories.

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `build` | Build dependency graph |
| `dependents` | Show modules that depend on a module |
| `dependencies` | Show dependencies of a module |
| `order` | Show upgrade order for a module |
| `stale` | Find stale dependencies |
| `stats` | Show graph statistics |
| `visualize` | Generate graph visualization |

## build

Build the dependency graph from repositories:

```bash
versionconductor graph build --orgs myorg
```

## dependents

Show which modules depend on a given module:

```bash
versionconductor graph dependents --orgs myorg --module github.com/myorg/shared-lib
```

## dependencies

Show dependencies of a module:

```bash
versionconductor graph dependencies --orgs myorg --module github.com/myorg/api-server
```

## order

Show the order to upgrade modules when updating a dependency:

```bash
versionconductor graph order --orgs myorg --module github.com/myorg/shared-lib
```

## stale

Find modules using outdated versions of dependencies:

```bash
versionconductor graph stale --orgs myorg
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

Generate a visual representation of the graph:

```bash
# DOT format (Graphviz)
versionconductor graph visualize --orgs myorg --format dot > deps.dot
dot -Tpng deps.dot -o deps.png

# Mermaid format
versionconductor graph visualize --orgs myorg --format mermaid > deps.md
```

## Common Flags

| Flag | Description |
|------|-------------|
| `--orgs` | Organizations to include |
| `--repos` | Specific repos |
| `--language` | Filter by language (go, typescript, etc.) |
| `--cache` | Enable caching |
| `--cache-dir` | Cache directory |
