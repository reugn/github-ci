package config

const defaultLinterDefault = "all"

// LinterConfig specifies which linters to enable and their behavior.
// Disabled linters take precedence over enabled linters.
type LinterConfig struct {
	Default  string          `yaml:"default"`            // "all" or "none"
	Enable   []string        `yaml:"enable"`             // Linters to enable
	Disable  []string        `yaml:"disable"`            // Linters to disable
	Settings *LinterSettings `yaml:"settings,omitempty"` // Per-linter settings
}

// LinterSettings contains per-linter configuration.
type LinterSettings struct {
	Format *FormatSettings `yaml:"format,omitempty"`
	Style  *StyleSettings  `yaml:"style,omitempty"`
}

// DefaultLinterConfig returns a minimal LinterConfig with default values.
func DefaultLinterConfig() *LinterConfig {
	return &LinterConfig{
		Default: defaultLinterDefault,
		Enable:  []string{},
		Disable: []string{},
	}
}

// FullDefaultLinterConfig returns a LinterConfig with all settings explicitly set.
func FullDefaultLinterConfig() *LinterConfig {
	return &LinterConfig{
		Default: defaultLinterDefault,
		Enable:  allLinters,
		Disable: []string{},
		Settings: &LinterSettings{
			Format: DefaultFormatSettings(),
			Style:  DefaultStyleSettings(),
		},
	}
}
