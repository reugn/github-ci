package actions

// Resolver defines the interface for GitHub Actions operations.
type Resolver interface {
	GetCommitHash(owner, repo, ref string) (string, error)
	GetLatestVersion(owner, repo, currentVersion, versionPattern string) (string, string, error)
	GetLatestVersionUnconstrained(owner, repo string) (string, string, error)
	GetTagForCommit(owner, repo, commitHash string) (string, error)
	GetLatestMinorVersion(owner, repo, majorVersion string) (string, string, error)
	GetCacheStats() CacheStats
}
