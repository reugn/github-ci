---
title: Usage
nav_order: 3
has_children: true
layout: default
---

# Usage

`github-ci` provides three main commands for managing GitHub Actions workflows:

| Command | Description |
|---------|-------------|
| [init](init) | Initialize configuration file |
| [lint](lint) | Lint workflows for issues |
| [upgrade](upgrade) | Upgrade actions to latest versions |

## Common Flags

All commands support these common flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--path` | `-p` | `.github/workflows` | Path to workflow directory or file |
| `--config` | `-c` | `.github-ci.yaml` | Path to configuration file |

## Examples

```bash
# Lint all workflows in default location
github-ci lint

# Lint a specific workflow file
github-ci lint --path .github/workflows/ci.yml

# Use a custom config file
github-ci lint --config custom-config.yaml

# Combine flags
github-ci upgrade --path .github/workflows --config .github-ci.yaml --dry-run
```
