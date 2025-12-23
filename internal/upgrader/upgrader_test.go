package upgrader

import (
	"context"
	"testing"

	"github.com/reugn/github-ci/internal/actions"
	"github.com/reugn/github-ci/internal/testutil"
	"github.com/reugn/github-ci/internal/workflow"
)

const testVersionV4 = "v4.0.0"

func TestNewWithWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
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

	upgrader := NewWithWorkflows(context.Background(), []*workflow.Workflow{wf}, "")
	if upgrader == nil {
		t.Fatal("NewWithWorkflows() returned nil")
	}
	if len(upgrader.workflows) != 1 {
		t.Errorf("upgrader.workflows length = %d, want 1", len(upgrader.workflows))
	}
}

func TestNewWithClient(t *testing.T) {
	tmpDir := t.TempDir()
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

	mockClient := &actions.MockResolver{}
	upgrader := NewWithClient([]*workflow.Workflow{wf}, "", mockClient)
	if upgrader == nil {
		t.Fatal("NewWithClient() returned nil")
	}
	if upgrader.client != mockClient {
		t.Error("NewWithClient() did not set the provided client")
	}
}

func TestUpgrader_DryRun_NoUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`)

	configPath := testutil.CreateConfig(t, tmpDir, `
upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ^1.0.0
`)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns same version (no update needed)
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return "v3", "abc123", nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.DryRun()
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}
}

func TestUpgrader_DryRun_WithUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`)

	configPath := testutil.CreateConfig(t, tmpDir, `
upgrade:
  version: tag
  actions:
    actions/checkout:
      version: ^1.0.0
`)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns newer version
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, "def456", nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.DryRun()
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}
}

func TestUpgrader_Upgrade(t *testing.T) {
	tmpDir := t.TempDir()
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  version: tag
  actions:
    actions/checkout:
      version: "^1.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns newer version
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, "def456789012345678901234567890abcdef1234", nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.Upgrade()
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	// Reload and verify the action was updated
	wf2, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() after upgrade error = %v", err)
	}

	wfActions, err := wf2.FindActions()
	if err != nil {
		t.Fatalf("FindActions() error = %v", err)
	}

	if len(wfActions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(wfActions))
	}

	// Should be updated to the new version tag
	if wfActions[0].Uses != "actions/checkout@"+testVersionV4 {
		t.Errorf("Action uses = %q, want %q", wfActions[0].Uses, "actions/checkout@"+testVersionV4)
	}
}

func TestUpgrader_Upgrade_UseHash(t *testing.T) {
	tmpDir := t.TempDir()
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  version: hash
  actions:
    actions/checkout:
      version: "^1.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	newHash := "def456789012345678901234567890abcdef1234"
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, newHash, nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.Upgrade()
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	// Reload and verify the action was updated with hash
	wf2, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() after upgrade error = %v", err)
	}

	wfActions, err := wf2.FindActions()
	if err != nil {
		t.Fatalf("FindActions() error = %v", err)
	}

	if len(wfActions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(wfActions))
	}

	// Should be updated to use the commit hash
	expectedUses := "actions/checkout@" + newHash
	if wfActions[0].Uses != expectedUses {
		t.Errorf("Action uses = %q, want %q", wfActions[0].Uses, expectedUses)
	}
}
