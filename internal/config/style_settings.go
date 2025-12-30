package config

const (
	defaultMinNameLength = 3
	defaultMaxNameLength = 50
)

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
}

// DefaultStyleSettings returns the default style linter settings.
func DefaultStyleSettings() *StyleSettings {
	return &StyleSettings{
		MinNameLength: defaultMinNameLength,
		MaxNameLength: defaultMaxNameLength,
	}
}

// GetStyleSettings returns the style linter settings from config.
func (c *Config) GetStyleSettings() *StyleSettings {
	if c != nil && c.Linters != nil && c.Linters.Settings != nil && c.Linters.Settings.Style != nil {
		return c.Linters.Settings.Style
	}
	return DefaultStyleSettings()
}
