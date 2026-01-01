package config

import (
	"fmt"
	"slices"
)

const (
	defaultMinNameLength = 3
	defaultMaxNameLength = 50
	defaultMaxRunLines   = 0 // 0 means disabled
)

// Valid naming conventions.
var validNamingConventions = []string{"title", "sentence"}

// StyleSettings contains settings for the style linter.
type StyleSettings struct {
	// MinNameLength is the minimum allowed characters for names (default: 3)
	MinNameLength int `yaml:"min-name-length"`
	// MaxNameLength is the maximum allowed characters for names (default: 50)
	MaxNameLength int `yaml:"max-name-length"`
	// NamingConvention enforces naming style (default: "" - no enforcement):
	//   - "title": Every word must start with uppercase (e.g., "Build And Test", "Setup Go")
	//   - "sentence": Name must start with uppercase (e.g., "Build and test", "Upload to Codecov")
	//   - "": No naming convention enforced
	NamingConvention string `yaml:"naming-convention"`
	// CheckoutFirst warns if actions/checkout is not the first step
	CheckoutFirst bool `yaml:"checkout-first"`
	// RequireStepNames requires all steps to have explicit names
	RequireStepNames bool `yaml:"require-step-names"`
	// MaxRunLines is the maximum allowed lines in a run script (0 = disabled)
	MaxRunLines int `yaml:"max-run-lines"`
}

// Validate checks StyleSettings for invalid values.
func (s *StyleSettings) Validate() error {
	if s == nil {
		return nil
	}
	if s.MinNameLength < 0 {
		return fmt.Errorf("style.min-name-length must be non-negative, got %d", s.MinNameLength)
	}
	if s.MaxNameLength < 0 {
		return fmt.Errorf("style.max-name-length must be non-negative, got %d", s.MaxNameLength)
	}
	if s.MinNameLength > 0 && s.MaxNameLength > 0 && s.MinNameLength > s.MaxNameLength {
		return fmt.Errorf("style.min-name-length (%d) cannot be greater than max-name-length (%d)",
			s.MinNameLength, s.MaxNameLength)
	}
	if s.NamingConvention != "" && !slices.Contains(validNamingConventions, s.NamingConvention) {
		return fmt.Errorf("style.naming-convention must be one of %v, got %q",
			validNamingConventions, s.NamingConvention)
	}
	if s.MaxRunLines < 0 {
		return fmt.Errorf("style.max-run-lines must be non-negative, got %d", s.MaxRunLines)
	}
	return nil
}

// DefaultStyleSettings returns the default style linter settings.
func DefaultStyleSettings() *StyleSettings {
	return &StyleSettings{
		MinNameLength: defaultMinNameLength,
		MaxNameLength: defaultMaxNameLength,
		MaxRunLines:   defaultMaxRunLines,
	}
}

// GetStyleSettings returns the style linter settings from config.
func (c *Config) GetStyleSettings() *StyleSettings {
	if c != nil && c.Linters != nil && c.Linters.Settings != nil && c.Linters.Settings.Style != nil {
		return c.Linters.Settings.Style
	}
	return DefaultStyleSettings()
}
