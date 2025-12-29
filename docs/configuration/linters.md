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

Per-linter settings. The `format` and `style` linters have configurable settings.

## Available Linters

| Linter | Description | Auto-fix |
|--------|-------------|----------|
| `permissions` | Missing permissions configuration | ✗ |
| `versions` | Actions using version tags instead of commit hashes | ✓ |
| `format` | Formatting issues | ✓ |
| `secrets` | Hardcoded secrets | ✗ |
| `injection` | Shell injection vulnerabilities | ✗ |
| `style` | Naming conventions and style best practices | ✗ |

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

## Style Linter Settings

```yaml
linters:
  settings:
    style:
      min-name-length: 3        # Minimum name length
      max-name-length: 50       # Maximum name length
      naming-convention: ""     # "title", "sentence", or ""
      checkout-first: false     # Check if checkout is first step
      require-step-names: false # Require all steps to have names
```

| Setting | Default | Description |
|---------|---------|-------------|
| `min-name-length` | `3` | Minimum characters for names |
| `max-name-length` | `50` | Maximum characters for names |
| `naming-convention` | `""` | `"title"` (Every Word Uppercase), `"sentence"` (First word only), or `""` (none) |
| `checkout-first` | `false` | Warn if checkout is not first step |
| `require-step-names` | `false` | Require all steps to have names |

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
