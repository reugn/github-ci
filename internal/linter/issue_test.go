package linter

import "testing"

func Test_newIssue(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		line    int
		message string
		wantNil bool
	}{
		{
			name:    "empty message returns nil",
			file:    "test.yml",
			line:    10,
			message: "",
			wantNil: true,
		},
		{
			name:    "non-empty message returns issue",
			file:    "test.yml",
			line:    10,
			message: "some issue",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := newIssue(tt.file, tt.line, tt.message)
			if tt.wantNil {
				if issue != nil {
					t.Errorf("newIssue() = %v, want nil", issue)
				}
			} else {
				if issue == nil {
					t.Error("newIssue() = nil, want non-nil")
				} else {
					if issue.File != tt.file {
						t.Errorf("issue.File = %q, want %q", issue.File, tt.file)
					}
					if issue.Line != tt.line {
						t.Errorf("issue.Line = %d, want %d", issue.Line, tt.line)
					}
					if issue.Message != tt.message {
						t.Errorf("issue.Message = %q, want %q", issue.Message, tt.message)
					}
				}
			}
		})
	}
}

func TestIssue_Key(t *testing.T) {
	issue := &Issue{
		File:    "test.yml",
		Line:    10,
		Linter:  "style",
		Message: "some issue",
	}
	want := "test.yml:10:style:some issue"
	if got := issue.Key(); got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}
}

func TestIssue_String(t *testing.T) {
	tests := []struct {
		name  string
		issue *Issue
		want  string
	}{
		{
			name: "with line number",
			issue: &Issue{
				File:    "test.yml",
				Line:    10,
				Linter:  "style",
				Message: "some issue",
			},
			want: "test.yml:10: (style) some issue",
		},
		{
			name: "without line number",
			issue: &Issue{
				File:    "test.yml",
				Line:    0,
				Linter:  "style",
				Message: "some issue",
			},
			want: "test.yml: (style) some issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.issue.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
