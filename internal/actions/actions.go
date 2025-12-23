package actions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v80/github"
	"github.com/reugn/github-ci/internal/version"
	"golang.org/x/oauth2"
)

const (
	timeout = 10 * time.Second
	// GitHubTokenEnvVar is the environment variable for GitHub authentication.
	GitHubTokenEnvVar = "GITHUB_TOKEN" //nolint:gosec // Not a credential, just env var name
)

// Resolver defines the interface for GitHub Actions operations.
type Resolver interface {
	GetCommitHash(owner, repo, ref string) (string, error)
	GetLatestVersion(owner, repo, currentVersion, versionPattern string) (string, string, error)
	GetLatestVersionUnconstrained(owner, repo string) (string, string, error)
	GetTagForCommit(owner, repo, commitHash string) (string, error)
	GetLatestMinorVersion(owner, repo, majorVersion string) (string, string, error)
	GetCacheStats() CacheStats
}

// tagInfo holds tag name and commit hash.
type tagInfo struct {
	tag  string
	hash string
}

// Client implements the Resolver interface.
type Client struct {
	ctx        context.Context
	github     *github.Client
	cache      *Cache
	clientOnce sync.Once
}

// NewClient creates a new Client instance with background context.
func NewClient() *Client {
	return NewClientWithContext(context.Background())
}

// NewClientWithContext creates a new Client instance with the provided context.
func NewClientWithContext(ctx context.Context) *Client {
	return &Client{
		ctx:   ctx,
		cache: NewCache(),
	}
}

// NewClientWithCache creates a Client with a shared cache.
func NewClientWithCache(ctx context.Context, cache *Cache) *Client {
	return &Client{
		ctx:   ctx,
		cache: cache,
	}
}

// getGitHubClient returns the GitHub client, initializing it lazily (thread-safe).
func (c *Client) getGitHubClient() *github.Client {
	c.clientOnce.Do(func() {
		httpClient := &http.Client{Timeout: timeout}

		if token := os.Getenv(GitHubTokenEnvVar); token != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			httpClient = oauth2.NewClient(c.ctx, ts)
			httpClient.Timeout = timeout
		}

		c.github = github.NewClient(httpClient)
	})
	return c.github
}

// ClearCache clears the version lookup cache.
func (c *Client) ClearCache() {
	c.cache.Clear()
}

// GetCacheStats returns the cache usage statistics.
func (c *Client) GetCacheStats() CacheStats {
	return c.cache.Stats()
}

// ActionInfo represents a parsed GitHub Action reference.
type ActionInfo struct {
	Owner      string
	Repo       string
	Ref        string
	CommitHash string
	Tag        string
}

// ParseActionUses parses "owner/repo@ref" into ActionInfo.
func ParseActionUses(uses string) (*ActionInfo, error) {
	atIdx := strings.LastIndex(uses, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid action format: %s", uses)
	}

	actionPath := uses[:atIdx]
	ref := uses[atIdx+1:]

	slashIdx := strings.Index(actionPath, "/")
	if slashIdx == -1 {
		return nil, fmt.Errorf("invalid action path: %s", actionPath)
	}

	return &ActionInfo{
		Owner: actionPath[:slashIdx],
		Repo:  actionPath[slashIdx+1:],
		Ref:   ref,
	}, nil
}

// IsCommitHash checks if a reference is a 40-char hex commit hash.
func IsCommitHash(ref string) bool {
	if len(ref) != 40 {
		return false
	}
	for _, c := range ref {
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}
	return true
}

// IsMajorVersionOnly checks if ref is only a major version (e.g., "v3" or "3").
func IsMajorVersionOnly(ref string) bool {
	ref = version.Normalize(ref)
	if ref == "" {
		return false
	}
	for _, c := range ref {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// GetCommitHash resolves a Git reference to its commit hash.
func (c *Client) GetCommitHash(owner, repo, ref string) (string, error) {
	if IsCommitHash(ref) {
		return ref, nil
	}

	ref = strings.TrimPrefix(ref, "refs/")

	// Handle major version only (e.g., "v3")
	if IsMajorVersionOnly(ref) {
		if _, hash, err := c.GetLatestMinorVersion(owner, repo, ref); err == nil {
			return hash, nil
		}
	}

	client := c.getGitHubClient()

	// Try as tag first
	if gitRef, _, err := client.Git.GetRef(c.ctx, owner, repo, "refs/tags/"+ref); err == nil && gitRef.Object != nil {
		return gitRef.Object.GetSHA(), nil
	}

	// Fall back to branch
	gitRef, _, err := client.Git.GetRef(c.ctx, owner, repo, "refs/heads/"+ref)
	if err != nil {
		return "", fmt.Errorf("failed to fetch ref %s: %w", ref, err)
	}
	if gitRef == nil || gitRef.Object == nil {
		return "", fmt.Errorf("ref not found: %s", ref)
	}

	return gitRef.Object.GetSHA(), nil
}

// GetLatestVersion fetches the latest compatible tag and commit hash.
// Results are cached.
func (c *Client) GetLatestVersion(owner, repo, currentVersion, versionPattern string) (string, string, error) {
	key := NewConstrainedKey(owner, repo, currentVersion, versionPattern)

	if result, ok := c.cache.GetConstrained(key); ok {
		return result.Tag, result.Hash, result.Err
	}

	tags, err := c.fetchMatchingTags(owner, repo, versionPattern)
	if err != nil {
		c.cache.SetConstrained(key, NewVersionResult("", "", err))
		return "", "", err
	}

	if len(tags) == 0 {
		err := fmt.Errorf("no compatible tags found for pattern %s", versionPattern)
		c.cache.SetConstrained(key, NewVersionResult("", "", err))
		return "", "", err
	}

	// Find semantically latest version
	latest := tags[0]
	for _, t := range tags[1:] {
		if version.Compare(t.tag, latest.tag) > 0 {
			latest = t
		}
	}

	c.cache.SetConstrained(key, NewVersionResult(latest.tag, latest.hash, nil))
	return latest.tag, latest.hash, nil
}

// fetchMatchingTags retrieves all tags matching the version pattern.
func (c *Client) fetchMatchingTags(owner, repo, pattern string) ([]tagInfo, error) {
	client := c.getGitHubClient()
	opts := &github.ListOptions{PerPage: 100}

	var matching []tagInfo
	for {
		tags, resp, err := client.Repositories.ListTags(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		for _, tag := range tags {
			if matchesVersionPattern(tag.GetName(), pattern) {
				matching = append(matching, tagInfo{
					tag:  tag.GetName(),
					hash: tag.GetCommit().GetSHA(),
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return matching, nil
}

// GetLatestVersionUnconstrained fetches the semantically latest version.
// First tries GitHub Releases API (single call), then falls back to tag pagination.
// Results are cached.
func (c *Client) GetLatestVersionUnconstrained(owner, repo string) (string, string, error) {
	key := NewUnconstrainedKey(owner, repo)

	if result, ok := c.cache.GetUnconstrained(key); ok {
		return result.Tag, result.Hash, result.Err
	}

	// Try GitHub Releases API first (most repos use releases)
	if tag, hash, ok := c.tryGetLatestRelease(owner, repo); ok {
		c.cache.SetUnconstrained(key, NewVersionResult(tag, hash, nil))
		return tag, hash, nil
	}

	// Fall back to tag pagination for repos without releases
	return c.getLatestVersionFromTags(owner, repo, key)
}

// tryGetLatestRelease attempts to get the latest version via GitHub Releases API.
// Returns the tag name, commit hash, and true if successful.
func (c *Client) tryGetLatestRelease(owner, repo string) (string, string, bool) {
	release, _, err := c.getGitHubClient().Repositories.GetLatestRelease(c.ctx, owner, repo)
	if err != nil || release == nil || release.TagName == nil {
		return "", "", false
	}

	tag := release.GetTagName()
	// Get the commit hash for this tag
	hash, err := c.GetCommitHash(owner, repo, tag)
	if err != nil {
		return "", "", false
	}

	return tag, hash, true
}

// getLatestVersionFromTags finds the semantically latest tag by paginating through all tags.
func (c *Client) getLatestVersionFromTags(owner, repo string, key VersionKey) (string, string, error) {
	client := c.getGitHubClient()
	opts := &github.ListOptions{PerPage: 100}

	var latest *tagInfo
	for {
		tags, resp, err := client.Repositories.ListTags(c.ctx, owner, repo, opts)
		if err != nil {
			err = fmt.Errorf("failed to fetch tags: %w", err)
			c.cache.SetUnconstrained(key, NewVersionResult("", "", err))
			return "", "", err
		}

		for _, tag := range tags {
			name := tag.GetName()
			if latest == nil || version.Compare(name, latest.tag) > 0 {
				latest = &tagInfo{tag: name, hash: tag.GetCommit().GetSHA()}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if latest == nil {
		err := fmt.Errorf("no tags found")
		c.cache.SetUnconstrained(key, NewVersionResult("", "", err))
		return "", "", err
	}

	c.cache.SetUnconstrained(key, NewVersionResult(latest.tag, latest.hash, nil))
	return latest.tag, latest.hash, nil
}

// GetTagForCommit finds which tag points to the given commit hash.
func (c *Client) GetTagForCommit(owner, repo, commitHash string) (string, error) {
	client := c.getGitHubClient()
	opts := &github.ListOptions{PerPage: 100}

	for {
		tags, resp, err := client.Repositories.ListTags(c.ctx, owner, repo, opts)
		if err != nil {
			return "", fmt.Errorf("failed to fetch tags: %w", err)
		}

		for _, tag := range tags {
			if tag.GetCommit().GetSHA() == commitHash {
				return tag.GetName(), nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return "", nil
}

// GetLatestMinorVersion finds the latest minor version for a major version.
// For example, "v3" might return "v3.5.2".
func (c *Client) GetLatestMinorVersion(owner, repo, majorVersion string) (string, string, error) {
	majorVersion = version.Normalize(majorVersion)
	prefix := "v" + majorVersion + "."

	client := c.getGitHubClient()
	opts := &github.ListOptions{PerPage: 100}

	var latest *tagInfo
	for {
		tags, resp, err := client.Repositories.ListTags(c.ctx, owner, repo, opts)
		if err != nil {
			return "", "", fmt.Errorf("failed to fetch tags: %w", err)
		}

		for _, tag := range tags {
			name := tag.GetName()
			// Match "vX." or exact "vX"
			if strings.HasPrefix(name, prefix) || name == "v"+majorVersion {
				info := &tagInfo{tag: name, hash: tag.GetCommit().GetSHA()}
				if latest == nil || version.Compare(name, latest.tag) > 0 {
					latest = info
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if latest == nil {
		return "", "", fmt.Errorf("no tags found for major version v%s", majorVersion)
	}

	return latest.tag, latest.hash, nil
}

// matchesVersionPattern checks if a tag matches the version pattern.
func matchesVersionPattern(tagVersion, pattern string) bool {
	if pattern == "" {
		return true
	}

	if after, ok := strings.CutPrefix(pattern, "^"); ok {
		patternMajor := version.ExtractMajor(after)
		tagMajor := version.ExtractMajor(tagVersion)

		// ^1.0.0 allows any version >= 1
		if patternMajor == 1 {
			return tagMajor >= 1
		}
		return tagMajor == patternMajor
	}

	if after, ok := strings.CutPrefix(pattern, "~"); ok {
		patternMajor, patternMinor := version.ExtractMajorMinor(after)
		tagMajor, tagMinor := version.ExtractMajorMinor(tagVersion)
		return tagMajor == patternMajor && tagMinor == patternMinor
	}

	return false
}
