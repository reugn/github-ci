---
title: style
parent: Linters
nav_order: 6
layout: default
---

# style

Checks for naming conventions, structure, and style best practices in workflow files.

## Why This Matters

Consistent styling:

- **Improves readability**: Clear, descriptive names make workflows easier to understand
- **Enhances maintainability**: Well-structured workflows are easier to debug and modify
- **Promotes best practices**: Catches common anti-patterns early

## What It Detects

| Issue | Description |
|-------|-------------|
| **Missing workflow name** | Workflow without a `name:` field |
| **Missing step name** | Step without a `name:` field (opt-in via `require-step-names`) |
| **Cryptic job ID** | Job with non-descriptive ID (e.g., `j1`, `test`) and no explicit name |
| **Name too short/long** | Names outside configured length bounds |
| **Naming convention** | Names not following configured convention (title/sentence case) |
| **Name not first** | `name:` field not first in step definition |
| **Checkout not first** | `actions/checkout` not the first step (optional) |
| **Env shadowing** | Job-level env var shadows workflow-level env var |

## Example Output

```
ci.yml:1: (style) Workflow is missing a name
ci.yml:5: (style) Job 'j1' has cryptic ID and is missing a name
ci.yml:10: (style) Step is missing a name
ci.yml:15: (style) Step 'name' should come first before other fields
ci.yml:8: (style) Job env var 'NODE_ENV' shadows workflow-level env var
```

## Auto-fix

**Not supported**. Style issues require manual review as they affect semantics and readability.

## Configuration

Configure style settings in `.github-ci.yaml`:

```yaml
linters:
  settings:
    style:
      min-name-length: 3        # Minimum name length (default: 3)
      max-name-length: 50       # Maximum name length (default: 50)
      naming-convention: ""     # "title", "sentence", or "" (default: none)
      checkout-first: false     # Check checkout is first step (default: false)
      require-step-names: false # Require all steps to have names (default: false)
```

### min-name-length

Minimum allowed characters for workflow, job, and step names.

| Value | Use Case |
|-------|----------|
| `3` | Default, prevents very short names like "CI" |
| `0` | Disable minimum length check |

### max-name-length

Maximum allowed characters for names.

| Value | Use Case |
|-------|----------|
| `50` | Default, keeps names concise |
| `0` | Disable maximum length check |

### naming-convention

Enforces consistent casing for workflow, job, and step names.

| Value | Description |
|-------|-------------|
| `""` | No enforcement (default) - any casing allowed |
| `title` | Title Case - every word must start with uppercase |
| `sentence` | Sentence case - name must start with uppercase letter |

**Title Case examples:**
- ✓ "Build And Test"
- ✓ "Setup Go"
- ✓ "Run Unit Tests"
- ✗ "Build and test" (and, test lowercase)
- ✗ "setup Go" (setup lowercase)

**Sentence case examples:**
- ✓ "Build and test"
- ✓ "Build And Test" (also valid - only checks first letter)
- ✓ "Upload to Codecov"
- ✗ "build and test" (first letter lowercase)

### checkout-first

When enabled, warns if `actions/checkout` is not the first step in a job.

| Value | Description |
|-------|-------------|
| `false` | Don't check checkout position (default) |
| `true` | Warn if checkout is not first |

### require-step-names

When enabled, requires all steps to have explicit `name:` fields.

| Value | Description |
|-------|-------------|
| `false` | Don't require step names (default) |
| `true` | Warn on steps without names |

Simple steps like `- run: go build` often don't need names. Enable this for stricter naming policies.

## Cryptic Job ID Detection

A job ID is considered "cryptic" if it:

- Is less than 3 characters (e.g., `j1`, `ci`)
- Ends with a number (e.g., `job1`, `build2`)
- Is 1-4 lowercase letters only (e.g., `test`, `run`)

These IDs require an explicit `name:` field for clarity.

**Examples:**

```yaml
# Flagged - cryptic IDs without names
jobs:
  j1:           # Too short
  job1:         # Ends with number
  test:         # 4 lowercase letters

# Good - descriptive IDs or explicit names
jobs:
  build:              # 5+ chars, descriptive
  build-and-test:     # Has hyphen, descriptive
  ci:
    name: Continuous Integration  # Has explicit name
```

## Examples

### Missing Names

```yaml
# Bad - no names
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make build
```

```yaml
# Good - descriptive names
name: Build Pipeline
on: push
jobs:
  build:
    name: Build Application
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Build
        run: make build
```

### Name Order

```yaml
# Bad - name comes after uses
steps:
  - uses: actions/checkout@v4
    name: Checkout

# Good - name comes first
steps:
  - name: Checkout
    uses: actions/checkout@v4
```

### Env Shadowing

```yaml
# Warning - job shadows workflow env
name: Build
env:
  NODE_ENV: production
jobs:
  build:
    env:
      NODE_ENV: test  # Shadows workflow-level NODE_ENV
```

### Naming Convention (Title Case)

```yaml
# With naming-convention: title
linters:
  settings:
    style:
      naming-convention: title
```

```yaml
# Bad
name: build and test

# Good
name: Build And Test
```

## See Also

- [Linters Configuration](../configuration/linters) - Configure style settings
- [format](format) - For formatting checks (indentation, line length)

