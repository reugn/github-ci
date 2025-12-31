package linter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/actions"
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

func TestVersionsLinter_WithMockClient(t *testing.T) {
	mock := &actions.MockResolver{
		GetCommitHashFunc: func(_, _, _ string) (string, error) {
			return "abc123def456abc123def456abc123def456abc1", nil
		},
	}

	linter := NewVersionsLinterWithClient(mock)
	if linter == nil {
		t.Fatal("NewVersionsLinterWithClient returned nil")
	}
	if linter.client != mock {
		t.Error("Linter client is not the mock")
	}
}

func TestVersionsLinter_FixWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		mock        *actions.MockResolver
		expectError bool
		checkResult func(t *testing.T, wf *workflow.Workflow)
	}{
		{
			name: "fix version tag to commit hash",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3.5.0
`,
			mock: &actions.MockResolver{
				GetCommitHashFunc: func(_, _, _ string) (string, error) {
					return "b4ffde65f46336ab88eb53be808477a3936bae11", nil
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, wf *workflow.Workflow) {
				content := string(wf.RawBytes)
				if !strings.Contains(content, "b4ffde65f46336ab88eb53be808477a3936bae11") {
					t.Error("Workflow should contain commit hash after fix")
				}
			},
		},
		{
			name: "fix major version resolves to latest minor",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			mock: &actions.MockResolver{
				GetLatestMinorVersionFunc: func(_, _, _ string) (string, string, error) {
					return "v3.5.2", "abc123def456abc123def456abc123def456abc1", nil
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, wf *workflow.Workflow) {
				content := string(wf.RawBytes)
				if !strings.Contains(content, "abc123def456abc123def456abc123def456abc1") {
					t.Error("Workflow should contain resolved commit hash")
				}
			},
		},
		{
			name: "error resolving commit hash",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3.1.0
`,
			mock: &actions.MockResolver{
				GetCommitHashFunc: func(_, _, _ string) (string, error) {
					return "", errors.New("API error")
				},
			},
			expectError: true,
		},
		{
			name: "fallback to GetCommitHash when GetLatestMinorVersion fails",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			mock: &actions.MockResolver{
				GetLatestMinorVersionFunc: func(_, _, _ string) (string, string, error) {
					return "", "", errors.New("no minor versions")
				},
				GetCommitHashFunc: func(_, _, _ string) (string, error) {
					return "fallback123fallback123fallback123fallback1", nil
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, wf *workflow.Workflow) {
				content := string(wf.RawBytes)
				if !strings.Contains(content, "fallback123fallback123fallback123fallback1") {
					t.Error("Workflow should contain fallback commit hash")
				}
			},
		},
		{
			name: "already uses commit hash - no change",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11
`,
			mock:        &actions.MockResolver{},
			expectError: false,
			checkResult: func(t *testing.T, wf *workflow.Workflow) {
				content := string(wf.RawBytes)
				if !strings.Contains(content, "b4ffde65f46336ab88eb53be808477a3936bae11") {
					t.Error("Workflow should still contain original commit hash")
				}
			},
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

			linter := NewVersionsLinterWithClient(tt.mock)
			err = linter.FixWorkflow(wf)

			if tt.expectError {
				if err == nil {
					t.Error("FixWorkflow() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("FixWorkflow() unexpected error = %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, wf)
			}
		})
	}
}

func TestVersionsLinter_GetCacheStats(t *testing.T) {
	mock := &actions.MockResolver{}
	linter := NewVersionsLinterWithClient(mock)

	stats := linter.GetCacheStats()
	// MockResolver returns zero stats
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("GetCacheStats() = {%d, %d}, want {0, 0}", stats.Hits, stats.Misses)
	}
}

func TestVersionsLinter_LintWorkflow_InvalidAction(t *testing.T) {
	// Test that invalid action format is skipped (continue path)
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Local action without @ - ParseActionUses will fail
	content := `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./local-action
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

	// Should only report issue for the valid action, not the local one
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
}

func TestVersionsLinter_FixWorkflow_InvalidAction(t *testing.T) {
	// Test that invalid action format is skipped during fix
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./local-action
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	mock := &actions.MockResolver{}
	linter := NewVersionsLinterWithClient(mock)

	// Should not error, just skip the invalid action
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Errorf("FixWorkflow() unexpected error = %v", err)
	}
}
