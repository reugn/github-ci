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

func TestActionInfo_Name(t *testing.T) {
	tests := []struct {
		owner string
		repo  string
		path  string
		want  string
	}{
		{"actions", "checkout", "", "actions/checkout"},
		{"github", "codeql-action", "", "github/codeql-action"},
		{"codecov", "codecov-action", "", "codecov/codecov-action"},
		{"github", "codeql-action", "upload-sarif", "github/codeql-action/upload-sarif"},
		{"aws-actions", "configure-aws-credentials", "assume-role",
			"aws-actions/configure-aws-credentials/assume-role"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			info := &ActionInfo{Owner: tt.owner, Repo: tt.repo, Path: tt.path}
			if got := info.Name(); got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestActionInfo_FormatUses(t *testing.T) {
	tests := []struct {
		owner string
		repo  string
		path  string
		ref   string
		want  string
	}{
		{"actions", "checkout", "", "v4", "actions/checkout@v4"},
		{"github", "codeql-action", "upload-sarif", "v2", "github/codeql-action/upload-sarif@v2"},
		{"actions", "setup-go", "", "abc123", "actions/setup-go@abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			info := &ActionInfo{Owner: tt.owner, Repo: tt.repo, Path: tt.path}
			if got := info.FormatUses(tt.ref); got != tt.want {
				t.Errorf("FormatUses(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestActionInfo_IsAtLatest(t *testing.T) {
	latestTag := "v4.2.1"
	latestHash := "abc1234567890123456789012345678901234567"

	tests := []struct {
		name string
		ref  string
		want bool
	}{
		// Hash cases
		{"hash matches latest", latestHash, true},
		{"hash does not match", "def1234567890123456789012345678901234567", false},

		// Tag cases
		{"tag matches latest", "v4.2.1", true},
		{"tag does not match", "v4.2.0", false},
		{"older tag", "v3.0.0", false},

		// Major version cases
		{"major matches latest", "v4", true},
		{"major does not match", "v3", false},
		{"major without v matches", "4", true},
		{"major without v does not match", "3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ActionInfo{
				Owner: "actions",
				Repo:  "checkout",
				Ref:   tt.ref,
			}
			got := info.IsAtLatest(latestTag, latestHash)
			if got != tt.want {
				t.Errorf("IsAtLatest(%q, %q) with ref=%q = %v, want %v",
					latestTag, latestHash, tt.ref, got, tt.want)
			}
		})
	}
}

func TestActionInfo_NeedsFormatChange(t *testing.T) {
	testHash := "abc1234567890123456789012345678901234567"
	tests := []struct {
		name          string
		ref           string
		desiredFormat string
		want          bool
	}{
		// Hash format tests
		{"hash: tag needs change", "v1.2.3", "hash", true},
		{"hash: major needs change", "v1", "hash", true},
		{"hash: hash no change", testHash, "hash", false},

		// Tag format tests
		{"tag: hash needs change", testHash, "tag", true},
		{"tag: tag no change", "v1.2.3", "tag", false},
		{"tag: major no change", "v1", "tag", false},

		// Major format tests
		{"major: hash needs change", testHash, "major", true},
		{"major: full tag needs change", "v1.2.3", "major", true},
		{"major: major no change", "v1", "major", false},
		{"major: two digit major no change", "v12", "major", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ActionInfo{Ref: tt.ref}
			got := info.NeedsFormatChange(tt.desiredFormat)
			if got != tt.want {
				t.Errorf("NeedsFormatChange(%q) with ref=%q = %v, want %v",
					tt.desiredFormat, tt.ref, got, tt.want)
			}
		})
	}
}
