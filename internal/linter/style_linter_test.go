package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
)

func TestStyleLinter_MissingWorkflowName(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `on: push
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

	linter := NewStyleLinter(nil)
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Workflow is missing a name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Workflow is missing a name' issue")
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
		t.Errorf("Expected 0 'Step is missing a name' issues with default settings, got %d", count)
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
		t.Errorf("Expected 2 'Step is missing a name' issues with RequireStepNames, got %d", count)
	}
}

func TestStyleLinter_CrypticJobID(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
jobs:
  j1:
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
	if !found {
		t.Error("Expected to find cryptic job ID issue")
	}
}

func TestStyleLinter_CheckoutNotFirst(t *testing.T) {
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

	// Explicitly enable checkout-first check
	settings := (&config.Config{}).GetStyleSettings()
	settings.CheckoutFirst = true
	linter := NewStyleLinter(settings)
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Checkout action should typically be the first step") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Checkout should be first' issue")
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
	settings := (&config.Config{}).GetStyleSettings()
	settings.CheckoutFirst = false
	linter := NewStyleLinter(settings)

	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	for _, issue := range issues {
		if strings.Contains(issue.Message, "Checkout action should typically be the first step") {
			t.Error("Did not expect 'Checkout should be first' issue when CheckoutFirst is disabled")
		}
	}
}

func TestStyleLinter_NameTooShort(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Go
        uses: actions/checkout@v3
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewStyleLinter(&config.StyleSettings{MinNameLength: 3, MaxNameLength: 50})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	count := 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, "too short") {
			count++
		}
	}
	if count < 2 {
		t.Errorf("Expected at least 2 'too short' issues, got %d", count)
	}
}

func TestStyleLinter_NamingConventionTitle(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: build and test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: checkout code
        uses: actions/checkout@v3
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	linter := NewStyleLinter(&config.StyleSettings{
		MinNameLength:    3,
		MaxNameLength:    50,
		NamingConvention: "title",
	})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Title Case") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Title Case naming convention issue")
	}
}

func TestStyleLinter_EnvShadowing(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
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
		if strings.Contains(issue.Message, "shadows workflow-level env var") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find env shadowing issue")
	}
}

func TestStyleLinter_CleanWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Build and Test
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

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for clean workflow, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}
}

func TestStyleLinter_NameNotFirst(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        name: Checkout
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
		if strings.Contains(issue.Message, "name' should come first") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'name should come first' issue")
	}
}

func TestStyleLinter_NamingConventionSentence(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: build and test
on: push
jobs:
  build:
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

	linter := NewStyleLinter(&config.StyleSettings{
		MinNameLength:    3,
		MaxNameLength:    50,
		NamingConvention: "sentence",
		CheckoutFirst:    true,
	})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "should start with uppercase (sentence case)") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find sentence case issue for lowercase first word")
	}
}

func TestStyleLinter_NameTooLong(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	longName := "This is an extremely long workflow name that definitely exceeds the maximum allowed length"
	content := `name: ` + longName + `
on: push
jobs:
  build:
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

	linter := NewStyleLinter(&config.StyleSettings{MaxNameLength: 50, CheckoutFirst: true})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "exceeds maximum length") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'exceeds maximum length' issue")
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

func TestStyleLinter_JobNameValidation(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test Workflow
on: push
jobs:
  build:
    name: build and test things
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

	linter := NewStyleLinter(&config.StyleSettings{
		MinNameLength:    3,
		MaxNameLength:    50,
		NamingConvention: "title",
		CheckoutFirst:    true,
	})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Job name should use Title Case") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Job name should use Title Case' issue")
	}
}

func TestStyleLinter_NoJobs(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
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

	// Should not panic and return no job-related issues
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Job") || strings.Contains(issue.Message, "Step") {
			t.Errorf("Unexpected job/step issue for workflow without jobs: %s", issue.Message)
		}
	}
}

func TestStyleLinter_NoSteps(t *testing.T) {
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

	linter := NewStyleLinter(nil)
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	// Should not panic and return no step-related issues
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Step") {
			t.Errorf("Unexpected step issue for job without steps: %s", issue.Message)
		}
	}
}

func TestStyleLinter_MultipleJobs(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test Workflow
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

	// Clean workflow with multiple jobs should have no issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for clean multi-job workflow, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}
}

func TestStyleLinter_StepNameFirstWithinBounds(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	// Test step where name is on a later line within the step
	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"
        env:
          FOO: bar
        name: Run Test
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
		if strings.Contains(issue.Message, "name' should come first") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'name should come first' issue")
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

func TestStyleLinter_JobNameTooShort(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test Workflow
on: push
jobs:
  build:
    name: Go
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

	linter := NewStyleLinter(&config.StyleSettings{
		MinNameLength: 3,
		MaxNameLength: 50,
		CheckoutFirst: true,
	})
	issues, err := linter.LintWorkflow(wf)
	if err != nil {
		t.Fatalf("LintWorkflow() error = %v", err)
	}

	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Job name") && strings.Contains(issue.Message, "too short") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Job name too short' issue")
	}
}
