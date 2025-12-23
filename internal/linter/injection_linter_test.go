package linter

import (
	"testing"

	"github.com/reugn/github-ci/internal/workflow"
)

func TestInjectionLinter_Lint(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantIssues   int
		wantContexts []string // Expected dangerous contexts found
	}{
		{
			name: "inline run with issue title - vulnerable",
			content: `name: Test
on: issues
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Issue: ${{ github.event.issue.title }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.issue.title"},
		},
		{
			name: "inline run with issue body - vulnerable",
			content: `name: Test
on: issues
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.body }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.issue.body"},
		},
		{
			name: "multiline run with PR title - vulnerable",
			content: `name: Test
on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: |
          echo "PR: ${{ github.event.pull_request.title }}"
          echo "Done"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.pull_request.title"},
		},
		{
			name: "head_ref in run command - vulnerable",
			content: `name: Test
on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Branch: ${{ github.head_ref }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.head_ref"},
		},
		{
			name: "comment body in run - vulnerable",
			content: `name: Test
on: issue_comment
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: process_comment "${{ github.event.comment.body }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.comment.body"},
		},
		{
			name: "commit message in run - vulnerable",
			content: `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.head_commit.message }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.head_commit.message"},
		},
		{
			name: "safe - using environment variable",
			content: `name: Test
on: issues
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Issue: $TITLE"
        env:
          TITLE: ${{ github.event.issue.title }}
`,
			wantIssues: 0,
		},
		{
			name: "safe - using secrets reference",
			content: `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Token: ${{ secrets.GITHUB_TOKEN }}"
`,
			wantIssues: 0,
		},
		{
			name: "safe - using safe context",
			content: `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "SHA: ${{ github.sha }}"
`,
			wantIssues: 0,
		},
		{
			name: "safe - github.ref is safe",
			content: `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Ref: ${{ github.ref }}"
`,
			wantIssues: 0,
		},
		{
			name: "multiple vulnerabilities",
			content: `name: Test
on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.pull_request.title }}"
      - run: |
          echo "${{ github.event.pull_request.body }}"
          echo "${{ github.head_ref }}"
`,
			wantIssues: 3,
		},
		{
			name: "PR head ref - vulnerable",
			content: `name: Test
on: pull_request
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: git checkout ${{ github.event.pull_request.head.ref }}
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.pull_request.head.ref"},
		},
		{
			name: "review body - vulnerable",
			content: `name: Test
on: pull_request_review
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.review.body }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.review.body"},
		},
		{
			name: "discussion title - vulnerable",
			content: `name: Test
on: discussion
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.discussion.title }}"
`,
			wantIssues:   1,
			wantContexts: []string{"github.event.discussion.title"},
		},
		{
			name: "no run commands",
			content: `name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
`,
			wantIssues: 0,
		},
		{
			name: "expression in env block only - safe",
			content: `name: Test
on: issues
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Safe step
        run: echo "$ISSUE_TITLE"
        env:
          ISSUE_TITLE: ${{ github.event.issue.title }}
`,
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := &workflow.Workflow{
				File:     "test.yml",
				RawBytes: []byte(tt.content),
			}

			linter := NewInjectionLinter()
			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				t.Fatalf("LintWorkflow() error = %v", err)
			}

			if len(issues) != tt.wantIssues {
				t.Errorf("LintWorkflow() got %d issues, want %d", len(issues), tt.wantIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}

			// Verify specific contexts were found
			for _, ctx := range tt.wantContexts {
				found := false
				for _, issue := range issues {
					if containsContext(issue.Message, ctx) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find context %q in issues", ctx)
				}
			}
		})
	}
}

func containsContext(message, context string) bool {
	return len(message) > 0 && len(context) > 0 &&
		(message != "" && context != "" &&
			(len(message) >= len(context) &&
				(message == context || containsSubstring(message, context))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInjectionLinter_Fix(t *testing.T) {
	linter := NewInjectionLinter()
	wf := &workflow.Workflow{
		File:     "test.yml",
		RawBytes: []byte("name: Test\n"),
	}

	// FixWorkflow should be a no-op
	err := linter.FixWorkflow(wf)
	if err != nil {
		t.Errorf("FixWorkflow() error = %v, want nil", err)
	}
}

func TestExtractRunContent(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"run: echo hello", "echo hello"},
		{"run : echo hello", "echo hello"},
		{"run: |", ""},
		{"run: |-", ""},
		{"run: >", ""},
		{"run: >-", ""},
		{"  run: some command", "some command"},
		{"name: test", ""},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := extractRunContent(tt.line)
			if got != tt.want {
				t.Errorf("extractRunContent(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsRunCommand(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"run: echo hello", true},
		{"run : echo", true},
		{"run:", true},
		{"- run: test", true},
		{"- run : test", true},
		{"name: run", false},
		{"running: something", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isRunCommand(tt.line)
			if got != tt.want {
				t.Errorf("isRunCommand(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
