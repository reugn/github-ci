package config

import (
	"fmt"
	"slices"
)

const (
	defaultVersionPattern = "^1.0.0"
	defaultUpgradeVersion = "tag"
)

// Valid version formats for upgrades.
var validVersionFormats = []string{"tag", "hash", "major"}

// UpgradeConfig specifies settings for the upgrade command.
type UpgradeConfig struct {
	Actions map[string]ActionConfig `yaml:"actions"`
	Version string                  `yaml:"version"` // "tag", "hash", or "major"
}

// ActionConfig specifies the version update pattern for a GitHub Action.
type ActionConfig struct {
	Version string `yaml:"version"`
}

// Validate checks UpgradeConfig for invalid values.
func (u *UpgradeConfig) Validate() error {
	if u == nil {
		return nil
	}
	if u.Version != "" && !slices.Contains(validVersionFormats, u.Version) {
		return fmt.Errorf("upgrade.version must be one of %v, got %q", validVersionFormats, u.Version)
	}
	return nil
}

// DefaultUpgradeConfig returns an UpgradeConfig with default values.
func DefaultUpgradeConfig() *UpgradeConfig {
	return &UpgradeConfig{
		Actions: make(map[string]ActionConfig),
		Version: defaultUpgradeVersion,
	}
}

// EnsureDefaults sets default values for any uninitialized fields.
func (u *UpgradeConfig) EnsureDefaults() {
	if u.Actions == nil {
		u.Actions = make(map[string]ActionConfig)
	}
	if u.Version == "" {
		u.Version = defaultUpgradeVersion
	}
}
