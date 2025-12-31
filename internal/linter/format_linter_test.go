package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestFormatLinter_Lint(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		settings       *config.FormatSettings
		expectContains string
	}{
		{
			name: "multiple blank lines",
			content: `name: Test
on: push


permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			settings:       &config.FormatSettings{IndentWidth: 2, MaxLineLength: 120},
			expectContains: "blank line",
		},
		{
			name: "trailing whitespace",
			content: `name: Test   
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			settings:       &config.FormatSettings{IndentWidth: 2, MaxLineLength: 120},
			expectContains: "trailing whitespace",
		},
		{
			name: "line too long",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo 'This is an extremely long command that exceeds the maximum line length'
`,
			settings:       &config.FormatSettings{IndentWidth: 2, MaxLineLength: 80},
			expectContains: "exceeds",
		},
		{
			name: "over-indentation",
			content: `name: Test
on: push
jobs:
    build:
        runs-on: ubuntu-latest
`,
			settings:       &config.FormatSettings{IndentWidth: 2, MaxLineLength: 120},
			expectContains: "indentation",
		},
		{
			name: "clean file - no issues",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			settings:       &config.FormatSettings{IndentWidth: 2, MaxLineLength: 120},
			expectContains: "", // No issue expected
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

			linter := NewFormatLinter(tt.settings)
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			if tt.expectContains == "" {
				// Expect no issues
				if len(issues) != 0 {
					t.Errorf("Expected 0 issues, got %d", len(issues))
					for _, issue := range issues {
						t.Logf("  Issue: %s", issue.Message)
					}
				}
				return
			}

			// Expect at least one issue containing the expected string
			found := false
			for _, issue := range issues {
				if strings.Contains(issue.Message, tt.expectContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected issue containing %q, got issues: %v", tt.expectContains, issues)
			}
		})
	}
}

func TestFormatLinter_Lint_TabIndentation(t *testing.T) {
	// Test the checkIndentation function directly with a tab-indented line
	// since YAML parsers reject tabs
	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})

	issue := linter.checkIndentation("\tbad-indent: value", "test.yml", 3, 0, 0, 0)
	if issue == nil {
		t.Error("Expected to find tab indentation issue")
	} else if !strings.Contains(issue.Message, "tabs") {
		t.Errorf("Issue message should mention tabs, got: %s", issue.Message)
	}
}

func TestFormatLinter_Fix(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		checkFunc func(t *testing.T, fixed []byte)
	}{
		{
			name: "remove trailing whitespace and multiple blank lines",
			content: `name: Test   
on: push


permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			checkFunc: func(t *testing.T, fixed []byte) {
				if strings.Contains(string(fixed), "\n\n\n") {
					t.Error("Fixed content still has multiple blank lines")
				}
				lines := strings.Split(string(fixed), "\n")
				if len(lines) > 0 && strings.HasSuffix(lines[0], " ") {
					t.Error("Fixed content still has trailing whitespace")
				}
			},
		},
		{
			name: "fix over-indentation",
			content: `name: Test
on: push
jobs:
    build:
        runs-on: ubuntu-latest
`,
			checkFunc: func(t *testing.T, fixed []byte) {
				lines := strings.Split(string(fixed), "\n")
				if len(lines) < 4 {
					t.Fatal("Expected at least 4 lines")
				}
				if !strings.HasPrefix(lines[3], "  build:") {
					t.Errorf("Expected 'build:' to have 2-space indent, got: %q", lines[3])
				}
			},
		},
		{
			name: "remove trailing empty lines",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest


`,
			checkFunc: func(t *testing.T, fixed []byte) {
				if strings.HasSuffix(string(fixed), "\n\n") {
					t.Error("Fixed content still has trailing empty lines")
				}
				if !strings.HasSuffix(string(fixed), "\n") {
					t.Error("Fixed content should end with a newline")
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

			linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
			err = linter.FixWorkflow(wf)
			if err != nil {
				t.Fatalf("FixWorkflow() error = %v", err)
			}

			fixed, err := os.ReadFile(workflowPath)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			tt.checkFunc(t, fixed)
		})
	}
}

func TestFormatLinter_NilSettings(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Create linter with nil settings - should not panic
	linter := NewFormatLinter(nil)
	_, err = linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}
}
