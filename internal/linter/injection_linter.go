package linter

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/reugn/github-ci/internal/stringutil"
	"github.com/reugn/github-ci/internal/workflow"
)

// dangerousContexts lists GitHub context expressions that can be attacker-controlled
// and are dangerous when used directly in run: commands.
// These should be passed through environment variables instead.
var dangerousContexts = []string{
	// Issue-related contexts
	"github.event.issue.title",
	"github.event.issue.body",
	// Pull request contexts
	"github.event.pull_request.title",
	"github.event.pull_request.body",
	"github.event.pull_request.head.ref",
	"github.event.pull_request.head.label",
	"github.event.pull_request.head.repo.default_branch",
	// Comment contexts
	"github.event.comment.body",
	"github.event.review.body",
	"github.event.review_comment.body",
	// Discussion contexts
	"github.event.discussion.title",
	"github.event.discussion.body",
	// Commit contexts (can be controlled via commit messages)
	"github.event.head_commit.message",
	"github.event.head_commit.author.name",
	"github.event.head_commit.author.email",
	"github.event.commits[*].message",
	"github.event.commits[*].author.name",
	"github.event.commits[*].author.email",
	// Branch name (attacker-controlled in forks)
	"github.head_ref",
	// Pages build contexts
	"github.event.pages[*].source.path",
	// Author/sender contexts
	"github.event.*.author.name",
	"github.event.*.author.email",
}

// dangerousPatterns contains compiled regex patterns for detecting dangerous contexts.
// Patterns support both exact matches and wildcards.
var (
	dangerousPatterns []*regexp.Regexp
	patternsOnce      sync.Once
)

// stepConfigKeys lists YAML keys that indicate step configuration rather than run content.
var stepConfigKeys = []string{
	"env:", "name:", "with:", "if:", "id:", "uses:",
	"continue-on-error:", "timeout-minutes:", "working-directory:",
	"shell:",
}

// initPatterns compiles dangerous context patterns once.
func initPatterns() {
	patternsOnce.Do(func() {
		dangerousPatterns = make([]*regexp.Regexp, 0, len(dangerousContexts))
		for _, ctx := range dangerousContexts {
			// Convert wildcard patterns to regex
			pattern := regexp.QuoteMeta(ctx)
			pattern = strings.ReplaceAll(pattern, `\*`, `[^}]+`)
			re := regexp.MustCompile(`\$\{\{\s*` + pattern + `\s*\}\}`)
			dangerousPatterns = append(dangerousPatterns, re)
		}
	})
}

// InjectionLinter checks for shell injection vulnerabilities in workflow files.
// It detects dangerous use of GitHub context expressions in run: commands
// that could allow attackers to inject arbitrary commands.
type InjectionLinter struct{}

// NewInjectionLinter creates a new InjectionLinter instance.
func NewInjectionLinter() *InjectionLinter {
	initPatterns()
	return &InjectionLinter{}
}

// lineContext tracks the parsing state while scanning workflow lines.
type lineContext struct {
	inRunBlock     bool
	inEnvBlock     bool
	runBlockIndent int
	envBlockIndent int
}

// LintWorkflow checks a single workflow for injection vulnerabilities.
func (l *InjectionLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	var issues []*Issue
	file := wf.BaseName()
	lines := wf.Lines()
	ctx := &lineContext{}

	for i, line := range lines {
		if issue := l.processLine(file, i+1, line, ctx); issue != nil {
			issues = append(issues, issue)
		}
	}

	return issues, nil
}

// processLine processes a single line and returns an issue if found.
func (l *InjectionLinter) processLine(file string, lineNum int, line string, ctx *lineContext) *Issue {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines and comments
	if stringutil.IsBlankOrComment(line) {
		return nil
	}

	currentIndent := stringutil.CountLeadingSpaces(line)

	// Update context based on current line
	l.updateContext(trimmed, currentIndent, ctx)

	// Detect start of run: block
	if isRunCommand(trimmed) {
		ctx.inRunBlock = true
		ctx.inEnvBlock = false
		ctx.runBlockIndent = currentIndent

		// Check inline run command (run: echo "...")
		if runContent := extractRunContent(trimmed); runContent != "" {
			return l.checkForInjection(file, lineNum, runContent)
		}
		return nil
	}

	// Check content inside run block (but not in env block)
	if ctx.inRunBlock && !ctx.inEnvBlock && !isStepConfigKey(trimmed) {
		return l.checkForInjection(file, lineNum, line)
	}

	return nil
}

// updateContext updates the parsing context based on the current line.
func (l *InjectionLinter) updateContext(trimmed string, currentIndent int, ctx *lineContext) {
	// Check if we've exited the env block
	if ctx.inEnvBlock && currentIndent <= ctx.envBlockIndent {
		ctx.inEnvBlock = false
	}

	// Check if we've entered an env block
	if strings.HasPrefix(trimmed, "env:") {
		ctx.inEnvBlock = true
		ctx.envBlockIndent = currentIndent
		return
	}

	// Check if we've exited the run block (new step)
	if ctx.inRunBlock && currentIndent <= ctx.runBlockIndent && isStepBoundary(trimmed) {
		ctx.inRunBlock = false
		ctx.inEnvBlock = false
	}
}

// isStepBoundary checks if a line starts a new step or is at step boundary.
func isStepBoundary(trimmed string) bool {
	return strings.HasPrefix(trimmed, "- name:") ||
		strings.HasPrefix(trimmed, "- uses:") ||
		strings.HasPrefix(trimmed, "- run:") ||
		strings.HasPrefix(trimmed, "- if:") ||
		strings.HasPrefix(trimmed, "- id:")
}

// isStepConfigKey checks if a line is a step configuration key (not run content).
func isStepConfigKey(trimmed string) bool {
	for _, key := range stepConfigKeys {
		if strings.HasPrefix(trimmed, key) {
			return true
		}
	}
	return false
}

// isRunCommand checks if a line contains a run: command.
func isRunCommand(trimmed string) bool {
	// Handle both direct "run:" and list item "- run:"
	if strings.HasPrefix(trimmed, "run:") || strings.HasPrefix(trimmed, "run :") {
		return true
	}
	if strings.HasPrefix(trimmed, "- run:") || strings.HasPrefix(trimmed, "- run :") {
		return true
	}
	return false
}

// extractRunContent extracts the content after "run:" on the same line.
func extractRunContent(line string) string {
	// Try "run:" first
	idx := strings.Index(line, "run:")
	offset := 4
	if idx == -1 {
		// Try "run :" with space
		idx = strings.Index(line, "run :")
		offset = 5
	}
	if idx == -1 {
		return ""
	}

	content := strings.TrimSpace(line[idx+offset:])
	if isBlockScalarIndicator(content) {
		return ""
	}
	return content
}

// isBlockScalarIndicator returns true if content is a YAML block scalar indicator.
func isBlockScalarIndicator(content string) bool {
	switch content {
	case "|", "|-", "|+", ">", ">-", ">+":
		return true
	}
	return false
}

// checkForInjection checks if a line contains dangerous GitHub context expressions.
func (l *InjectionLinter) checkForInjection(file string, lineNum int, line string) *Issue {
	// First check if line contains any expression
	if !strings.Contains(line, "${{") {
		return nil
	}

	// Check against each dangerous pattern
	for _, pattern := range dangerousPatterns {
		if matches := pattern.FindStringSubmatch(line); len(matches) > 0 {
			expr := matches[0]
			message := fmt.Sprintf(
				"Potential shell injection: %s in run command. Use an environment variable instead",
				expr,
			)
			return newIssue(file, lineNum, message)
		}
	}

	return nil
}

// FixWorkflow is a no-op as injection fixes require manual refactoring to use environment variables.
func (l *InjectionLinter) FixWorkflow(_ *workflow.Workflow) error {
	return nil
}
