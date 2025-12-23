---
title: Upgrade Settings
parent: Configuration
nav_order: 3
layout: default
---

# Upgrade Configuration

The `upgrade` section controls how actions are upgraded.

## Options

```yaml
upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ^1.0.0
    actions/setup-go:
      version: ~1.0.0
```

### version

Controls the format of action references after upgrade.

| Value | Output | Description |
|-------|--------|-------------|
| `tag` | `@v5.2.0` | Full version tag (default) |
| `major` | `@v5` | Major version only |
| `hash` | `@abc123... # v5.2.0` | Commit hash with version comment |

### actions

Per-action version patterns controlling which versions are allowed.

## Version Patterns

### Caret (`^`) - Allow Minor Updates

```yaml
actions:
  actions/checkout:
    version: ^1.0.0  # Allows 1.x.x but not 2.x.x
```

| Pattern | Allowed | Not Allowed |
|---------|---------|-------------|
| `^1.0.0` | `1.0.1`, `1.2.0`, `1.99.0` | `2.0.0` |
| `^2.0.0` | `2.0.1`, `2.5.0` | `3.0.0` |

{: .note }
> `^1.0.0` is special: it allows any version >= 1.0.0, including 2.x, 3.x, etc. This matches npm semver behavior.

### Tilde (`~`) - Allow Patch Updates Only

```yaml
actions:
  actions/checkout:
    version: ~1.2.0  # Allows 1.2.x but not 1.3.x
```

| Pattern | Allowed | Not Allowed |
|---------|---------|-------------|
| `~1.2.0` | `1.2.1`, `1.2.5` | `1.3.0`, `2.0.0` |
| `~2.5.0` | `2.5.1`, `2.5.99` | `2.6.0` |

### Default Behavior

Actions not explicitly configured use `^1.0.0`, allowing any newer version.

## Examples

### Conservative Upgrades

Only allow patch updates for stability:

```yaml
upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ~4.0.0
    actions/setup-go:
      version: ~5.0.0
```

### Major Version Pinning

Use major version tags for cleaner workflow files:

```yaml
upgrade:
  version: major
  actions:
    actions/checkout:
      version: ^1.0.0
```

Result:
```yaml
- uses: actions/checkout@v4
```

### Security-Focused (Hash Pinning)

Pin to exact commits for maximum security:

```yaml
upgrade:
  version: hash
  actions:
    actions/checkout:
      version: ^1.0.0
```

Result:
```yaml
- uses: actions/checkout@8f4b7f84856dbbe3f95729c4cd48d901b28810a  # v4.1.1
```

### Mixed Strategies

Different patterns for different actions:

```yaml
upgrade:
  version: tag
  actions:
    # Critical actions - patch updates only
    actions/checkout:
      version: ~4.0.0
    
    # Less critical - minor updates allowed
    actions/cache:
      version: ^4.0.0
    
    # Third-party - more conservative
    docker/build-push-action:
      version: ~5.0.0
```

## Upgrade Process

1. **Discover**: Scan workflows for action usages
2. **Resolve**: Get current version (resolve hash to tag if needed)
3. **Fetch**: Get latest version from GitHub API
4. **Compare**: Check if update matches version pattern
5. **Update**: Apply update based on `version` format setting

## See Also

- [upgrade command](../usage/upgrade) - Running the upgrade command
- [init command](../usage/init) - Creating initial configuration
