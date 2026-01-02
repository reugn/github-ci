package config

import (
	"fmt"
	"slices"
)

const (
	defaultVersionConstraint = "^1.0.0"
	defaultUpgradeFormat     = "tag"
)

// Valid version formats for upgrades.
var validVersionFormats = []string{"tag", "hash", "major"}

// UpgradeConfig specifies settings for the upgrade command.
type UpgradeConfig struct {
	Actions map[string]ActionConfig `yaml:"actions"`
	Format  string                  `yaml:"format"` // "tag", "hash", or "major"
}

// ActionConfig specifies the version constraint for a GitHub Action.
type ActionConfig struct {
	Constraint string `yaml:"constraint"`
}

// Validate checks UpgradeConfig for invalid values.
func (u *UpgradeConfig) Validate() error {
	if u == nil {
		return nil
	}
	if u.Format != "" && !slices.Contains(validVersionFormats, u.Format) {
		return fmt.Errorf("upgrade.format must be one of %v, got %q", validVersionFormats, u.Format)
	}
	return nil
}

// DefaultUpgradeConfig returns an UpgradeConfig with default values.
func DefaultUpgradeConfig() *UpgradeConfig {
	return &UpgradeConfig{
		Actions: make(map[string]ActionConfig),
		Format:  defaultUpgradeFormat,
	}
}

// EnsureDefaults sets default values for any uninitialized fields.
func (u *UpgradeConfig) EnsureDefaults() {
	if u.Actions == nil {
		u.Actions = make(map[string]ActionConfig)
	}
	if u.Format == "" {
		u.Format = defaultUpgradeFormat
	}
}
