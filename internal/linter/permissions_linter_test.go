package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reugn/github-ci/internal/workflow"
)

func TestPermissionsLinter_Lint(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectIssues bool
	}{
		{
			name: "missing permissions",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: true,
		},
		{
			name: "has permissions string",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: false,
		},
		{
			name: "has permissions object",
			content: `name: Test
on: push
permissions:
  contents: read
  packages: write
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: false,
		},
		{
			name: "empty permissions",
			content: `name: Test
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: false,
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

			linter := NewPermissionsLinter()
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			hasIssues := len(issues) > 0
			if hasIssues != tt.expectIssues {
				t.Errorf("LintWorkflow() returned %d issues, expectIssues = %v", len(issues), tt.expectIssues)
			}
		})
	}
}

func TestPermissionsLinter_Fix(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
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

	linter := NewPermissionsLinter()
	// FixWorkflow is a no-op, should not return an error
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Fatalf("FixWorkflow() error = %v", err)
	}
}
