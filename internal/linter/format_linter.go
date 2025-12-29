package linter

import (
	"fmt"
	"os"
	"strings"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/stringutil"
	"github.com/reugn/github-ci/internal/workflow"
)

// FormatLinter checks for YAML formatting issues in workflow files.
type FormatLinter struct {
	settings *config.FormatSettings
}

// NewFormatLinter creates a new FormatLinter instance.
func NewFormatLinter(settings *config.FormatSettings) *FormatLinter {
	if settings == nil {
		settings = config.DefaultFormatSettings()
	}
	return &FormatLinter{settings: settings}
}

// LintWorkflow checks a single workflow for formatting issues.
func (l *FormatLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	file := wf.BaseName()
	lines := wf.Lines()
	minIndent := l.findMinIndentation(lines)

	var (
		issues       []*Issue
		prevIndent   int
		prevWasBlank bool
	)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		isBlank := trimmed == ""
		isComment := strings.HasPrefix(trimmed, "#")
		leadingSpaces := stringutil.CountLeadingSpaces(line)

		// Check multiple consecutive blank lines
		if isBlank && prevWasBlank {
			issues = append(issues, newIssue(file, lineNum, "Multiple consecutive blank lines found"))
		}

		// Check trailing whitespace
		if stringutil.HasTrailingWhitespace(line) {
			issues = append(issues, newIssue(file, lineNum, "Line has trailing whitespace"))
		}

		// Check line length
		if issue := l.checkLineLength(line, file, lineNum); issue != nil {
			issues = append(issues, issue)
		}

		// Check indentation (skip blank lines and comments)
		if !isBlank && !isComment {
			if issue := l.checkIndentation(line, file, lineNum, leadingSpaces, minIndent, prevIndent); issue != nil {
				issues = append(issues, issue)
			}
			prevIndent = leadingSpaces
		}

		prevWasBlank = isBlank
	}

	return issues, nil
}

// checkLineLength checks if a line exceeds the configured maximum.
func (l *FormatLinter) checkLineLength(line, file string, lineNum int) *Issue {
	if l.settings == nil || l.settings.MaxLineLength <= 0 {
		return nil
	}
	if len(line) > l.settings.MaxLineLength {
		message := fmt.Sprintf("Line exceeds maximum length of %d characters (found %d)", l.settings.MaxLineLength, len(line))
		return newIssue(file, lineNum, message)
	}
	return nil
}

// checkIndentation validates indentation rules for a line.
func (l *FormatLinter) checkIndentation(line, file string, lineNum, leadingSpaces, minIndent, prevIndent int) *Issue {
	if l.settings == nil || l.settings.IndentWidth <= 0 {
		return nil
	}

	var message string
	switch {
	// Check for tabs
	case strings.HasPrefix(line, "\t"):
		message = fmt.Sprintf("Line uses tabs for indentation, expected %d spaces",
			l.settings.IndentWidth)
	// Check indentation is multiple of indent-width
	case leadingSpaces > 0 && leadingSpaces%l.settings.IndentWidth != 0:
		message = fmt.Sprintf("Line indentation is %d spaces, expected multiple of %d",
			leadingSpaces, l.settings.IndentWidth)
	// Check base indentation level
	case minIndent > 0 && minIndent != l.settings.IndentWidth && leadingSpaces == minIndent:
		message = fmt.Sprintf("Line uses %d spaces for base indentation, expected %d spaces",
			leadingSpaces, l.settings.IndentWidth)
	// Check indentation increase is exactly indent-width
	case leadingSpaces > prevIndent && (leadingSpaces-prevIndent) != l.settings.IndentWidth:
		increase := leadingSpaces - prevIndent
		message = fmt.Sprintf("Line indentation increased by %d spaces, expected increase of %d (should be %d spaces)",
			increase, l.settings.IndentWidth, prevIndent+l.settings.IndentWidth)
	}

	return newIssue(file, lineNum, message)
}

// findMinIndentation finds the minimum non-zero indentation in the file.
func (l *FormatLinter) findMinIndentation(lines []string) int {
	minIndent := -1
	for _, line := range lines {
		if stringutil.IsBlankOrComment(line) {
			continue
		}
		if spaces := stringutil.CountLeadingSpaces(line); spaces > 0 {
			if minIndent == -1 || spaces < minIndent {
				minIndent = spaces
			}
		}
	}
	return minIndent
}

// FixWorkflow automatically fixes formatting issues in a single workflow.
func (l *FormatLinter) FixWorkflow(wf *workflow.Workflow) error {
	lines := wf.Lines()
	fixed := l.fixLines(lines)

	// Remove trailing empty lines
	for len(fixed) > 0 && strings.TrimSpace(fixed[len(fixed)-1]) == "" {
		fixed = fixed[:len(fixed)-1]
	}

	content := strings.Join(fixed, "\n") + "\n"

	// Keep in-memory state in sync with the fixed content
	wf.RawBytes = []byte(content)

	// Write the fixed content to the file
	return os.WriteFile(wf.File, wf.RawBytes, 0600)
}

// fixLines applies formatting fixes to lines.
func (l *FormatLinter) fixLines(lines []string) []string {
	fixed := make([]string, 0, len(lines))
	var prevWasBlank bool
	var prevIndent int

	for _, line := range lines {
		// Trim trailing whitespace
		line = strings.TrimRight(line, " \t")

		// Fix over-indentation
		line = l.fixIndentation(line, prevIndent)

		isBlank := strings.TrimSpace(line) == ""

		// Skip consecutive blank lines (keep only first)
		if isBlank {
			if !prevWasBlank {
				fixed = append(fixed, line)
			}
			prevWasBlank = true
			continue
		}

		fixed = append(fixed, line)
		prevWasBlank = false
		prevIndent = stringutil.CountLeadingSpaces(line)
	}

	return fixed
}

// fixIndentation reduces indentation when it increases by more than indent-width.
func (l *FormatLinter) fixIndentation(line string, prevIndent int) string {
	if l.settings == nil || l.settings.IndentWidth <= 0 {
		return line
	}

	if stringutil.IsBlankOrComment(line) {
		return line
	}

	leadingSpaces := stringutil.CountLeadingSpaces(line)
	if leadingSpaces <= prevIndent {
		return line
	}

	increase := leadingSpaces - prevIndent
	if increase <= l.settings.IndentWidth {
		return line
	}

	// Reduce to correct indentation
	correctIndent := prevIndent + l.settings.IndentWidth
	return strings.Repeat(" ", correctIndent) + strings.TrimLeft(line, " \t")
}
