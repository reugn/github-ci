package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestFormatLinter_Lint_MultipleBlankLines(t *testing.T) {
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	// Should detect multiple blank lines
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "blank line") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find multiple blank lines issue")
	}
}

func TestFormatLinter_Lint_TrailingWhitespace(t *testing.T) {
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	// Should detect trailing whitespace
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "trailing whitespace") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find trailing whitespace issue")
	}
}

func TestFormatLinter_Lint_LineLength(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Create a line that's definitely too long (>80 chars)
	content := `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo 'This is an extremely long command that exceeds the maximum line length'
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Use a max line length that the long line exceeds
	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 80})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	// Should detect line length issue
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "exceeds") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find line length issue")
	}
}

func TestFormatLinter_Lint_TabIndentation(t *testing.T) {
	// Note: YAML doesn't allow tabs for indentation, so we test by checking
	// that the linter would detect tabs if they existed in a file.
	// We create the raw bytes directly since YAML parsers reject tabs.
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.txt") // Use .txt to avoid YAML validation

	// Create content with tab at start of a line (raw bytes)
	content := "name: Test\non: push\n\tbad-indent: value\n"
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test the checkIndentation function directly with a tab-indented line
	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})

	// checkIndentation(line, file, lineNum, leadingSpaces, minIndent, prevIndent)
	issue := linter.checkIndentation("\tbad-indent: value", "test.yml", 3, 0, 0, 0)
	if issue == nil {
		t.Error("Expected to find tab indentation issue")
	} else if !strings.Contains(issue.Message, "tabs") {
		t.Errorf("Issue message should mention tabs, got: %s", issue.Message)
	}
}

func TestFormatLinter_Fix(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Content with multiple blank lines and trailing whitespace
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Fatalf("FixWorkflow() error = %v", err)
	}

	// Reload and check
	fixed, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Should not have multiple consecutive blank lines
	if strings.Contains(string(fixed), "\n\n\n") {
		t.Error("Fixed content still has multiple blank lines")
	}

	// Should not have trailing whitespace on the first line
	lines := strings.Split(string(fixed), "\n")
	if len(lines) > 0 && strings.HasSuffix(lines[0], " ") {
		t.Error("Fixed content still has trailing whitespace")
	}
}

func TestFormatLinter_CleanFile(t *testing.T) {
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for clean file, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Issue: %s", issue.Message)
		}
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

	// Create linter with nil settings
	linter := NewFormatLinter(nil)

	// Should not panic with nil settings
	_, err = linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}
}

func TestFormatLinter_Lint_OverIndentation(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Content with over-indentation (jumps by 4 instead of 2)
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	// Should detect over-indentation
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "indentation") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find indentation issue for over-indented content")
	}
}

func TestFormatLinter_Fix_OverIndentation(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Content with over-indentation
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Fatalf("FixWorkflow() error = %v", err)
	}

	// Verify in-memory state was updated
	lines := wf.Lines()
	if len(lines) < 4 {
		t.Fatal("Expected at least 4 lines")
	}

	// Check that "build:" line has correct indentation (2 spaces)
	buildLine := lines[3]
	if !strings.HasPrefix(buildLine, "  build:") {
		t.Errorf("Expected 'build:' to have 2-space indent, got: %q", buildLine)
	}
}

func TestFormatLinter_TrailingEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Content with trailing empty lines
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

	linter := NewFormatLinter(&config.FormatSettings{IndentWidth: 2, MaxLineLength: 120})
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Fatalf("FixWorkflow() error = %v", err)
	}

	// Verify trailing empty lines were removed
	fixed, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Should end with exactly one newline
	if strings.HasSuffix(string(fixed), "\n\n") {
		t.Error("Fixed content still has trailing empty lines")
	}
	if !strings.HasSuffix(string(fixed), "\n") {
		t.Error("Fixed content should end with a newline")
	}
}
