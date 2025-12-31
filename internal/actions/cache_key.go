package actions

import "fmt"

// VersionKey represents a cache key for version lookups.
type VersionKey struct {
	Owner   string
	Repo    string
	Ref     string // Current version reference (e.g., "v1.2.0"); empty for unconstrained
	Pattern string // Version constraint (e.g., "^1.0.0"); empty for unconstrained
}

// NewConstrainedKey creates a key for constrained version lookups.
func NewConstrainedKey(owner, repo, ref, pattern string) VersionKey {
	return VersionKey{
		Owner:   owner,
		Repo:    repo,
		Ref:     ref,
		Pattern: pattern,
	}
}

// NewUnconstrainedKey creates a key for unconstrained version lookups.
// Unconstrained lookups get the absolute latest version for a repo.
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
	return fmt.Sprintf("%s/%s:%s:%s", k.Owner, k.Repo, k.Ref, k.Pattern)
}

// IsConstrained returns true if this is a constrained key.
func (k VersionKey) IsConstrained() bool {
	return k.Ref != "" || k.Pattern != ""
}
