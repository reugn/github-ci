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
	settings := DefaultFormatSettings()

	if c == nil || c.Linters == nil || c.Linters.Settings == nil {
		return settings
	}

	formatMap, ok := c.Linters.Settings["format"].(map[string]any)
	if !ok {
		return settings
	}

	if v, ok := toInt(formatMap["indent-width"]); ok {
		settings.IndentWidth = v
	}
	if v, ok := toInt(formatMap["max-line-length"]); ok {
		settings.MaxLineLength = v
	}

	return settings
}

// toInt converts a value to int, handling both int and int64 from YAML unmarshaling.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}
