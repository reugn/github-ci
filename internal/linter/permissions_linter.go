package linter

import (
	"github.com/reugn/github-ci/internal/workflow"
)

// PermissionsLinter checks for missing permissions configuration in workflows.
type PermissionsLinter struct {
	noOpFixer
}

// NewPermissionsLinter creates a new PermissionsLinter instance.
func NewPermissionsLinter() *PermissionsLinter {
	return &PermissionsLinter{}
}

// LintWorkflow checks a single workflow for missing permissions configuration.
func (l *PermissionsLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	if !wf.HasPermissions() {
		issue := newIssue(wf.BaseName(), 0, "Workflow is missing permissions configuration")
		return []*Issue{issue}, nil
	}
	return nil, nil
}
