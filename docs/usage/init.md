---
title: init
parent: Usage
nav_order: 1
layout: default
---

# init Command

Initialize a configuration file with default settings.

## Synopsis

```bash
github-ci init [flags]
```

## Description

The `init` command creates a `.github-ci.yaml` configuration file and scans your workflows to discover actions.

This command:
1. Creates `.github-ci.yaml` with default linter and upgrade settings
2. Scans workflows to discover actions and adds them to the config
3. Fails if config already exists (use `--update` to add new actions)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--update` | `false` | Update existing config with new actions from workflows |
| `--path` | `.github/workflows` | Path to workflow directory or file |
| `--config` | `.github-ci.yaml` | Path to configuration file |

## Examples

### Create New Config

```bash
$ github-ci init

✓ Created .github-ci.yaml with 4 action(s)
```

### Update Existing Config

Add newly discovered actions to an existing config:

```bash
$ github-ci init --update

✓ Updated .github-ci.yaml with 2 new action(s):
  - actions/cache
  - docker/build-push-action
```

### Custom Paths

```bash
# Different workflows directory
github-ci init --path .github/custom-workflows

# Different config file name
github-ci init --config my-config.yaml
```

## Generated Config

The init command generates a config like this:

```yaml
linters:
  default: all
  enable:
    - permissions
    - versions
  settings: {}

upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ^1.0.0
    actions/setup-go:
      version: ^1.0.0
```

See [Configuration](../configuration/) for details on customizing the config.
