---
title: Linters
nav_order: 5
has_children: true
layout: default
---

# Linters

`github-ci` includes multiple linters to check GitHub Actions workflows for best practices and potential issues.

## Available Linters

| Linter | Description | Auto-fix |
|--------|-------------|----------|
| [permissions](permissions) | Missing permissions configuration | ✗ |
| [versions](versions) | Actions using version tags instead of commit hashes | ✓ |
| [format](format) | Formatting issues (indentation, line length, whitespace) | ✓ |
| [secrets](secrets) | Hardcoded secrets and sensitive information | ✗ |
| [injection](injection) | Shell injection vulnerabilities | ✗ |
| [style](style) | Naming conventions and style best practices | ✗ |

## Enabling/Disabling Linters

Configure linters in `.github-ci.yaml`:

```yaml
linters:
  default: all  # or 'none'
  enable:
    - permissions
    - versions
  disable:
    - format
```

See [Linters Configuration](../configuration/linters) for details.

## Auto-fix Support

Some linters support automatic fixing with `--fix`:

```bash
github-ci lint --fix
```

| Linter | What's Fixed |
|--------|--------------|
| versions | Replaces version tags with commit hashes |
| format | Removes trailing whitespace, deduplicates blank lines |

## Output Format

Issues are displayed with file, line number, linter name, and message:

```
  ci.yml:15: (versions) Action actions/checkout@v3 uses version tag 'v3' instead of commit hash
```

## Categories

### Security Linters

- **secrets**: Detects hardcoded credentials
- **injection**: Detects shell injection vulnerabilities
- **permissions**: Ensures least-privilege permissions

### Code Quality Linters

- **versions**: Enforces pinned action versions
- **format**: Maintains consistent formatting
- **style**: Enforces naming conventions and best practices
