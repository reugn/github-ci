package actions

import "testing"

func TestMatchesVersionConstraint(t *testing.T) {
	tests := []struct {
		name       string
		tagVersion string
		constraint string
		expected   bool
	}{
		// Empty constraint - allow all
		{"empty constraint allows any", "2.0.0", "", true},

		// ^1.0.0 constraint - latest overall
		{"^1.0.0 allows v1", "v1.5.0", "^1.0.0", true},
		{"^1.0.0 allows v2", "v2.0.0", "^1.0.0", true},
		{"^1.0.0 allows v3", "v3.5.0", "^1.0.0", true},

		// ^2.0.0 constraint - same major version
		{"^2.0.0 allows v2.x", "v2.5.0", "^2.0.0", true},
		{"^2.0.0 rejects v3.x", "v3.0.0", "^2.0.0", false},
		{"^2.0.0 rejects v1.x", "v1.9.0", "^2.0.0", false},

		// ~2.5.0 constraint - same major.minor version
		{"~2.5.0 allows v2.5.x", "v2.5.1", "~2.5.0", true},
		{"~2.5.0 rejects v2.6.x", "v2.6.0", "~2.5.0", false},
		{"~2.5.0 rejects v2.4.x", "v2.4.9", "~2.5.0", false},
		{"~2.5.0 rejects v3.5.x", "v3.5.0", "~2.5.0", false},

		// Unknown constraint
		{"unknown constraint", "2.0.0", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesVersionConstraint(tt.tagVersion, tt.constraint)
			if result != tt.expected {
				t.Errorf("matchesVersionConstraint(%q, %q) = %v, want %v",
					tt.tagVersion, tt.constraint, result, tt.expected)
			}
		})
	}
}
