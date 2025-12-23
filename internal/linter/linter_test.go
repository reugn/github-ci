package linter

import (
	"context"
	"testing"

	"github.com/reugn/github-ci/internal/testutil"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestNewWithWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, "")
	if linter == nil {
		t.Fatal("NewWithWorkflows() returned nil")
	}
	if len(linter.workflows) != 1 {
		t.Errorf("linter.workflows length = %d, want 1", len(linter.workflows))
	}
}

func TestWorkflowLinter_Lint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a workflow with multiple issues
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, "")
	issues, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}

	// Should have at least a permissions issue
	if len(issues) == 0 {
		t.Error("Lint() returned 0 issues, expected at least 1 (missing permissions)")
	}

	// Check that issues have linter names set
	for _, issue := range issues {
		if issue.Linter == "" {
			t.Errorf("Issue %q has empty Linter field", issue.Message)
		}
	}
}

func TestWorkflowLinter_LintCleanWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a workflow with no issues (has permissions, uses commit hash)
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11
`)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, "")
	issues, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}

	// Filter out format issues since we can't control exact formatting
	var nonFormatIssues []*Issue
	for _, issue := range issues {
		if issue.Linter != LinterFormat {
			nonFormatIssues = append(nonFormatIssues, issue)
		}
	}

	if len(nonFormatIssues) != 0 {
		t.Errorf("Lint() returned %d non-format issues for clean workflow, want 0", len(nonFormatIssues))
		for _, issue := range nonFormatIssues {
			t.Logf("  Issue: [%s] %s", issue.Linter, issue.Message)
		}
	}
}
