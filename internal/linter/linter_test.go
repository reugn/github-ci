package linter

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/testutil"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	_ = testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`)

	linter := New(context.Background(), tmpDir)
	if linter == nil {
		t.Fatal("New() returned nil")
	}
	if len(linter.workflows) != 1 {
		t.Errorf("linter.workflows length = %d, want 1", len(linter.workflows))
	}
}

func TestNew_InvalidDirectory(t *testing.T) {
	linter := New(context.Background(), "/nonexistent/path")
	if linter == nil {
		t.Fatal("New() returned nil for invalid directory")
	}
	// Should have empty workflows
	if len(linter.workflows) != 0 {
		t.Errorf("linter.workflows length = %d, want 0", len(linter.workflows))
	}
}

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

	// Create a workflow with no issues (has permissions, uses commit hash, has step names)
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11
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
		if issue.Linter != config.LinterFormat {
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

func TestWorkflowLinter_Fix(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a workflow with fixable issues (trailing whitespace)
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `name: Test   
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

	// Fix the workflow
	if err := linter.Fix(); err != nil {
		t.Fatalf("Fix() error = %v", err)
	}

	// Verify the file was fixed
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Should not have trailing whitespace on first line
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasSuffix(lines[0], " ") {
		t.Error("Fix() did not remove trailing whitespace")
	}
}

func TestWorkflowLinter_Fix_UpdatesInMemoryState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a workflow with multiple blank lines (fixable)
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `name: Test
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

	// Lint before fix
	issuesBefore, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() before fix error = %v", err)
	}

	// Fix
	if err := linter.Fix(); err != nil {
		t.Fatalf("Fix() error = %v", err)
	}

	// Lint after fix - should have fewer format issues
	issuesAfter, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() after fix error = %v", err)
	}

	// Count format issues
	countFormat := func(issues []*Issue) int {
		count := 0
		for _, issue := range issues {
			if issue.Linter == config.LinterFormat {
				count++
			}
		}
		return count
	}

	formatBefore := countFormat(issuesBefore)
	formatAfter := countFormat(issuesAfter)

	if formatAfter >= formatBefore && formatBefore > 0 {
		t.Errorf("Fix() did not reduce format issues: before=%d, after=%d", formatBefore, formatAfter)
	}
}

func TestWorkflowLinter_GetCacheStats(t *testing.T) {
	tests := []struct {
		name   string
		linter *WorkflowLinter
	}{
		{
			name: "with workflows",
			linter: func() *WorkflowLinter {
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
				wf, _ := workflow.LoadWorkflow(workflowPath)
				return NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, "")
			}(),
		},
		{
			name:   "nil linters",
			linter: &WorkflowLinter{linters: nil},
		},
		{
			name: "versions linter disabled",
			linter: &WorkflowLinter{
				linters: map[string]Linter{
					config.LinterPermissions: NewPermissionsLinter(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := tt.linter.GetCacheStats()
			if stats.Hits != 0 || stats.Misses != 0 {
				t.Errorf("GetCacheStats() = {Hits: %d, Misses: %d}, want {0, 0}", stats.Hits, stats.Misses)
			}
		})
	}
}

func TestSupportsAutoFix(t *testing.T) {
	tests := []struct {
		linter string
		want   bool
	}{
		{config.LinterVersions, true},
		{config.LinterFormat, true},
		{config.LinterPermissions, false},
		{config.LinterSecrets, false},
		{config.LinterInjection, false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.linter, func(t *testing.T) {
			if got := SupportsAutoFix(tt.linter); got != tt.want {
				t.Errorf("SupportsAutoFix(%q) = %v, want %v", tt.linter, got, tt.want)
			}
		})
	}
}

func TestWorkflowLinter_Lint_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config that disables all linters except permissions
	configPath := testutil.CreateWorkflow(t, tmpDir, ".github-ci.yaml", `
linters:
  default: none
  enable:
    - permissions
`)

	// Create a workflow with multiple potential issues
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

	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, configPath)
	issues, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}

	// Should only have permissions issues since other linters are disabled
	for _, issue := range issues {
		if issue.Linter != config.LinterPermissions {
			t.Errorf("Got issue from disabled linter %s: %s", issue.Linter, issue.Message)
		}
	}
}

func TestWorkflowLinter_Lint_EmptyWorkflows(t *testing.T) {
	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{}, "")
	issues, err := linter.Lint()
	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Lint() with empty workflows returned %d issues, want 0", len(issues))
	}
}

func TestWorkflowLinter_Fix_EmptyWorkflows(t *testing.T) {
	linter := NewWithWorkflows(context.Background(), []*workflow.Workflow{}, "")
	err := linter.Fix()
	if err != nil {
		t.Fatalf("Fix() with empty workflows error = %v", err)
	}
}

func TestWorkflowLinter_Lint_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
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

	// Create linter with nil config to test lazy loading
	linter := &WorkflowLinter{
		ctx:        context.Background(),
		workflows:  []*workflow.Workflow{wf},
		configFile: "",  // Will load default config
		cfg:        nil, // Force lazy loading in Lint()
		linters:    nil,
	}

	_, err = linter.Lint()
	if err != nil {
		t.Fatalf("Lint() with nil config error = %v", err)
	}
}

func TestWorkflowLinter_Fix_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `name: Test   
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

	// Create linter with nil config to test lazy loading
	linter := &WorkflowLinter{
		ctx:        context.Background(),
		workflows:  []*workflow.Workflow{wf},
		configFile: "",  // Will load default config
		cfg:        nil, // Force lazy loading in Fix()
		linters:    nil,
	}

	err = linter.Fix()
	if err != nil {
		t.Fatalf("Fix() with nil config error = %v", err)
	}

	// Verify trailing whitespace was fixed
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasSuffix(lines[0], " ") {
		t.Error("Fix() did not remove trailing whitespace with lazy config loading")
	}
}

func TestCreateLinters_NilConfig(t *testing.T) {
	linters := createLinters(context.Background(), nil)

	// With nil config, all linters should be created
	if len(linters) == 0 {
		t.Error("createLinters(nil) returned empty map")
	}

	// Check some expected linters exist
	if _, ok := linters[config.LinterVersions]; !ok {
		t.Error("createLinters(nil) missing versions linter")
	}
	if _, ok := linters[config.LinterPermissions]; !ok {
		t.Error("createLinters(nil) missing permissions linter")
	}
}
