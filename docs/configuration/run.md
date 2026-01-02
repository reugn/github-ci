---
title: Run Settings
parent: Configuration
nav_order: 1
layout: default
---

# Run Configuration

The `run` section controls general runtime settings.

## Options

```yaml
run:
  timeout: 5m
  issues-exit-code: 1
```

### timeout

Maximum time allowed for command execution.

| Value | Description |
|-------|-------------|
| `30s` | 30 seconds |
| `2m` | 2 minutes |
| `5m` | 5 minutes (default) |
| `1h` | 1 hour |

When the timeout is reached, the command is cancelled.

```yaml
run:
  timeout: 2m  # Maximum duration for the entire command execution
```

### issues-exit-code

Exit code returned when lint issues are found.

| Value | Description |
|-------|-------------|
| `1` | Default exit code |
| `0` | Don't fail on issues (for soft warnings) |
| `2-255` | Custom exit codes |

This is useful for CI/CD pipelines that need specific exit codes for different failure types.

```yaml
run:
  issues-exit-code: 2  # Use exit code 2 for lint failures
```

## Examples

### Strict CI Configuration

Fail fast with short timeout:

```yaml
run:
  timeout: 1m
  issues-exit-code: 1
```

### Soft Warnings

Report issues but don't fail the build:

```yaml
run:
  timeout: 5m
  issues-exit-code: 0
```

### Custom Exit Code

Use different exit codes for different failure types:

```yaml
run:
  issues-exit-code: 42  # Custom code for lint failures
```
