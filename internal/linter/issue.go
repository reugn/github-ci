package linter

import "fmt"

// Issue represents a linting problem found in a workflow file.
// It contains the file name, line number, linter name, and a descriptive message about the issue.
type Issue struct {
	File    string // Name of the workflow file with the issue
	Line    int    // Line number where the issue was found (0 if not applicable)
	Linter  string // Name of the linter that found this issue
	Message string // Description of the linting issue
}

// newIssue creates an Issue if message is non-empty, otherwise returns nil.
func newIssue(file string, line int, message string) *Issue {
	if message == "" {
		return nil
	}
	return &Issue{
		File:    file,
		Line:    line,
		Message: message,
	}
}

// Key returns a unique identifier for this issue.
func (i *Issue) Key() string {
	return fmt.Sprintf("%s:%d:%s:%s", i.File, i.Line, i.Linter, i.Message)
}

// String implements fmt.Stringer for Issue.
func (i *Issue) String() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s:%d: (%s) %s", i.File, i.Line, i.Linter, i.Message)
	}
	return fmt.Sprintf("%s: (%s) %s", i.File, i.Linter, i.Message)
}
