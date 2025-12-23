package actions

import "fmt"

// VersionKey represents a cache key for version lookups.
type VersionKey struct {
	Owner          string
	Repo           string
	CurrentVersion string // Empty for unconstrained lookups
	Pattern        string // Empty for unconstrained lookups
}

// NewConstrainedKey creates a key for constrained version lookups.
// Constrained lookups consider the current version and pattern.
func NewConstrainedKey(owner, repo, currentVersion, pattern string) VersionKey {
	return VersionKey{
		Owner:          owner,
		Repo:           repo,
		CurrentVersion: currentVersion,
		Pattern:        pattern,
	}
}

// NewUnconstrainedKey creates a key for unconstrained version lookups.
// Unconstrained lookups just get the latest version for a repo.
func NewUnconstrainedKey(owner, repo string) VersionKey {
	return VersionKey{
		Owner: owner,
		Repo:  repo,
	}
}

// String returns the string representation of the cache key.
func (k VersionKey) String() string {
	if !k.IsConstrained() {
		return fmt.Sprintf("%s/%s", k.Owner, k.Repo)
	}
	return fmt.Sprintf("%s/%s:%s:%s", k.Owner, k.Repo, k.CurrentVersion, k.Pattern)
}

// IsConstrained returns true if this is a constrained key.
func (k VersionKey) IsConstrained() bool {
	return k.CurrentVersion != "" || k.Pattern != ""
}
