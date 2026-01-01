package config

import (
	"fmt"
	"slices"
)

const defaultLinterDefault = "all"

// LinterConfig specifies which linters to enable and their behavior.
// Disabled linters take precedence over enabled linters.
type LinterConfig struct {
	Default  string          `yaml:"default"`            // "all" or "none"
	Enable   []string        `yaml:"enable"`             // Linters to enable
	Disable  []string        `yaml:"disable"`            // Linters to disable
	Settings *LinterSettings `yaml:"settings,omitempty"` // Per-linter settings
}

// Validate checks LinterConfig for invalid values.
func (l *LinterConfig) Validate() error {
	if l == nil {
		return nil
	}
	if l.Default != "" && l.Default != "all" && l.Default != "none" {
		return fmt.Errorf("linters.default must be \"all\" or \"none\", got %q", l.Default)
	}
	for _, name := range l.Enable {
		if !slices.Contains(allLinters, name) {
			return fmt.Errorf("unknown linter %q in linters.enable", name)
		}
	}
	for _, name := range l.Disable {
		if !slices.Contains(allLinters, name) {
			return fmt.Errorf("unknown linter %q in linters.disable", name)
		}
	}
	if err := l.Settings.Validate(); err != nil {
		return err
	}
	return nil
}

// LinterSettings contains per-linter configuration.
type LinterSettings struct {
	Format *FormatSettings `yaml:"format,omitempty"`
	Style  *StyleSettings  `yaml:"style,omitempty"`
}

// Validate checks LinterSettings for invalid values.
func (s *LinterSettings) Validate() error {
	if s == nil {
		return nil
	}
	if err := s.Format.Validate(); err != nil {
		return err
	}
	if err := s.Style.Validate(); err != nil {
		return err
	}
	return nil
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
