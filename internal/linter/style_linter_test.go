package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestStyleLinter_Lint(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		settings       *config.StyleSettings
		expectContains string
	}{
		{
			name: "missing workflow name",
			content: `on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			settings:       nil,
			expectContains: "Workflow is missing a name",
		},
		{
			name: "cryptic job ID",
			content: `name: Test
on: push
jobs:
  j1:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings:       nil,
			expectContains: "cryptic ID",
		},
		{
			name: "checkout not first",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Setup
        run: echo "setup"
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings:       &config.StyleSettings{CheckoutFirst: true},
			expectContains: "Checkout action should typically be the first step",
		},
		{
			name: "name too short",
			content: `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Go
        uses: actions/checkout@v3
`,
			settings:       &config.StyleSettings{MinNameLength: 3, MaxNameLength: 50},
			expectContains: "too short",
		},
		{
			name: "name too long",
			content: `name: This is an extremely long workflow name that definitely exceeds the maximum allowed length
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings:       &config.StyleSettings{MaxNameLength: 50, CheckoutFirst: true},
			expectContains: "exceeds maximum length",
		},
		{
			name: "naming convention title case",
			content: `name: build and test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: checkout code
        uses: actions/checkout@v3
`,
			settings: &config.StyleSettings{
				MinNameLength: 3, MaxNameLength: 50, NamingConvention: "title",
			},
			expectContains: "Title Case",
		},
		{
			name: "naming convention sentence case",
			content: `name: build and test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings: &config.StyleSettings{
				MinNameLength: 3, MaxNameLength: 50, NamingConvention: "sentence", CheckoutFirst: true,
			},
			expectContains: "should start with uppercase (sentence case)",
		},
		{
			name: "env shadowing",
			content: `name: Test
on: push
env:
  NODE_ENV: production
jobs:
  build:
    runs-on: ubuntu-latest
    env:
      NODE_ENV: test
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings:       nil,
			expectContains: "shadows workflow-level env var",
		},
		{
			name: "name not first in step",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        name: Checkout
`,
			settings:       nil,
			expectContains: "name' should come first",
		},
		{
			name: "job name validation - title case",
			content: `name: Test Workflow
on: push
jobs:
  build:
    name: build and test things
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings: &config.StyleSettings{
				MinNameLength: 3, MaxNameLength: 50, NamingConvention: "title", CheckoutFirst: true,
			},
			expectContains: "Job name should use Title Case",
		},
		{
			name: "job name too short",
			content: `name: Test Workflow
on: push
jobs:
  build:
    name: Go
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`,
			settings:       &config.StyleSettings{MinNameLength: 3, MaxNameLength: 50, CheckoutFirst: true},
			expectContains: "Job name",
		},
		{
			name: "step name first within bounds",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"
        env:
          FOO: bar
        name: Run Test
`,
			settings:       nil,
			expectContains: "name' should come first",
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

			linter := NewStyleLinter(tt.settings)
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			found := false
			for _, issue := range issues {
				if strings.Contains(issue.Message, tt.expectContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected issue containing %q", tt.expectContains)
				for _, issue := range issues {
					t.Logf("  Got: %s", issue.Message)
				}
			}
		})
	}
}

func TestStyleLinter_CleanWorkflows(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "single job clean",
			content: `name: Build and Test
on: push
jobs:
  build:
    name: Build Project
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Build
        run: make build
`,
		},
		{
			name: "multiple jobs clean",
			content: `name: Test Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
  deploy:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Deploy
        run: echo "deploying"
`,
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

			linter := NewStyleLinter(nil)
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			if len(issues) != 0 {
				t.Errorf("Expected 0 issues, got %d", len(issues))
				for _, issue := range issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestStyleLinter_MissingStepName(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo "hello"
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// With default settings, no step name issues should be reported
	linter := NewStyleLinter(nil)
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	count := 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Step is missing a name") {
			count++
		}
	}
	if count != 0 {
		t.Errorf("Expected 0 issues with default settings, got %d", count)
	}

	// With RequireStepNames enabled, issues should be reported
	linter = NewStyleLinter(&config.StyleSettings{RequireStepNames: true})
	issues, err = linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	count = 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Step is missing a name") {
			count++
		}
	}
	if count != 2 {
		t.Errorf("Expected 2 issues with RequireStepNames, got %d", count)
	}
}

func TestStyleLinter_CheckoutFirstDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Setup
        run: echo "setup"
      - name: Checkout
        uses: actions/checkout@v3
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Disable checkout-first check
	linter := NewStyleLinter(&config.StyleSettings{CheckoutFirst: false})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	for _, issue := range issues {
		if strings.Contains(issue.Message, "Checkout action should typically be the first step") {
			t.Error("Did not expect issue when CheckoutFirst is disabled")
		}
	}
}

func TestStyleLinter_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		checkNotContain string
	}{
		{
			name: "no jobs",
			content: `name: Test
on: push
`,
			checkNotContain: "Job",
		},
		{
			name: "no steps",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
`,
			checkNotContain: "Step",
		},
		{
			name: "uses step only - no run script check",
			content: `name: Test Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
`,
			checkNotContain: "Run script has",
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

			linter := NewStyleLinter(&config.StyleSettings{MaxRunLines: 1})
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			for _, issue := range issues {
				if strings.Contains(issue.Message, tt.checkNotContain) {
					t.Errorf("Unexpected issue containing %q: %s", tt.checkNotContain, issue.Message)
				}
			}
		})
	}
}

func TestStyleLinter_CrypticJobIDPatterns(t *testing.T) {
	tests := []struct {
		name     string
		jobID    string
		wantFlag bool
	}{
		{"too short", "ab", true},
		{"ends with number", "job1", true},
		{"short lowercase", "test", true},
		{"short lowercase 3 chars", "run", true},
		{"descriptive", "build", false},
		{"with hyphen", "build-test", false},
		{"with uppercase", "Build", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")

			content := `name: Test
on: push
jobs:
  ` + tt.jobID + `:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
`
			if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
				t.Fatalf("Failed to write test workflow: %v", err)
			}

			wf, err := workflow.LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			linter := NewStyleLinter(nil)
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			found := false
			for _, issue := range issues {
				if strings.Contains(issue.Message, "cryptic ID") {
					found = true
					break
				}
			}

			if found != tt.wantFlag {
				t.Errorf("Job ID %q: got flagged=%v, want flagged=%v", tt.jobID, found, tt.wantFlag)
			}
		})
	}
}

func TestStyleLinter_RunScriptLength(t *testing.T) {
	tests := []struct {
		name        string
		maxLines    int
		runScript   string
		wantFlagged bool
	}{
		{"too long", 5, "echo 1\necho 2\necho 3\necho 4\necho 5\necho 6", true},
		{"within limit", 5, "echo 1\necho 2\necho 3", false},
		{"check disabled", 0, "echo 1\necho 2\necho 3\necho 4\necho 5\necho 6\necho 7\necho 8\necho 9\necho 10", false},
		{"single line", 1, "echo hello", false},
		{"exact boundary", 5, "echo 1\necho 2\necho 3\necho 4\necho 5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")

			content := fmt.Sprintf(`name: Test Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Test Step
        run: |
          %s
`, strings.ReplaceAll(tt.runScript, "\n", "\n          "))

			if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
				t.Fatalf("Failed to write test workflow: %v", err)
			}

			wf, err := workflow.LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			linter := NewStyleLinter(&config.StyleSettings{MaxRunLines: tt.maxLines})
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			found := false
			for _, issue := range issues {
				if strings.Contains(issue.Message, "Run script has") {
					found = true
					break
				}
			}

			if found != tt.wantFlagged {
				t.Errorf("got flagged=%v, want flagged=%v", found, tt.wantFlagged)
			}
		})
	}
}

func TestStyleLinter_EmptyName(t *testing.T) {
	linter := NewStyleLinter(&config.StyleSettings{
		NamingConvention: "title",
		CheckoutFirst:    true,
	})

	// Empty name should return nil (no issue)
	issue := linter.checkNamingConvention("", "test.yml", 1, "Step")
	if issue != nil {
		t.Errorf("Expected nil for empty name, got: %s", issue.Message)
	}

	// Whitespace-only name should return nil
	issue = linter.checkNamingConvention("   ", "test.yml", 1, "Step")
	if issue != nil {
		t.Errorf("Expected nil for whitespace name, got: %s", issue.Message)
	}
}
