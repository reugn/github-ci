---
title: versions
parent: Linters
nav_order: 2
layout: default
---

# versions

Checks that actions use commit hashes instead of version tags.

## Why This Matters

Using commit hashes instead of version tags:

- **Prevents supply chain attacks**: Tags can be moved to point to malicious code
- **Ensures reproducibility**: The same commit always runs
- **Audit trail**: Know exactly what code ran in each workflow run

## What It Detects

Actions using version tags (`@v3`, `@v3.5.0`) instead of commit hashes.

### ❌ Bad

```yaml
- uses: actions/checkout@v4
- uses: actions/setup-go@v5.0.0
```

### ✅ Good

```yaml
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v4.1.1
- uses: actions/setup-go@0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32  # v5.0.0
```

## Example Output

```
ci.yml:15: (versions) Action actions/checkout@v4 uses version tag 'v4' instead of commit hash
```

## Auto-fix

**Supported** with `--fix`:

```bash
github-ci lint --fix
```

The fix:
1. Resolves the version tag to a commit hash
2. If a major version (e.g., `v4`), finds the latest minor version first
3. Replaces the tag with the commit hash
4. Adds a comment with the version for reference

### Example Transformation

```yaml
# Before
- uses: actions/checkout@v4

# After
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v4.1.1
```

## Major Version Resolution

When you specify a major version like `v4`, the tool:

1. Fetches all tags for the action
2. Finds the latest version in that major series (e.g., `v4.1.1`)
3. Gets the commit hash for that version
4. Comments with the actual version

This ensures you get the latest stable release within that major version.

## Verifying Hashes

To manually verify a commit hash:

```bash
# Check what version a hash corresponds to
curl -s https://api.github.com/repos/actions/checkout/git/refs/tags | \
  jq '.[] | select(.object.sha | startswith("8f4b7f8"))'
```

## Renovate/Dependabot Compatibility

Both Renovate and Dependabot can update hash-pinned actions:

```yaml
# Dependabot will update this
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v4.1.1
```

They read the version from the comment to understand the current version.

## See Also

- [upgrade command](../usage/upgrade) - Automatically upgrade actions
- [Upgrade Configuration](../configuration/upgrade) - Version patterns
