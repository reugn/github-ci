---
title: Installation
nav_order: 2
layout: default
render_with_liquid: false
---

# Installation

## Using Go Install

The recommended way to install `github-ci`:

```bash
go install github.com/reugn/github-ci/cmd/github-ci@latest
```

Make sure `$GOPATH/bin` or `$GOBIN` is in your `$PATH`.

## From Releases

Download the latest binary for your platform from [Releases](https://github.com/reugn/github-ci/releases).

## From Source

Clone the repository and build manually:

```bash
git clone https://github.com/reugn/github-ci.git
cd github-ci
go build -o github-ci ./cmd/github-ci
sudo mv github-ci /usr/local/bin/
```

## Verify Installation

```bash
github-ci --version
github-ci --help
```

You should see the version and the available commands.

## Authentication

The tool can optionally use a GitHub personal access token for authentication, which provides:

- **Higher rate limits**: 5,000 requests/hour (authenticated) vs 60 requests/hour (unauthenticated)
- **Access to private repositories**: If you need to work with private action repositories

### Setting up Authentication

Provide a GitHub token via the `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=ghp_your_token_here
github-ci lint
```

The token is automatically read from the environment variable during initialization. No additional configuration is required.

{: .note }
> The token is optional. The tool will work without it, but you may hit rate limits more quickly when processing many workflows.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Lint workflows
  run: |
    go install github.com/reugn/github-ci/cmd/github-ci@latest
    github-ci lint
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
github-ci lint --path .github/workflows
```
