package actions

// MockResolver is a mock implementation of the Resolver interface for testing.
type MockResolver struct {
	GetCommitHashFunc            func(owner, repo, ref string) (string, error)
	GetLatestVersionFunc         func(owner, repo, currentVersion, versionPattern string) (string, string, error)
	GetLatestVersionUnconstrFunc func(owner, repo string) (string, string, error)
	GetTagForCommitFunc          func(owner, repo, commitHash string) (string, error)
	GetLatestMinorVersionFunc    func(owner, repo, majorVersion string) (string, string, error)
}

// Ensure MockResolver implements Resolver
var _ Resolver = (*MockResolver)(nil)

func (m *MockResolver) GetCommitHash(owner, repo, ref string) (string, error) {
	if m.GetCommitHashFunc != nil {
		return m.GetCommitHashFunc(owner, repo, ref)
	}
	return "", nil
}

func (m *MockResolver) GetLatestVersion(owner, repo, currentVersion, versionPattern string) (string, string, error) {
	if m.GetLatestVersionFunc != nil {
		return m.GetLatestVersionFunc(owner, repo, currentVersion, versionPattern)
	}
	return "", "", nil
}

func (m *MockResolver) GetLatestVersionUnconstrained(owner, repo string) (string, string, error) {
	if m.GetLatestVersionUnconstrFunc != nil {
		return m.GetLatestVersionUnconstrFunc(owner, repo)
	}
	return "", "", nil
}

func (m *MockResolver) GetTagForCommit(owner, repo, commitHash string) (string, error) {
	if m.GetTagForCommitFunc != nil {
		return m.GetTagForCommitFunc(owner, repo, commitHash)
	}
	return "", nil
}

func (m *MockResolver) GetLatestMinorVersion(owner, repo, majorVersion string) (string, string, error) {
	if m.GetLatestMinorVersionFunc != nil {
		return m.GetLatestMinorVersionFunc(owner, repo, majorVersion)
	}
	return "", "", nil
}

func (m *MockResolver) GetCacheStats() CacheStats {
	return CacheStats{} // Mock always returns zero stats
}
