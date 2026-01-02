package upgrader

import (
	"context"
	"testing"

	"github.com/reugn/github-ci/internal/actions"
	"github.com/reugn/github-ci/internal/testutil"
	"github.com/reugn/github-ci/internal/workflow"
)

const (
	testVersionV4 = "v4.0.0"
	testHash      = "def456789012345678901234567890abcdef1234"
)

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
  format: tag
  actions:
    actions/checkout:
      constraint: ^1.0.0
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
  format: tag
  actions:
    actions/checkout:
      constraint: ^1.0.0
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
  format: tag
  actions:
    actions/checkout:
      constraint: "^1.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns newer version
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, testHash, nil
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
  format: hash
  actions:
    actions/checkout:
      constraint: "^1.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, testHash, nil
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
	expectedUses := "actions/checkout@" + testHash
	if wfActions[0].Uses != expectedUses {
		t.Errorf("Action uses = %q, want %q", wfActions[0].Uses, expectedUses)
	}
}

func TestUpgrader_Upgrade_FormatChangeOnly_HashToTag(t *testing.T) {
	tmpDir := t.TempDir()
	currentHash := "abc123456789012345678901234567890abcdef12"
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@` + currentHash + `
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  format: tag
  actions:
    actions/checkout:
      constraint: "^4.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns the same hash (already at latest) but we want tag format
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, currentHash, nil
		},
		GetTagForCommitFunc: func(_, _, hash string) (string, error) {
			if hash == currentHash {
				return testVersionV4, nil
			}
			return "", nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.Upgrade()
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	// Reload and verify the action was updated to tag format
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

	// Should be updated to use the tag
	expectedUses := "actions/checkout@" + testVersionV4
	if wfActions[0].Uses != expectedUses {
		t.Errorf("Action uses = %q, want %q", wfActions[0].Uses, expectedUses)
	}
}

func TestUpgrader_Upgrade_UseMajor(t *testing.T) {
	tmpDir := t.TempDir()
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3.0.0
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  format: major
  actions:
    actions/checkout:
      constraint: "^1.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, testHash, nil
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	err = upgrader.Upgrade()
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	// Reload and verify the action was updated to major version
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

	// Should be updated to use major version only
	expectedUses := "actions/checkout@v4"
	if wfActions[0].Uses != expectedUses {
		t.Errorf("Action uses = %q, want %q", wfActions[0].Uses, expectedUses)
	}
}

func TestUpgrader_DryRun_UnresolvableHash(t *testing.T) {
	tmpDir := t.TempDir()
	unknownHash := "1234567890123456789012345678901234567890"
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@` + unknownHash + `
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  format: tag
  actions:
    actions/checkout:
      constraint: "^4.0.0"
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	// Mock returns error for GetTagForCommit (unresolvable hash)
	// and returns a newer version
	mockClient := &actions.MockResolver{
		GetLatestVersionFunc: func(_, _, _, _ string) (string, string, error) {
			return testVersionV4, "def456789012345678901234567890abcdef12", nil
		},
		GetTagForCommitFunc: func(_, _, _ string) (string, error) {
			return "", nil // Empty tag = unresolvable
		},
	}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	// Should not error, just print a warning
	err = upgrader.DryRun()
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}
}

func TestUpgrader_DryRun_UnparseableAction(t *testing.T) {
	tmpDir := t.TempDir()
	// Create workflow with invalid action format (missing @)
	workflowContent := `name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: invalid-action-without-ref
`
	workflowPath := testutil.CreateWorkflow(t, tmpDir, "test.yml", workflowContent)

	configContent := `upgrade:
  format: tag
`
	configPath := testutil.CreateConfig(t, tmpDir, configContent)

	wf, err := workflow.LoadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("LoadWorkflow() error = %v", err)
	}

	mockClient := &actions.MockResolver{}

	upgrader := NewWithClient([]*workflow.Workflow{wf}, configPath, mockClient)
	// Should not error, just print a warning for unparseable action
	err = upgrader.DryRun()
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}
}
