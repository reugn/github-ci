package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `
name: Test Workflow
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	if wf.File != workflowPath {
		t.Errorf("wf.File = %q, want %q", wf.File, workflowPath)
	}
	if wf.Content.Name != "Test Workflow" {
		t.Errorf("wf.Content.Name = %q, want %q", wf.Content.Name, "Test Workflow")
	}
	if wf.RawBytes == nil {
		t.Error("wf.RawBytes is nil")
	}
}

func TestLoadWorkflow_InvalidPath(t *testing.T) {
	_, err := LoadWorkflow("/nonexistent/path/workflow.yml")
	if err == nil {
		t.Error("LoadWorkflow() expected error for non-existent file")
	}
}

func TestLoadWorkflow_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "invalid.yml")

	content := `invalid: yaml: [unclosed`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	_, err := LoadWorkflow(workflowPath)
	if err == nil {
		t.Error("LoadWorkflow() expected error for invalid YAML")
	}
}

func TestLoadWorkflows(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two workflow files
	workflow1 := `name: Workflow 1
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	workflow2 := `name: Workflow 2
on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`
	if err := os.WriteFile(filepath.Join(tmpDir, "build.yml"), []byte(workflow1), 0600); err != nil {
		t.Fatalf("Failed to write workflow1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte(workflow2), 0600); err != nil {
		t.Fatalf("Failed to write workflow2: %v", err)
	}

	// Create a non-workflow file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0600); err != nil {
		t.Fatalf("Failed to write readme: %v", err)
	}

	workflows, err := LoadWorkflows(tmpDir)
	if err != nil {
		t.Fatalf("LoadWorkflows() error = %v", err)
	}

	if len(workflows) != 2 {
		t.Errorf("LoadWorkflows() returned %d workflows, want 2", len(workflows))
	}
}

func TestLoadWorkflows_InvalidDir(t *testing.T) {
	_, err := LoadWorkflows("/nonexistent/directory")
	if err == nil {
		t.Error("LoadWorkflows() expected error for non-existent directory")
	}
}

func TestWorkflow_FindActions(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `
name: Test Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
      - run: echo "Hello"
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: codecov/codecov-action@v3
`
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	actions, err := wf.FindActions()
	if err != nil {
		t.Fatalf("FindActions() error = %v", err)
	}

	if len(actions) != 3 {
		t.Fatalf("FindActions() returned %d actions, want 3", len(actions))
	}

	// Check the found actions
	expectedActions := []string{
		"actions/checkout@v3",
		"actions/setup-go@v4",
		"codecov/codecov-action@v3",
	}

	for i, expected := range expectedActions {
		if actions[i].Uses != expected {
			t.Errorf("actions[%d].Uses = %q, want %q", i, actions[i].Uses, expected)
		}
		if actions[i].Line == 0 {
			t.Errorf("actions[%d].Line = 0, want non-zero", i)
		}
	}
}

func TestWorkflow_HasPermissions(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "has permissions",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expected: true,
		},
		{
			name: "has complex permissions",
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
			expected: true,
		},
		{
			name: "no permissions",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")
			if err := os.WriteFile(workflowPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to write test workflow: %v", err)
			}

			wf, err := LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			result := wf.HasPermissions()
			if result != tt.expected {
				t.Errorf("HasPermissions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWorkflow_UpdateActionUses(t *testing.T) {
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

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Update the action
	oldUses := "actions/checkout@v3"
	newUses := "actions/checkout@abc123def456789012345678901234567890abcd"
	comment := "v3"

	err = wf.UpdateActionUses(oldUses, newUses, comment)
	if err != nil {
		t.Fatalf("UpdateActionUses() error = %v", err)
	}

	// Reload and verify
	wf2, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() after update error = %v", err)
	}

	foundActions, err := wf2.FindActions()
	if err != nil {
		t.Fatalf("FindActions() error = %v", err)
	}

	if len(foundActions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(foundActions))
	}

	if foundActions[0].Uses != newUses {
		t.Errorf("Updated action uses = %q, want %q", foundActions[0].Uses, newUses)
	}
}

func TestWorkflow_Save(t *testing.T) {
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

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Modify and save
	err = wf.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file still exists and is valid
	wf2, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() after save error = %v", err)
	}

	if wf2.Content.Name != "Test" {
		t.Errorf("After save, name = %q, want %q", wf2.Content.Name, "Test")
	}
}

func TestWorkflow_NormalizeCommentSpacing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		modified bool
	}{
		{
			name: "normalizes extra spaces before version comment",
			input: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123       # v1.0.0
`,
			expected: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123 # v1.0.0
`,
			modified: true,
		},
		{
			name: "normalizes single space before version comment",
			input: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123  # v2.0.0
`,
			expected: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123 # v2.0.0
`,
			modified: true,
		},
		{
			name: "leaves correct spacing unchanged",
			input: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123 # v1.0.0
`,
			expected: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123 # v1.0.0
`,
			modified: false,
		},
		{
			name: "ignores non-version comments",
			input: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123       # some explanation
`,
			expected: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123       # some explanation
`,
			modified: false,
		},
		{
			name: "handles multiple uses lines",
			input: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123       # v1.0.0
      - uses: actions/setup-go@def456    # v2.0.0
`,
			expected: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc123 # v1.0.0
      - uses: actions/setup-go@def456 # v2.0.0
`,
			modified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowPath := filepath.Join(tmpDir, "test.yml")

			if err := os.WriteFile(workflowPath, []byte(tt.input), 0600); err != nil {
				t.Fatalf("Failed to write test workflow: %v", err)
			}

			wf, err := LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			modified := wf.NormalizeCommentSpacing()
			if modified != tt.modified {
				t.Errorf("NormalizeCommentSpacing() modified = %v, want %v", modified, tt.modified)
			}

			if modified {
				if err := wf.Save(); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			// Read back and verify
			data, err := os.ReadFile(workflowPath)
			if err != nil {
				t.Fatalf("Failed to read workflow: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("NormalizeCommentSpacing() result:\n%s\nwant:\n%s", string(data), tt.expected)
			}
		})
	}
}

func TestIsVersionTag(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"v1.0.0", true},
		{"v2", true},
		{"v10.20.30", true},
		{"V1.0.0", true},
		{"1.0.0", true},
		{"1", true},
		{"some explanation", false},
		{"", false},
		{"abc", false},
		{"v", false},
		{"vx.y.z", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isVersionTag(tt.input)
			if result != tt.expected {
				t.Errorf("isVersionTag(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWorkflow_FindJobLine(t *testing.T) {
	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	tests := []struct {
		jobID    string
		expected int
	}{
		{"build", 4},
		{"test", 8},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.jobID, func(t *testing.T) {
			line := wf.FindJobLine(tt.jobID)
			if line != tt.expected {
				t.Errorf("FindJobLine(%q) = %d, want %d", tt.jobID, line, tt.expected)
			}
		})
	}
}

func TestWorkflow_FindStepLine(t *testing.T) {
	content := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
      - run: go build
`
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}

	wf, err := LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	tests := []struct {
		name      string
		jobID     string
		stepIndex int
		expected  int
	}{
		{"first step", "build", 0, 7},
		{"second step", "build", 1, 8},
		{"third step", "build", 2, 10},
		{"out of bounds", "build", 10, 0},
		{"nonexistent job", "nonexistent", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := wf.FindStepLine(tt.jobID, tt.stepIndex)
			if line != tt.expected {
				t.Errorf("FindStepLine(%q, %d) = %d, want %d", tt.jobID, tt.stepIndex, line, tt.expected)
			}
		})
	}
}

func TestWorkflow_ExtractWorkflowEnv(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]bool
	}{
		{
			name: "with env vars",
			content: `name: Test
on: push
env:
  GO_VERSION: "1.21"
  NODE_VERSION: "18"
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expected: map[string]bool{
				"GO_VERSION":   true,
				"NODE_VERSION": true,
			},
		},
		{
			name: "no env vars",
			content: `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expected: map[string]bool{},
		},
		{
			name: "env with comments",
			content: `name: Test
on: push
env:
  # Comment
  MY_VAR: "value"
jobs:
  build:
    runs-on: ubuntu-latest
`,
			expected: map[string]bool{
				"MY_VAR": true,
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

			wf, err := LoadWorkflow(workflowPath)
			if err != nil {
				t.Fatalf("LoadWorkflow() error = %v", err)
			}

			result := wf.ExtractWorkflowEnv()
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractWorkflowEnv() returned %d vars, want %d", len(result), len(tt.expected))
			}
			for key := range tt.expected {
				if !result[key] {
					t.Errorf("ExtractWorkflowEnv() missing key %q", key)
				}
			}
		})
	}
}
