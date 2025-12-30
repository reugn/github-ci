package config

const (
	defaultIndentWidth   = 2
	defaultMaxLineLength = 120
)

// FormatSettings contains settings for the format linter.
type FormatSettings struct {
	// IndentWidth is the number of spaces per indentation level (default: 2)
	IndentWidth int `yaml:"indent-width"`
	// MaxLineLength is the maximum allowed line length (default: 120)
	MaxLineLength int `yaml:"max-line-length"`
}

// DefaultFormatSettings returns the default format linter settings.
func DefaultFormatSettings() *FormatSettings {
	return &FormatSettings{
		IndentWidth:   defaultIndentWidth,
		MaxLineLength: defaultMaxLineLength,
	}
}

// GetFormatSettings returns the format linter settings from config.
func (c *Config) GetFormatSettings() *FormatSettings {
	if c != nil && c.Linters != nil && c.Linters.Settings != nil && c.Linters.Settings.Format != nil {
		return c.Linters.Settings.Format
	}
	return DefaultFormatSettings()
}
