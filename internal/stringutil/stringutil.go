package stringutil

import "strings"

// IsComment checks if a line is a YAML comment (starts with #).
func IsComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// IsBlankOrComment checks if a line is blank or a YAML comment.
func IsBlankOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}
