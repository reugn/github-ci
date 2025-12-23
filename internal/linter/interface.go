package linter

import "github.com/reugn/github-ci/internal/workflow"

// Linter is the interface that all individual linters must implement.
// Each linter operates on a single workflow; iteration is handled by the orchestrator.
type Linter interface {
	// LintWorkflow checks a single workflow and returns issues found.
	LintWorkflow(wf *workflow.Workflow) ([]*Issue, error)

	// FixWorkflow attempts to fix issues in a single workflow.
	// For linters that don't support fixing, this should be a no-op (return nil).
	FixWorkflow(wf *workflow.Workflow) error
}
