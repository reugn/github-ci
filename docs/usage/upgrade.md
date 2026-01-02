---
title: upgrade
parent: Usage
nav_order: 3
layout: default
---

# upgrade Command

Upgrade GitHub Actions in workflows to their latest versions.

## Synopsis

```bash
github-ci upgrade [flags]
```

## Description

The `upgrade` command checks for newer versions of actions in all workflows and updates them based on configured version constraints.

This command:
1. Scans all workflows to discover actions
2. Updates `.github-ci.yaml` if it exists (use `init` command to create one)
3. Checks for newer versions of each action
4. Updates actions based on version constraints defined in the config

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Print updates without modifying files |
| `--path` | `.github/workflows` | Path to workflow directory or file |
| `--config` | `.github-ci.yaml` | Path to configuration file |

## Examples

### Preview Updates

```bash
$ github-ci upgrade --dry-run

Would update 2 action(s):

  .github/workflows/ci.yml:15
    actions/checkout@v3
    → actions/checkout@v4.1.1

  .github/workflows/ci.yml:22
    actions/setup-go@v4
    → actions/setup-go@v5.0.0
```

### Apply Updates

```bash
$ github-ci upgrade

✓ Upgrade completed successfully

GitHub API: 4 call(s), 2 from cache
```

## Version Format

The `upgrade.format` config option controls how actions are referenced after upgrade:

| Format | Example | Description |
|--------|---------|-------------|
| `tag` | `@v5.2.0` | Full version tag (default) |
| `major` | `@v5` | Major version only |
| `hash` | `@abc123...` | Commit hash with version comment |

### Tag Format (Default)

```yaml
- uses: actions/checkout@v4.1.1
```

### Major Format

```yaml
- uses: actions/checkout@v4
```

### Hash Format

```yaml
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v4.1.1
```

## Version Constraints

Control which versions are allowed for each action:

| Constraint | Behavior | Example |
|---------|----------|---------|
| `^1.0.0` | Same major, any minor/patch | `1.x.x` |
| `~1.2.0` | Same major.minor, any patch | `1.2.x` |

See [Upgrade Configuration](../configuration/upgrade) for details.

## Warnings

The upgrade command may show warnings in certain situations:

### Unreleased Commit Hash

```
⚠ Warning: cannot resolve hash abc123... to a tag (may be unreleased commit)
```

This appears when:
- A workflow uses a commit hash that doesn't correspond to any release tag
- The hash might be from an unreleased commit that is newer than the upgrade target

## See Also

- [Configuration](../configuration/) - Configure version constraints
- [init](init) - Create configuration file
