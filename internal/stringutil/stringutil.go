package stringutil

import (
	"regexp"
	"strings"
)

// Pre-compiled patterns for cryptic name detection.
var (
	crypticEndsWithNumber = regexp.MustCompile(`^[a-z]+\d+$`)
	crypticShortLowercase = regexp.MustCompile(`^[a-z]{1,4}$`)
)

// IsComment checks if a line is a YAML comment (starts with #).
func IsComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// IsBlankOrComment checks if a line is blank or a YAML comment.
func IsBlankOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

// CountLeadingSpaces returns the number of leading spaces in a line.
func CountLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// HasTrailingWhitespace checks if a line ends with spaces or tabs.
func HasTrailingWhitespace(line string) bool {
	return strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t")
}

// IsCrypticName checks if a name looks cryptic (e.g., job1, j1, abc).
func IsCrypticName(name string) bool {
	// Too short
	if len(name) < 3 {
		return true
	}
	// Ends with a number (job1, step2)
	if crypticEndsWithNumber.MatchString(name) {
		return true
	}
	// All lowercase single word with no meaning indicators
	if crypticShortLowercase.MatchString(name) {
		return true
	}
	return false
}
