---
title: lint
parent: Usage
nav_order: 2
layout: default
---

# lint Command

Analyze workflows for common issues using configurable linters.

## Synopsis

```bash
github-ci lint [flags]
```

## Description

The `lint` command scans GitHub Actions workflow files and checks for various issues based on enabled linters:

- **permissions**: Missing permissions configuration
- **versions**: Actions using version tags instead of commit hashes
- **format**: Formatting issues (indentation, line length, trailing whitespace)
- **secrets**: Hardcoded secrets and sensitive information
- **injection**: Shell injection vulnerabilities from untrusted input
- **style**: Naming conventions and style best practices

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--fix` | `false` | Automatically fix issues where possible |
| `--path` | `.github/workflows` | Path to workflow directory or file |
| `--config` | `.github-ci.yaml` | Path to configuration file |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No issues found |
| 1 | Issues found (configurable via `issues-exit-code`) |

The exit code when issues are found can be customized in the configuration file.

## Examples

### Basic Linting

```bash
$ github-ci lint

Issues:
  ci.yml: (permissions) Workflow is missing permissions configuration
  ci.yml:15: (versions) Action actions/checkout@v3 uses version tag 'v3' instead of commit hash
  ci.yml:22: (format) Line exceeds maximum length of 120 characters

Run with --fix to automatically fix some issues

3 issue(s).
```

### Auto-fix Issues

```bash
$ github-ci lint --fix

Fixed:
  ci.yml:15: (versions) Action actions/checkout@v3 uses version tag 'v3' instead of commit hash

Issues:
  ci.yml: (permissions) Workflow is missing permissions configuration

1 issue(s).
```

### Lint Specific File

```bash
github-ci lint --path .github/workflows/ci.yml
```

## Auto-fix Support

Not all linters support `--fix`. Currently supported:

| Linter | Auto-fix |
|--------|----------|
| versions | ✓ Replaces version tags with commit hashes |
| format | ✓ Fixes trailing whitespace and multiple blank lines |
| permissions | ✗ |
| secrets | ✗ |
| injection | ✗ |
| style | ✗ |

### Fix Transformation Example

When using `--fix`, version tags are replaced with commit hashes:

```yaml
# Before
- uses: actions/checkout@v3

# After
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v3.5.0
```

If a major version is specified (e.g., `v3`), the tool finds the latest minor version in that series and uses its commit hash with the version in a comment.

## Output Format

Issues are displayed with:
- File name
- Line number (when applicable)
- Linter name in parentheses
- Issue message

```
  file.yml:15: (linter) Message describing the issue
```

## See Also

- [Linters](../linters/) - Detailed documentation for each linter
- [Configuration](../configuration/) - Configure which linters to enable
