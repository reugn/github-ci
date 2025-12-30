---
title: Configuration
nav_order: 4
has_children: true
layout: default
---

# Configuration

The `github-ci` tool uses a YAML configuration file (`.github-ci.yaml` by default) to control its behavior.

## Configuration File

Create a configuration file using:

```bash
github-ci init
```

Or create `.github-ci.yaml` manually in your repository root.

## Full Example

```yaml
run:
  timeout: 5m          # timeout for GitHub API operations
  issues-exit-code: 1  # exit code when issues are found

linters:
  default: all  # 'all' or 'none'
  enable:
    - permissions
    - versions
    - format
    - secrets
    - injection
    - style
  disable: []   # linters to disable (overrides enable)
  settings:
    format:
      indent-width: 2
      max-line-length: 120

upgrade:
  version: tag  # 'tag', 'major', or 'hash'
  actions:
    actions/checkout:
      version: ^1.0.0
    actions/setup-go:
      version: ~1.0.0
```

## Sections

| Section | Description |
|---------|-------------|
| [run](run) | Runtime settings (timeout, exit codes) |
| [linters](linters) | Which linters to enable and their settings |
| [upgrade](upgrade) | Version patterns for action upgrades |

## Defaults

If no configuration file exists:

- All linters are enabled
- Actions use `^1.0.0` version pattern (allow minor/patch updates)
- Timeout is 5 minutes
- Exit code for issues is 1
- Version format is `tag`

## Using a Different Config File

```bash
github-ci lint --config custom-config.yaml
github-ci upgrade --config custom-config.yaml
```
