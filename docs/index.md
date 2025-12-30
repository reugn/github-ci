---
title: Home
nav_order: 1
layout: default
---

# github-ci

A CLI tool for managing GitHub Actions workflows. It helps you lint workflows for best practices and automatically upgrade actions to their latest versions.

## Features

- **Lint Workflows**: Check workflows for best practices with multiple configurable linters
- **Auto-fix Issues**: Automatically fix formatting issues and replace version tags with commit hashes
- **Upgrade Actions**: Discover and upgrade GitHub Actions to their latest versions based on semantic versioning patterns
- **Config Management**: Configure linters and version update patterns via `.github-ci.yaml`

## Available Linters

| Linter | Description | Auto-fix |
|--------|-------------|----------|
| [permissions](linters/permissions) | Missing permissions configuration | ✗ |
| [versions](linters/versions) | Actions using version tags instead of commit hashes | ✓ |
| [format](linters/format) | Formatting issues (indentation, line length, whitespace) | ✓ |
| [secrets](linters/secrets) | Hardcoded secrets and sensitive information | ✗ |
| [injection](linters/injection) | Shell injection vulnerabilities from untrusted input | ✗ |
| [style](linters/style) | Naming conventions and style best practices | ✗ |

## Quick Start

```bash
# Install
go install github.com/reugn/github-ci/cmd/github-ci@latest

# Initialize config
github-ci init

# Lint workflows
github-ci lint

# Auto-fix issues
github-ci lint --fix

# Upgrade actions
github-ci upgrade --dry-run
github-ci upgrade
```

## Requirements

- Go 1.24 or later
- Internet connection (for fetching action versions from GitHub API)

## License

Licensed under the Apache 2.0 License.
