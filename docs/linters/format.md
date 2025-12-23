---
title: format
parent: Linters
nav_order: 3
layout: default
---

# format

Checks for YAML formatting issues in workflow files.

## Why This Matters

Consistent formatting:

- **Improves readability**: Easier to review and maintain
- **Reduces merge conflicts**: Consistent style means fewer formatting-only changes
- **Catches errors early**: Some formatting issues can cause YAML parsing problems

## What It Detects

| Issue | Description |
|-------|-------------|
| **Trailing whitespace** | Spaces at the end of lines |
| **Multiple blank lines** | More than one consecutive empty line |
| **Line length** | Lines exceeding max length (default: 120) |
| **Indentation** | Incorrect indentation width or tabs |

## Example Output

```
ci.yml:1: (format) Line has trailing whitespace
ci.yml:10: (format) Multiple consecutive blank lines
ci.yml:25: (format) Line exceeds maximum length of 120 characters (got 145)
ci.yml:30: (format) Line uses tabs for indentation; use spaces instead
```

## Auto-fix

**Partially supported** with `--fix`:

| Issue | Auto-fix |
|-------|----------|
| Trailing whitespace | ✓ |
| Multiple blank lines | ✓ |
| Line length | ✗ |
| Indentation | ✗ |

```bash
github-ci lint --fix
```

{: .note }
> Line length and indentation issues require manual fixing as they may affect the meaning of the YAML.

## Configuration

Configure format settings in `.github-ci.yaml`:

```yaml
linters:
  settings:
    format:
      indent-width: 2      # Expected spaces per indent level
      max-line-length: 120 # Maximum line length
```

### indent-width

The expected number of spaces for each indentation level.

| Value | Use Case |
|-------|----------|
| `2` | Default, common in YAML |
| `4` | More readable for deeply nested content |

### max-line-length

Maximum allowed characters per line.

| Value | Use Case |
|-------|----------|
| `80` | Traditional terminal width |
| `100` | Common modern standard |
| `120` | Default, balances readability |
| `0` | Disable line length check |

## Examples

### Trailing Whitespace

```yaml
# Bad - has invisible spaces after the colon
name: CI   
on: push
```

```yaml
# Good
name: CI
on: push
```

### Multiple Blank Lines

```yaml
# Bad
name: CI


on: push
```

```yaml
# Good
name: CI

on: push
```

### Line Length

```yaml
# Bad - too long
- run: echo "This is an extremely long command that exceeds the maximum line length limit and should be split"

# Good - use YAML multiline
- run: |
    echo "This is a long command that has been"
    echo "split across multiple lines for readability"
```

### Indentation

```yaml
# Bad - using 4 spaces when 2 is expected
jobs:
    build:
        runs-on: ubuntu-latest

# Good - consistent 2-space indent
jobs:
  build:
    runs-on: ubuntu-latest
```

## See Also

- [Linters Configuration](../configuration/linters) - Configure format settings
