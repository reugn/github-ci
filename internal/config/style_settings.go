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
	settings := DefaultStyleSettings()

	if c == nil || c.Linters == nil || c.Linters.Settings == nil {
		return settings
	}

	styleMap, ok := c.Linters.Settings["style"].(map[string]any)
	if !ok {
		return settings
	}

	if v, ok := toInt(styleMap["min-name-length"]); ok {
		settings.MinNameLength = v
	}
	if v, ok := toInt(styleMap["max-name-length"]); ok {
		settings.MaxNameLength = v
	}
	if v, ok := styleMap["naming-convention"].(string); ok {
		settings.NamingConvention = v
	}
	if v, ok := styleMap["checkout-first"].(bool); ok {
		settings.CheckoutFirst = v
	}
	if v, ok := styleMap["require-step-names"].(bool); ok {
		settings.RequireStepNames = v
	}

	return settings
}
