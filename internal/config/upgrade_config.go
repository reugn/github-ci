package config

const (
	defaultVersionPattern = "^1.0.0"
	defaultUpgradeVersion = "tag"
)

// UpgradeConfig specifies settings for the upgrade command.
type UpgradeConfig struct {
	Actions map[string]ActionConfig `yaml:"actions"`
	Version string                  `yaml:"version"` // "tag", "hash", or "major"
}

// ActionConfig specifies the version update pattern for a GitHub Action.
type ActionConfig struct {
	Version string `yaml:"version"`
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
