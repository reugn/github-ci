---
title: secrets
parent: Linters
nav_order: 4
layout: default
render_with_liquid: false
---

# secrets

Detects hardcoded secrets and sensitive information in workflow files.

## Why This Matters

Hardcoded secrets in workflow files:

- **Expose credentials publicly**: GitHub repos (even private ones) may become public
- **Cannot be rotated easily**: Changing requires updating all workflow files
- **Appear in logs**: May be printed in workflow run logs
- **Violate security policies**: Most organizations prohibit hardcoded secrets

## What It Detects

| Pattern | Description |
|---------|-------------|
| AWS Access Keys | `AKIA...` pattern |
| AWS Secret Keys | 40-character base64 strings after `aws_secret` |
| GitHub Tokens | `ghp_`, `gho_`, `ghu_`, `ghs_`, `ghr_` prefixes |
| Private Keys | `-----BEGIN ... PRIVATE KEY-----` |
| Generic API Keys | `api_key`, `apikey`, `api-key` with values |
| Generic Secrets | `secret`, `password`, `token` with values |

## Example Output

```
ci.yml:15: (secrets) Possible hardcoded AWS access key detected
ci.yml:22: (secrets) Possible hardcoded GitHub token detected
ci.yml:30: (secrets) Possible hardcoded private key detected
```

## Auto-fix

**Not supported** - Secrets must be removed and replaced with proper secret references manually.

## How to Fix

### Use GitHub Secrets

```yaml
# Bad
env:
  AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE

# Good
env:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
```

### Use Environment Secrets

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production  # Uses secrets from 'production' environment
    steps:
      - run: deploy.sh
        env:
          API_KEY: ${{ secrets.API_KEY }}
```

### Use OIDC for Cloud Providers

For AWS, Azure, and GCP, use OIDC instead of long-lived credentials:

```yaml
permissions:
  id-token: write
  contents: read

steps:
  - uses: aws-actions/configure-aws-credentials@v4
    with:
      role-to-assume: arn:aws:iam::123456789:role/my-role
      aws-region: us-east-1
```

## Common Patterns Detected

### AWS Credentials

```yaml
# Detected as potential secret
AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

### GitHub Tokens

```yaml
# Detected as potential secret
GITHUB_TOKEN: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### API Keys

```yaml
# Detected as potential secret
api_key: sk-1234567890abcdef
API_KEY: "some-secret-value"
```

### Private Keys

```yaml
# Detected as potential secret
SSH_KEY: |
  -----BEGIN RSA PRIVATE KEY-----
  MIIEpAIBAAKCAQEA...
  -----END RSA PRIVATE KEY-----
```

## False Positives

Some patterns may trigger false positives:

- Example/placeholder values in comments
- Base64-encoded non-secret data
- Test fixtures with fake credentials

If you encounter false positives, you can:

1. Disable the secrets linter for specific projects
2. Use obviously fake values in examples (e.g., `AKIAEXAMPLE123456789`)

## Security Best Practices

1. **Never commit real secrets** - Even if you remove them later, they're in git history
2. **Use GitHub Secrets** - Encrypted and only exposed to workflows
3. **Rotate if exposed** - If a secret was ever in code, treat it as compromised
4. **Use OIDC** - Prefer short-lived credentials over long-lived secrets
5. **Audit regularly** - Run this linter in CI to catch accidental commits

## See Also

- [GitHub Docs: Encrypted Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [GitHub Docs: OIDC](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
