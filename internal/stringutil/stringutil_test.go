package stringutil

import (
	"testing"
)

func TestIsComment(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"comment with hash", "# this is a comment", true},
		{"comment with leading spaces", "  # indented comment", true},
		{"comment with tabs", "\t# tabbed comment", true},
		{"empty line", "", false},
		{"whitespace only", "   ", false},
		{"yaml key", "name: test", false},
		{"yaml value with hash", "value: test#notcomment", false},
		{"hash in middle", "key: value # comment", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsComment(tt.line); got != tt.want {
				t.Errorf("IsComment(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsBlankOrComment(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"empty line", "", true},
		{"whitespace only", "   ", true},
		{"tabs only", "\t\t", true},
		{"comment", "# comment", true},
		{"indented comment", "  # comment", true},
		{"yaml key", "name: test", false},
		{"yaml list item", "- item", false},
		{"number", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlankOrComment(tt.line); got != tt.want {
				t.Errorf("IsBlankOrComment(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestCountLeadingSpaces(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		{"no spaces", "hello", 0},
		{"two spaces", "  hello", 2},
		{"four spaces", "    hello", 4},
		{"empty string", "", 0},
		{"only spaces", "    ", 4},
		{"tab not counted", "\thello", 0},
		{"mixed tab and spaces", "\t  hello", 0},
		{"spaces then tab", "  \thello", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountLeadingSpaces(tt.line); got != tt.want {
				t.Errorf("CountLeadingSpaces(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestHasTrailingWhitespace(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"no trailing", "hello", false},
		{"trailing space", "hello ", true},
		{"trailing spaces", "hello   ", true},
		{"trailing tab", "hello\t", true},
		{"empty string", "", false},
		{"only spaces", "   ", true},
		{"leading only", "  hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasTrailingWhitespace(tt.line); got != tt.want {
				t.Errorf("HasTrailingWhitespace(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsCrypticName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"too short - 1 char", "a", true},
		{"too short - 2 chars", "ab", true},
		{"ends with number", "job1", true},
		{"ends with number - step", "step2", true},
		{"short lowercase", "abc", true},
		{"short lowercase 4", "abcd", true},
		{"descriptive name", "build", false},
		{"longer name", "build-and-test", false},
		{"camelCase", "buildProject", false},
		{"snake_case", "build_project", false},
		{"uppercase", "BUILD", false},
		{"mixed case short", "Job", false},
		{"number in middle", "step2build", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCrypticName(tt.input); got != tt.want {
				t.Errorf("IsCrypticName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
