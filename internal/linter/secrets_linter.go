package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/reugn/github-ci/internal/stringutil"
	"github.com/reugn/github-ci/internal/workflow"
)

// secretPattern defines a regex pattern for detecting potential hardcoded secrets.
type secretPattern struct {
	re   *regexp.Regexp
	name string
}

// secretPatterns contains compiled patterns for common secret types.
var secretPatterns = []secretPattern{
	{regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "API key"},
	{regexp.MustCompile(`(?i)(token|access[_-]?token)\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "Token"},
	{regexp.MustCompile(`(?i)(secret|secret[_-]?key)\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "Secret"},
	{regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*['"]?.{8,}['"]?`), "Password"},
	{regexp.MustCompile(`(?i)aws[_-]?(access[_-]?key[_-]?id|secret[_-]?access[_-]?key)` +
		`\s*[:=]\s*['"]?[A-Z0-9]{20,}['"]?`), "AWS key"},
	{regexp.MustCompile(`(?i)ghp_[a-zA-Z0-9]{36}`), "GitHub personal access token"},
	{regexp.MustCompile(`(?i)github[_-]?token\s*[:=]\s*['"]?[a-zA-Z0-9]{20,}['"]?`), "GitHub token"},
	{regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`), "Private key"},
	{regexp.MustCompile(`(?i)(key|credential|auth)\s*[:=]\s*['"]?[a-zA-Z0-9+/=]{32,}['"]?`), "Potential credential"},
}

// SecretsLinter checks for hardcoded secrets in workflow files.
type SecretsLinter struct {
	noOpFixer
}

// NewSecretsLinter creates a new SecretsLinter instance.
func NewSecretsLinter() *SecretsLinter {
	return &SecretsLinter{}
}

// LintWorkflow checks a single workflow for hardcoded secrets.
func (l *SecretsLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	var issues []*Issue
	file := wf.BaseName()
	lines := wf.Lines()

	for i, line := range lines {
		if issue := l.checkLine(file, i+1, line); issue != nil {
			issues = append(issues, issue)
		}
	}

	return issues, nil
}

// checkLine checks a single line for potential secrets.
func (l *SecretsLinter) checkLine(file string, lineNum int, line string) *Issue {
	if stringutil.IsBlankOrComment(line) {
		return nil
	}

	// Skip lines using GitHub Actions expressions for secrets/env
	if isSecretReference(line) {
		return nil
	}

	for _, p := range secretPatterns {
		if p.re.MatchString(line) {
			message := fmt.Sprintf("Potential hardcoded %s detected", p.name)
			return newIssue(file, lineNum, message)
		}
	}

	return nil
}

// isSecretReference returns true if the line references secrets via GitHub Actions expressions.
func isSecretReference(line string) bool {
	if !strings.Contains(line, "${{") {
		return false
	}
	return strings.Contains(line, "secrets.") || strings.Contains(line, "env.")
}
