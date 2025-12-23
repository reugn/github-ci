package version

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with v prefix", "v1.2.3", "1.2.3"},
		{"without v prefix", "1.2.3", "1.2.3"},
		{"with whitespace", "  v1.2.3  ", "1.2.3"},
		{"major only with v", "v3", "3"},
		{"major only without v", "3", "3"},
		{"empty string", "", ""},
		{"only v", "v", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractMajor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"full version with v", "v1.2.3", 1},
		{"full version without v", "2.5.0", 2},
		{"major only with v", "v3", 3},
		{"major only without v", "4", 4},
		{"major.minor with v", "v5.1", 5},
		{"zero version", "0.0.1", 0},
		{"empty string", "", 0},
		{"invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMajor(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractMajor(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractMajorMinor(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedMajor int
		expectedMinor int
	}{
		{"full version with v", "v1.2.3", 1, 2},
		{"full version without v", "2.5.0", 2, 5},
		{"major.minor with v", "v5.1", 5, 1},
		{"major.minor without v", "3.7", 3, 7},
		{"major only with v", "v3", 3, 0},
		{"major only without v", "4", 4, 0},
		{"empty string", "", 0, 0},
		{"invalid", "abc.def", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor := ExtractMajorMinor(tt.input)
			if major != tt.expectedMajor || minor != tt.expectedMinor {
				t.Errorf("ExtractMajorMinor(%q) = (%d, %d), want (%d, %d)",
					tt.input, major, minor, tt.expectedMajor, tt.expectedMinor)
			}
		})
	}
}

func TestToMajorTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full version with v", "v5.2.0", "v5"},
		{"full version without v", "5.2.0", "v5"},
		{"major.minor with v", "v3.1", "v3"},
		{"major only with v", "v7", "v7"},
		{"major only without v", "4", "v4"},
		{"zero version", "0.1.2", "v0"},
		{"empty string", "", "v0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMajorTag(tt.input)
			if result != tt.expected {
				t.Errorf("ToMajorTag(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"equal with v prefix", "v1.2.3", "1.2.3", 0},
		{"v1 less than v2", "1.2.3", "1.2.4", -1},
		{"v1 greater than v2", "1.2.4", "1.2.3", 1},
		{"major difference", "2.0.0", "1.9.9", 1},
		{"minor difference", "1.3.0", "1.2.9", 1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"different lengths 2", "1.0.1", "1.0", 1},
		{"major only", "v3", "v2", 1},
		{"major only equal", "v3", "3", 0},
		{"empty vs version", "", "1.0.0", -1},
		{"version vs empty", "1.0.0", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
