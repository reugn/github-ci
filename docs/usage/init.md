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

The `init` command creates a `.github-ci.yaml` configuration file.

This command:
1. Creates `.github-ci.yaml` with default linter and upgrade settings
2. Fails if config already exists (use `--update` to add new actions)

Use `--defaults` to include all linter settings and scan workflows to discover actions.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--update`, `-u` | `false` | Update existing config with new actions from workflows |
| `--defaults`, `-d` | `false` | Include all linter settings and discover actions from workflows |
| `--path` | `.github/workflows` | Path to workflow directory or file |
| `--config` | `.github-ci.yaml` | Path to configuration file |

## Examples

### Create Minimal Config

```bash
$ github-ci init

✓ Created .github-ci.yaml
```

### Create Config with All Defaults and Actions

Include all linter settings and discover actions from workflows:

```bash
$ github-ci init --defaults

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

The init command generates a minimal config by default:

```yaml
linters:
  default: all
  enable: []
  disable: []
upgrade:
  actions: {}
  version: tag
```

With `--defaults`, all linter settings and discovered actions are included:

```yaml
linters:
  default: all
  enable:
    - permissions
    - versions
    - format
    - secrets
    - injection
    - style
  disable: []
  settings:
    format:
      indent-width: 2
      max-line-length: 120
    style:
      min-name-length: 3
      max-name-length: 50
      naming-convention: ""
      checkout-first: false
      require-step-names: false

upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ^1.0.0
    actions/setup-go:
      version: ^1.0.0
```

See [Configuration](../configuration/) for details on customizing the config.
