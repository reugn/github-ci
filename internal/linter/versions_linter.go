package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/reugn/github-ci/internal/actions"
	"github.com/reugn/github-ci/internal/workflow"
)

// VersionsLinter checks for actions using version tags instead of commit hashes.
type VersionsLinter struct {
	client actions.Resolver
}

// NewVersionsLinter creates a new VersionsLinter instance with the provided context.
func NewVersionsLinter(ctx context.Context) *VersionsLinter {
	return &VersionsLinter{
		client: actions.NewClientWithContext(ctx),
	}
}

// NewVersionsLinterWithClient creates a new VersionsLinter instance with a custom client.
// This is useful for testing with a mock client.
func NewVersionsLinterWithClient(client actions.Resolver) *VersionsLinter {
	return &VersionsLinter{
		client: client,
	}
}

// LintWorkflow checks a single workflow for actions using version tags instead of commit hashes.
func (l *VersionsLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	workflowActions, err := wf.FindActions()
	if err != nil {
		return nil, fmt.Errorf("failed to find actions: %w", err)
	}

	var issues []*Issue
	for _, action := range workflowActions {
		actionInfo, err := actions.ParseActionUses(action.Uses)
		if err != nil {
			continue
		}

		if !actions.IsCommitHash(actionInfo.Ref) {
			message := fmt.Sprintf("Action %s uses version tag '%s' instead of commit hash",
				action.Uses, actionInfo.Ref)
			issues = append(issues, newIssue(wf.BaseName(), action.Line, message))
		}
	}

	return issues, nil
}

// FixWorkflow fixes issues in a single workflow by replacing version tags with commit hashes.
func (l *VersionsLinter) FixWorkflow(wf *workflow.Workflow) error {
	workflowActions, err := wf.FindActions()
	if err != nil {
		return fmt.Errorf("failed to find actions: %w", err)
	}

	for _, action := range workflowActions {
		actionInfo, err := actions.ParseActionUses(action.Uses)
		if err != nil {
			continue
		}

		if !actions.IsCommitHash(actionInfo.Ref) {
			if err := l.resolveAndUpdateAction(wf, action, actionInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

// resolveAndUpdateAction resolves an action reference to a commit hash and updates the workflow.
// If the ref is a major version only (e.g., "v3"), it finds the latest minor version in that series.
func (l *VersionsLinter) resolveAndUpdateAction(wf *workflow.Workflow, action *workflow.Action,
	info *actions.ActionInfo) error {
	ref := strings.TrimPrefix(info.Ref, "tags/")
	tag, hash, err := l.resolveVersion(info.Owner, info.Repo, ref)
	if err != nil {
		return fmt.Errorf("failed to get commit hash for %s: %w", action.Uses, err)
	}

	newUses := fmt.Sprintf("%s/%s@%s", info.Owner, info.Repo, hash)
	if err := wf.UpdateActionUses(action.Uses, newUses, tag); err != nil {
		return fmt.Errorf("failed to update action in %s: %w", wf.File, err)
	}

	return nil
}

// resolveVersion resolves a version ref to its tag name and commit hash.
func (l *VersionsLinter) resolveVersion(owner, repo, ref string) (tag, hash string, err error) {
	// For major versions (e.g., "v3"), find the latest minor version
	if actions.IsMajorVersionOnly(ref) {
		tag, hash, err = l.client.GetLatestMinorVersion(owner, repo, ref)
		if err == nil {
			return tag, hash, nil
		}
		// Fall through to regular resolution on error
	}

	hash, err = l.client.GetCommitHash(owner, repo, ref)
	return ref, hash, err
}

// GetCacheStats returns cache statistics for GitHub API calls.
func (l *VersionsLinter) GetCacheStats() actions.CacheStats {
	return l.client.GetCacheStats()
}
