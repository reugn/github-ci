package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reugn/github-ci/internal/workflow"
)

func TestSecretsLinter_Lint(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectIssues bool
	}{
		{
			name: "no secrets",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo "Hello"
`,
			expectIssues: false,
		},
		{
			name: "aws secret key pattern",
			content: `name: Test
on: push
permissions: read-all
env:
  AWS_SECRET_KEY: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: true,
		},
		{
			name: "github token reference (allowed)",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    env:
      GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    steps:
      - uses: actions/checkout@v3
`,
			expectIssues: false,
		},
		{
			name: "private key pattern",
			content: `name: Test
on: push
permissions: read-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: |
          echo "-----BEGIN RSA PRIVATE KEY-----
          MIIEpAIBAAKCAQEA0abcdefghijklmnopq
          -----END RSA PRIVATE KEY-----"
`,
			expectIssues: true,
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

			linter := NewSecretsLinter()
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			hasIssues := len(issues) > 0
			if hasIssues != tt.expectIssues {
				t.Errorf("LintWorkflow() returned %d issues, expectIssues = %v", len(issues), tt.expectIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestSecretsLinter_Fix(t *testing.T) {
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")

	content := `name: Test
on: push
permissions: read-all
env:
  SECRET: some_hardcoded_secret
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

	linter := NewSecretsLinter()
	// FixWorkflow is a no-op, should not return an error
	err = linter.FixWorkflow(wf)
	if err != nil {
		t.Fatalf("FixWorkflow() error = %v", err)
	}
}
