# github-ci

[![Build](https://github.com/reugn/github-ci/actions/workflows/build.yml/badge.svg)](https://github.com/reugn/github-ci/actions/workflows/build.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/reugn/github-ci)](https://pkg.go.dev/github.com/reugn/github-ci)
[![Go Report Card](https://goreportcard.com/badge/github.com/reugn/github-ci)](https://goreportcard.com/report/github.com/reugn/github-ci)
[![codecov](https://codecov.io/gh/reugn/github-ci/graph/badge.svg?token=CTG5JY369G)](https://codecov.io/gh/reugn/github-ci)

A CLI tool for managing GitHub Actions workflows. It helps lint workflows for best practices and automatically upgrade actions to their latest versions.

## Documentation

- [Index](https://reugn.github.io/github-ci/)
- [Installation](https://reugn.github.io/github-ci/install)
- [Usage Guide](https://reugn.github.io/github-ci/usage/)
- [Configuration](https://reugn.github.io/github-ci/configuration/)
- [Linters Reference](https://reugn.github.io/github-ci/linters/)

## Features

- **Lint Workflows**: Check workflows for best practices with multiple configurable linters:
  - **permissions**: Missing permissions configuration
  - **versions**: Actions using version tags instead of commit hashes
  - **format**: Formatting issues (indentation, line length, trailing whitespace)
  - **secrets**: Hardcoded secrets and sensitive information
  - **injection**: Shell injection vulnerabilities from untrusted input
- **Auto-fix Issues**: Automatically fix formatting issues and replace version tags with commit hashes
- **Upgrade Actions**: Discover and upgrade GitHub Actions to their latest versions based on semantic versioning patterns
- **Config Management**: Configure linters and version update patterns via `.github-ci.yaml`

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

# Upgrade actions (preview)
github-ci upgrade --dry-run

# Upgrade actions
github-ci upgrade
```

## Installation

### Using Go Install

```bash
go install github.com/reugn/github-ci/cmd/github-ci@latest
```

Make sure `$GOPATH/bin` or `$GOBIN` is in your `$PATH`.

### From Source

```bash
git clone https://github.com/reugn/github-ci.git
cd github-ci
go build -o github-ci ./cmd/github-ci
sudo mv github-ci /usr/local/bin/
```

## Example Usage

### Linting Workflows

```bash
$ github-ci lint

Issues:
  ci.yml: (permissions) Workflow is missing permissions configuration
  ci.yml:15: (versions) Action actions/checkout@v3 uses version tag 'v3' instead of commit hash

Run with --fix to automatically fix some issues

2 issue(s).
```

### Auto-fixing Issues

```bash
$ github-ci lint --fix

Fixed:
  ci.yml:15: (versions) Action actions/checkout@v3 uses version tag 'v3' instead of commit hash

Issues:
  ci.yml: (permissions) Workflow is missing permissions configuration

1 issue(s).
```

### Upgrading Actions

```bash
$ github-ci upgrade --dry-run

Would update 2 action(s):

  .github/workflows/ci.yml:15
    actions/checkout@v3
    → actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 (v4.1.1)

  .github/workflows/ci.yml:22
    actions/setup-go@v4
    → actions/setup-go@0c52d547c9bc32b1aa3301fd7a9cb496313a4491 (v5.0.0)
```

## Configuration

Create a `.github-ci.yaml` file to configure the tool:

```yaml
run:
  timeout: 5m
  issues-exit-code: 1

linters:
  default: all
  enable:
    - permissions
    - versions
    - format
  settings:
    format:
      indent-width: 2
      max-line-length: 120

upgrade:
  version: tag  # or 'major', 'hash'
  actions:
    actions/checkout:
      version: ^1.0.0
```

See the [Configuration Guide](https://reugn.github.io/github-ci/configuration/) for all options.

## Authentication

For higher rate limits and private repository access, set a GitHub token:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

## Requirements

- Go 1.24 or later
- Internet connection (for GitHub API access)

## License

Licensed under the Apache 2.0 License.
