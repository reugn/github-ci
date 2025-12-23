package stringutil

import "testing"

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
