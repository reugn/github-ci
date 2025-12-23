---
title: permissions
parent: Linters
nav_order: 1
layout: default
---

# permissions

Checks that workflows have explicit permissions configuration.

## Why This Matters

By default, GitHub Actions workflows have broad permissions. Explicitly defining permissions:

- **Reduces attack surface**: Limits what a compromised workflow can access
- **Follows least-privilege principle**: Only grant permissions that are needed
- **Prevents accidental damage**: Limits blast radius of bugs or mistakes

## What It Detects

Workflows missing the `permissions` key at the workflow or job level.

### ❌ Bad

```yaml
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
```

### ✅ Good

```yaml
name: CI
on: push
permissions:
  contents: read
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
```

## Example Output

```
ci.yml: (permissions) Workflow is missing permissions configuration
```

## Auto-fix

**Not supported** - Permissions depend on what the workflow actually needs to do. You must add them manually.

## Common Permission Configurations

### Read-only (Most Restrictive)

```yaml
permissions:
  contents: read
```

### Minimal for CI

```yaml
permissions:
  contents: read
  checks: write
```

### Publishing Packages

```yaml
permissions:
  contents: read
  packages: write
```

### Creating Releases

```yaml
permissions:
  contents: write
```

### Read All (Convenience)

```yaml
permissions: read-all
```

### Job-level Permissions

You can also set permissions per job:

```yaml
permissions:
  contents: read

jobs:
  build:
    runs-on: ubuntu-latest
    steps: ...

  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write  # Override for this job only
    steps: ...
```

## See Also

- [GitHub Docs: Permissions for GITHUB_TOKEN](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token)
