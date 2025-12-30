package config

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/reugn/github-ci/internal/osutil"
	"github.com/reugn/github-ci/internal/version"
	"gopkg.in/yaml.v3"
)

// DefaultConfigFileName is the default name of the configuration file.
const DefaultConfigFileName = ".github-ci.yaml"

// DefaultActionConfig is the default configuration for newly discovered actions.
var DefaultActionConfig = ActionConfig{Version: defaultVersionPattern}

// Config represents the GitHub CI configuration file structure.
type Config struct {
	Run     *RunConfig     `yaml:"run,omitempty"`
	Linters *LinterConfig  `yaml:"linters,omitempty"`
	Upgrade *UpgradeConfig `yaml:"upgrade,omitempty"`
}

// RunConfig specifies general runtime settings.
type RunConfig struct {
	Timeout        string `yaml:"timeout"`          // Duration string (e.g., "2m", "30s")
	IssuesExitCode int    `yaml:"issues-exit-code"` // Exit code when issues are found (default: 1)
}

const (
	// DefaultTimeout is the default timeout for operations.
	DefaultTimeout = 5 * time.Minute
	// DefaultIssuesExitCode is the default exit code when lint issues are found.
	DefaultIssuesExitCode = 1
)

// GetTimeout returns the configured timeout duration.
// Returns DefaultTimeout if not configured or invalid.
func (c *Config) GetTimeout() time.Duration {
	if c.Run == nil || c.Run.Timeout == "" {
		return DefaultTimeout
	}
	d, err := time.ParseDuration(c.Run.Timeout)
	if err != nil {
		return DefaultTimeout
	}
	return d
}

// GetIssuesExitCode returns the configured exit code for when issues are found.
// Returns DefaultIssuesExitCode (1) if not configured or invalid.
// Exit codes must be in range 1-255; values outside this range return the default.
func (c *Config) GetIssuesExitCode() int {
	if c.Run == nil || c.Run.IssuesExitCode <= 0 || c.Run.IssuesExitCode > 255 {
		return DefaultIssuesExitCode
	}
	return c.Run.IssuesExitCode
}

// LoadConfig loads configuration from the specified file.
// Returns defaults if file doesn't exist.
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = DefaultConfigFileName
	}

	if !osutil.FileExists(filename) {
		return NewDefaultConfig(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	cfg.ensureDefaults()
	return &cfg, nil
}

// SaveConfig saves the configuration to the specified file.
func SaveConfig(cfg *Config, filename string) error {
	if filename == "" {
		filename = DefaultConfigFileName
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config file: %w", err)
	}

	return os.WriteFile(filename, data, 0600)
}

// NewDefaultConfig creates a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		Linters: DefaultLinterConfig(),
		Upgrade: DefaultUpgradeConfig(),
	}
}

// NewFullDefaultConfig creates a new Config with all settings explicitly set to defaults.
// This is useful for generating a complete configuration file with all options visible.
func NewFullDefaultConfig() *Config {
	return &Config{
		Linters: FullDefaultLinterConfig(),
		Upgrade: DefaultUpgradeConfig(),
	}
}

// ensureDefaults initializes nil fields with default values.
func (c *Config) ensureDefaults() {
	if c.Linters == nil {
		c.Linters = DefaultLinterConfig()
	}

	if c.Upgrade == nil {
		c.Upgrade = DefaultUpgradeConfig()
	} else {
		c.Upgrade.EnsureDefaults()
	}
}

// GetActionConfig returns the action config, or default if not found.
func (c *Config) GetActionConfig(actionName string) ActionConfig {
	if c.Upgrade != nil {
		if cfg, ok := c.Upgrade.Actions[actionName]; ok {
			return cfg
		}
	}
	return ActionConfig{Version: defaultVersionPattern}
}

// SetActionConfig sets the configuration for an action.
func (c *Config) SetActionConfig(actionName string, cfg ActionConfig) {
	c.ensureDefaults()
	c.Upgrade.Actions[actionName] = cfg
}

// GetVersionFormat returns the version format for upgrades: "tag", "hash", or "major".
// Defaults to "tag" if not specified.
func (c *Config) GetVersionFormat() string {
	if c.Upgrade == nil || c.Upgrade.Version == "" {
		return defaultUpgradeVersion
	}
	return c.Upgrade.Version
}

// IsLinterEnabled checks if a linter is enabled based on configuration.
func (c *Config) IsLinterEnabled(linterName string) bool {
	if c.Linters == nil {
		return true
	}

	if slices.Contains(c.Linters.Disable, linterName) {
		return false
	}

	if c.Linters.Default == defaultLinterDefault {
		return true
	}

	return slices.Contains(c.Linters.Enable, linterName)
}

// NormalizeActionName extracts the action name from a uses string.
func NormalizeActionName(uses string) string {
	if name, _, ok := strings.Cut(uses, "@"); ok {
		return name
	}
	return uses
}

// ShouldUpdate determines if a version update should be applied.
// Patterns:
//   - "": any newer version
//   - "^X.0.0": same major version (X.y.z)
//   - "~X.Y.0": same major.minor version (X.Y.z)
func ShouldUpdate(currentVersion, newVersion, pattern string) bool {
	if version.Compare(newVersion, currentVersion) <= 0 {
		return false
	}

	if pattern == "" {
		return true
	}

	if after, ok := strings.CutPrefix(pattern, "^"); ok {
		return matchesMajorPattern(newVersion, after)
	}

	if after, ok := strings.CutPrefix(pattern, "~"); ok {
		return matchesMinorPattern(newVersion, after)
	}

	return false
}

// matchesMajorPattern checks if version matches ^X.0.0 pattern.
func matchesMajorPattern(newVersion, patternVersion string) bool {
	patternMajor := version.ExtractMajor(patternVersion)
	newMajor := version.ExtractMajor(newVersion)

	// ^1.0.0 is special: allows any version >= 1.0.0
	if patternMajor == 1 {
		return newMajor >= 1
	}

	// Otherwise: same major version only
	return newMajor == patternMajor
}

// matchesMinorPattern checks if version matches ~X.Y.0 pattern.
func matchesMinorPattern(newVersion, patternVersion string) bool {
	patternMajor, patternMinor := version.ExtractMajorMinor(patternVersion)
	newMajor, newMinor := version.ExtractMajorMinor(newVersion)

	return newMajor == patternMajor && newMinor == patternMinor
}
