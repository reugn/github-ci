---
title: Linters Settings
parent: Configuration
nav_order: 2
layout: default
---

# Linters Configuration

The `linters` section controls which linters are enabled and their settings.

## Options

```yaml
linters:
  default: all
  enable:
    - permissions
    - versions
  disable:
    - format
  settings:
    format:
      indent-width: 2
      max-line-length: 120
```

### default

Controls the baseline for which linters are enabled.

| Value | Description |
|-------|-------------|
| `all` | All linters enabled by default (then use `disable` to turn off specific ones) |
| `none` | All linters disabled by default (then use `enable` to turn on specific ones) |

### enable

List of linters to enable. When `default: all`, this is redundant but can be used for documentation purposes.

```yaml
linters:
  default: none
  enable:
    - permissions
    - versions
```

### disable

List of linters to disable. Takes precedence over `enable`.

```yaml
linters:
  default: all
  disable:
    - format
    - secrets
```

### settings

Per-linter settings. Currently only `format` linter has configurable settings.

## Available Linters

| Linter | Description | Auto-fix |
|--------|-------------|----------|
| `permissions` | Missing permissions configuration | ✗ |
| `versions` | Actions using version tags instead of commit hashes | ✓ |
| `format` | Formatting issues | ✓ |
| `secrets` | Hardcoded secrets | ✗ |
| `injection` | Shell injection vulnerabilities | ✗ |

## Format Linter Settings

```yaml
linters:
  settings:
    format:
      indent-width: 2      # Expected indentation width (spaces)
      max-line-length: 120 # Maximum line length
```

| Setting | Default | Description |
|---------|---------|-------------|
| `indent-width` | `2` | Expected number of spaces per indentation level |
| `max-line-length` | `120` | Maximum allowed line length |

## Examples

### Enable Only Security Linters

```yaml
linters:
  default: none
  enable:
    - permissions
    - secrets
    - injection
```

### All Linters Except Format

```yaml
linters:
  default: all
  disable:
    - format
```

### Custom Format Settings

```yaml
linters:
  default: all
  settings:
    format:
      indent-width: 4
      max-line-length: 80
```

### Minimal Config (Defaults)

When `default: all` is used (the default), all linters run without explicit configuration:

```yaml
linters:
  default: all
```

## See Also

- [Linters Reference](../linters/) - Detailed documentation for each linter
