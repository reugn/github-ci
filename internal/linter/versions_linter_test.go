package linter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reugn/github-ci/internal/workflow"
)

func TestVersionsLinter_Lint(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectIssues int
	}{
		{
			name: "uses version tag",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: 1,
		},
		{
			name: "uses commit hash",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11
`,
			expectIssues: 0,
		},
		{
			name: "multiple actions with tags",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - uses: codecov/codecov-action@v3.1.4
`,
			expectIssues: 3,
		},
		{
			name: "mixed tags and hashes",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11
      - uses: actions/setup-go@v4
`,
			expectIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")
			if err := os.WriteFile(workflowPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to write test workflow: %v", err)
			}

			wf, err := workflow.LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			linter := NewVersionsLinter(context.Background())
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			if len(issues) != tt.expectIssues {
				t.Errorf("LintWorkflow() returned %d issues, want %d", len(issues), tt.expectIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestVersionsLinter_LintHasLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewVersionsLinter(context.Background())
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Line == 0 {
		t.Error("Issue should have non-zero line number")
	}
}
