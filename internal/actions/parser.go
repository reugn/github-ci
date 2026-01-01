package actions

import (
	"fmt"
	"strings"

	"github.com/reugn/github-ci/internal/version"
)

// ActionInfo represents a parsed GitHub Action reference.
type ActionInfo struct {
	Owner string
	Repo  string
	Path  string // Subdirectory path for composite actions (e.g., "upload-sarif")
	Ref   string // Git reference: tag (e.g., "v2"), commit hash, or branch name
}

// Name returns the normalized action name (owner/repo or owner/repo/path).
func (a *ActionInfo) Name() string {
	if a.Path != "" {
		return a.Owner + "/" + a.Repo + "/" + a.Path
	}
	return a.Owner + "/" + a.Repo
}

// FormatUses returns the action uses string with the given ref (e.g., "owner/repo@ref").
func (a *ActionInfo) FormatUses(ref string) string {
	return a.Name() + "@" + ref
}

// IsAtLatest checks if the current ref points to the latest version.
// Works for hashes, tags, and major versions.
func (a *ActionInfo) IsAtLatest(latestTag, latestHash string) bool {
	if IsCommitHash(a.Ref) {
		return a.Ref == latestHash
	}
	// Exact tag match
	if a.Ref == latestTag {
		return true
	}
	// Major version match (e.g., v4 or 4 matches v4.2.1)
	if IsMajorVersionOnly(a.Ref) {
		major := version.Normalize(a.Ref)
		prefix := "v" + major + "."
		return strings.HasPrefix(latestTag, prefix) || latestTag == "v"+major
	}
	return false
}

// NeedsFormatChange checks if the current ref format differs from the desired format.
func (a *ActionInfo) NeedsFormatChange(desiredFormat string) bool {
	isHash := IsCommitHash(a.Ref)

	switch desiredFormat {
	case "hash":
		return !isHash
	case "major":
		if isHash {
			return true
		}
		return !IsMajorVersionOnly(a.Ref)
	default: // "tag"
		return isHash
	}
}

// ParseActionUses parses "owner/repo@ref" or "owner/repo/path@ref" into ActionInfo.
// For composite actions like "github/codeql-action/upload-sarif@v2", the repo
// is extracted as "codeql-action" and path as "upload-sarif".
func ParseActionUses(uses string) (*ActionInfo, error) {
	atIdx := strings.LastIndex(uses, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid action format: %s", uses)
	}

	actionPath := uses[:atIdx]
	ref := uses[atIdx+1:]

	// Find first slash for owner
	firstSlash := strings.Index(actionPath, "/")
	if firstSlash == -1 {
		return nil, fmt.Errorf("invalid action path: %s", actionPath)
	}

	owner := actionPath[:firstSlash]
	rest := actionPath[firstSlash+1:]

	// Check for second slash (composite action path)
	secondSlash := strings.Index(rest, "/")
	var repo, path string
	if secondSlash == -1 {
		// Simple case: owner/repo@ref
		repo = rest
	} else {
		// Composite action: owner/repo/path@ref
		repo = rest[:secondSlash]
		path = rest[secondSlash+1:]
	}

	return &ActionInfo{
		Owner: owner,
		Repo:  repo,
		Path:  path,
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
