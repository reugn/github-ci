package actions

import "testing"

func TestParseActionUses(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOwner   string
		wantRepo    string
		wantPath    string
		wantRef     string
		expectError bool
	}{
		{
			name:      "standard action",
			input:     "actions/checkout@v3",
			wantOwner: "actions",
			wantRepo:  "checkout",
			wantPath:  "",
			wantRef:   "v3",
		},
		{
			name:      "action with commit hash",
			input:     "actions/setup-go@4ab4c1d02e2b3d0af1e9f9c2a3b2c3d4e5f6a7b8c9",
			wantOwner: "actions",
			wantRepo:  "setup-go",
			wantPath:  "",
			wantRef:   "4ab4c1d02e2b3d0af1e9f9c2a3b2c3d4e5f6a7b8c9",
		},
		{
			name:      "action with full version",
			input:     "codecov/codecov-action@v3.1.4",
			wantOwner: "codecov",
			wantRepo:  "codecov-action",
			wantPath:  "",
			wantRef:   "v3.1.4",
		},
		{
			name:      "composite action with path",
			input:     "github/codeql-action/upload-sarif@v2",
			wantOwner: "github",
			wantRepo:  "codeql-action",
			wantPath:  "upload-sarif",
			wantRef:   "v2",
		},
		{
			name:      "composite action with deep path",
			input:     "aws-actions/configure-aws-credentials/assume-role@v4",
			wantOwner: "aws-actions",
			wantRepo:  "configure-aws-credentials",
			wantPath:  "assume-role",
			wantRef:   "v4",
		},
		{
			name:        "missing @",
			input:       "actions/checkout",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "missing repo",
			input:       "actions@v3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseActionUses(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseActionUses(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseActionUses(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result.Owner != tt.wantOwner {
				t.Errorf("ParseActionUses(%q).Owner = %q, want %q", tt.input, result.Owner, tt.wantOwner)
			}
			if result.Repo != tt.wantRepo {
				t.Errorf("ParseActionUses(%q).Repo = %q, want %q", tt.input, result.Repo, tt.wantRepo)
			}
			if result.Path != tt.wantPath {
				t.Errorf("ParseActionUses(%q).Path = %q, want %q", tt.input, result.Path, tt.wantPath)
			}
			if result.Ref != tt.wantRef {
				t.Errorf("ParseActionUses(%q).Ref = %q, want %q", tt.input, result.Ref, tt.wantRef)
			}
		})
	}
}

func TestIsCommitHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid 40-char lowercase", "abcdef1234567890abcdef1234567890abcdef12", true},
		{"valid 40-char uppercase", "ABCDEF1234567890ABCDEF1234567890ABCDEF12", true},
		{"valid 40-char mixed", "AbCdEf1234567890abCDef1234567890abcDEF12", true},
		{"short hash", "abcdef1234", false},
		{"version tag", "v3.1.0", false},
		{"version number", "3", false},
		{"empty string", "", false},
		{"39 chars", "abcdef1234567890abcdef1234567890abcdef1", false},
		{"41 chars", "abcdef1234567890abcdef1234567890abcdef123", false},
		{"40 chars with invalid char", "abcdef1234567890abcdef1234567890abcdefgh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommitHash(tt.input)
			if result != tt.expected {
				t.Errorf("IsCommitHash(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsMajorVersionOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"major with v", "v3", true},
		{"major without v", "3", true},
		{"major.minor with v", "v3.1", false},
		{"major.minor without v", "3.1", false},
		{"full version with v", "v3.1.0", false},
		{"full version without v", "3.1.0", false},
		{"empty string", "", false},
		{"non-numeric", "abc", false},
		{"hash", "abcdef1234567890abcdef1234567890abcdef12", false},
		{"zero", "0", true},
		{"v0", "v0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMajorVersionOnly(tt.input)
			if result != tt.expected {
				t.Errorf("IsMajorVersionOnly(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
