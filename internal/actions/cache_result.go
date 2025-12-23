package actions

// VersionResult represents a cached version lookup result.
type VersionResult struct {
	Tag  string
	Hash string
	Err  error
}

// NewVersionResult creates a new VersionResult.
func NewVersionResult(tag, hash string, err error) VersionResult {
	return VersionResult{
		Tag:  tag,
		Hash: hash,
		Err:  err,
	}
}

// IsError returns true if the result contains an error.
func (r VersionResult) IsError() bool {
	return r.Err != nil
}
