package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_NonExistent(t *testing.T) {
	// Test loading a non-existent config file returns defaults
	cfg, err := LoadConfig("/nonexistent/path/.github-ci.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.Linters == nil {
		t.Error("cfg.Linters is nil, want non-nil")
	}
	if cfg.Linters.Default != "all" {
		t.Errorf("cfg.Linters.Default = %q, want %q", cfg.Linters.Default, "all")
	}
	if cfg.Upgrade == nil {
		t.Error("cfg.Upgrade is nil, want non-nil")
	}
	if cfg.Upgrade.Version != "tag" {
		t.Errorf("cfg.Upgrade.Version = %q, want %q", cfg.Upgrade.Version, "tag")
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".github-ci.yaml")

	content := `
linters:
  default: none
  enable:
    - permissions
  disable:
    - security
  settings:
    format:
      indent-width: 4
upgrade:
  version: hash
  actions:
    actions/checkout:
      version: ^2.0.0
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Check linters config
	if cfg.Linters.Default != "none" {
		t.Errorf("cfg.Linters.Default = %q, want %q", cfg.Linters.Default, "none")
	}
	if len(cfg.Linters.Enable) != 1 || cfg.Linters.Enable[0] != "permissions" {
		t.Errorf("cfg.Linters.Enable = %v, want [permissions]", cfg.Linters.Enable)
	}
	if len(cfg.Linters.Disable) != 1 || cfg.Linters.Disable[0] != "security" {
		t.Errorf("cfg.Linters.Disable = %v, want [security]", cfg.Linters.Disable)
	}

	// Check upgrade config
	if cfg.Upgrade.Version != "hash" {
		t.Errorf("cfg.Upgrade.Version = %q, want %q", cfg.Upgrade.Version, "hash")
	}
	actionCfg, ok := cfg.Upgrade.Actions["actions/checkout"]
	if !ok {
		t.Fatal("actions/checkout not found in config")
	}
	if actionCfg.Version != "^2.0.0" {
		t.Errorf("actions/checkout version = %q, want %q", actionCfg.Version, "^2.0.0")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".github-ci.yaml")

	content := `invalid: yaml: content: [unclosed`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() expected error for invalid YAML")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".github-ci.yaml")

	cfg := &Config{
		Linters: &LinterConfig{
			Default: "all",
			Enable:  []string{"permissions"},
		},
		Upgrade: &UpgradeConfig{
			Version: "tag",
			Actions: map[string]ActionConfig{
				"actions/checkout": {Version: "^1.0.0"},
			},
		},
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Reload and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.Linters.Default != cfg.Linters.Default {
		t.Errorf("Loaded default = %q, want %q", loaded.Linters.Default, cfg.Linters.Default)
	}
	if loaded.Upgrade.Version != cfg.Upgrade.Version {
		t.Errorf("Loaded version = %q, want %q", loaded.Upgrade.Version, cfg.Upgrade.Version)
	}
}

func TestConfig_GetActionConfig(t *testing.T) {
	cfg := &Config{
		Upgrade: &UpgradeConfig{
			Actions: map[string]ActionConfig{
				"actions/checkout": {Version: "^2.0.0"},
			},
		},
	}

	// Test existing action
	actionCfg := cfg.GetActionConfig("actions/checkout")
	if actionCfg.Version != "^2.0.0" {
		t.Errorf("GetActionConfig() version = %q, want %q", actionCfg.Version, "^2.0.0")
	}

	// Test non-existing action (should return default)
	actionCfg = cfg.GetActionConfig("actions/setup-go")
	if actionCfg.Version != "^1.0.0" {
		t.Errorf("GetActionConfig() default version = %q, want %q", actionCfg.Version, "^1.0.0")
	}
}

func TestConfig_GetActionConfig_NilUpgrade(t *testing.T) {
	cfg := &Config{}

	actionCfg := cfg.GetActionConfig("actions/checkout")
	if actionCfg.Version != "^1.0.0" {
		t.Errorf("GetActionConfig() with nil Upgrade = %q, want %q", actionCfg.Version, "^1.0.0")
	}
}

func TestConfig_SetActionConfig(t *testing.T) {
	cfg := &Config{}

	cfg.SetActionConfig("actions/checkout", ActionConfig{Version: "^3.0.0"})

	if cfg.Upgrade == nil {
		t.Fatal("SetActionConfig() did not initialize Upgrade")
	}
	if cfg.Upgrade.Actions == nil {
		t.Fatal("SetActionConfig() did not initialize Actions map")
	}

	actionCfg := cfg.Upgrade.Actions["actions/checkout"]
	if actionCfg.Version != "^3.0.0" {
		t.Errorf("SetActionConfig() version = %q, want %q", actionCfg.Version, "^3.0.0")
	}
}

func TestConfig_GetVersionFormat(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name:     "nil Upgrade",
			cfg:      &Config{},
			expected: "tag",
		},
		{
			name: "version is tag",
			cfg: &Config{
				Upgrade: &UpgradeConfig{Version: "tag"},
			},
			expected: "tag",
		},
		{
			name: "version is hash",
			cfg: &Config{
				Upgrade: &UpgradeConfig{Version: "hash"},
			},
			expected: "hash",
		},
		{
			name: "version is major",
			cfg: &Config{
				Upgrade: &UpgradeConfig{Version: "major"},
			},
			expected: "major",
		},
		{
			name: "empty version defaults to tag",
			cfg: &Config{
				Upgrade: &UpgradeConfig{Version: ""},
			},
			expected: "tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetVersionFormat()
			if result != tt.expected {
				t.Errorf("GetVersionFormat() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConfig_IsLinterEnabled(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *Config
		linterName string
		expected   bool
	}{
		{
			name:       "nil Linters (default all)",
			cfg:        &Config{},
			linterName: "permissions",
			expected:   true,
		},
		{
			name: "default all, not in disable",
			cfg: &Config{
				Linters: &LinterConfig{Default: "all", Disable: []string{}},
			},
			linterName: "permissions",
			expected:   true,
		},
		{
			name: "default all, in disable",
			cfg: &Config{
				Linters: &LinterConfig{Default: "all", Disable: []string{"permissions"}},
			},
			linterName: "permissions",
			expected:   false,
		},
		{
			name: "default none, in enable",
			cfg: &Config{
				Linters: &LinterConfig{Default: "none", Enable: []string{"permissions"}},
			},
			linterName: "permissions",
			expected:   true,
		},
		{
			name: "default none, not in enable",
			cfg: &Config{
				Linters: &LinterConfig{Default: "none", Enable: []string{"format"}},
			},
			linterName: "permissions",
			expected:   false,
		},
		{
			name: "disable takes precedence over enable",
			cfg: &Config{
				Linters: &LinterConfig{
					Default: "none",
					Enable:  []string{"permissions"},
					Disable: []string{"permissions"},
				},
			},
			linterName: "permissions",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.IsLinterEnabled(tt.linterName)
			if result != tt.expected {
				t.Errorf("IsLinterEnabled(%q) = %v, want %v", tt.linterName, result, tt.expected)
			}
		})
	}
}

func TestConfig_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected time.Duration
	}{
		{
			name:     "nil Run",
			cfg:      &Config{},
			expected: 5 * time.Minute,
		},
		{
			name:     "empty timeout",
			cfg:      &Config{Run: &RunConfig{Timeout: ""}},
			expected: 5 * time.Minute,
		},
		{
			name:     "invalid timeout",
			cfg:      &Config{Run: &RunConfig{Timeout: "invalid"}},
			expected: 5 * time.Minute,
		},
		{
			name:     "30 seconds",
			cfg:      &Config{Run: &RunConfig{Timeout: "30s"}},
			expected: 30 * time.Second,
		},
		{
			name:     "5 minutes",
			cfg:      &Config{Run: &RunConfig{Timeout: "5m"}},
			expected: 5 * time.Minute,
		},
		{
			name:     "1 hour",
			cfg:      &Config{Run: &RunConfig{Timeout: "1h"}},
			expected: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetTimeout()
			if result != tt.expected {
				t.Errorf("GetTimeout() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_GetIssuesExitCode(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected int
	}{
		{
			name:     "nil Run",
			cfg:      &Config{},
			expected: 1,
		},
		{
			name:     "zero exit code defaults to 1",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 0}},
			expected: 1,
		},
		{
			name:     "negative exit code defaults to 1",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: -1}},
			expected: 1,
		},
		{
			name:     "exit code 1",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 1}},
			expected: 1,
		},
		{
			name:     "exit code 2",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 2}},
			expected: 2,
		},
		{
			name:     "exit code 42",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 42}},
			expected: 42,
		},
		{
			name:     "exit code 255 (max valid)",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 255}},
			expected: 255,
		},
		{
			name:     "exit code 256 defaults to 1 (out of range)",
			cfg:      &Config{Run: &RunConfig{IssuesExitCode: 256}},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetIssuesExitCode()
			if result != tt.expected {
				t.Errorf("GetIssuesExitCode() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestNormalizeActionName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard action", "actions/checkout@v3", "actions/checkout"},
		{"with commit hash", "actions/checkout@abc123def456", "actions/checkout"},
		{"without version", "actions/checkout", "actions/checkout"},
		{"empty string", "", ""},
		{"just @", "@", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeActionName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeActionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfig_GetFormatSettings(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *Config
		wantIndentWidth   int
		wantMaxLineLength int
	}{
		{
			name:              "nil config",
			cfg:               nil,
			wantIndentWidth:   2,
			wantMaxLineLength: 120,
		},
		{
			name:              "empty config",
			cfg:               &Config{},
			wantIndentWidth:   2,
			wantMaxLineLength: 120,
		},
		{
			name: "nil settings",
			cfg: &Config{
				Linters: &LinterConfig{},
			},
			wantIndentWidth:   2,
			wantMaxLineLength: 120,
		},
		{
			name: "custom settings",
			cfg: &Config{
				Linters: &LinterConfig{
					Settings: &LinterSettings{
						Format: &FormatSettings{
							IndentWidth:   4,
							MaxLineLength: 100,
						},
					},
				},
			},
			wantIndentWidth:   4,
			wantMaxLineLength: 100,
		},
		{
			name: "partial settings",
			cfg: &Config{
				Linters: &LinterConfig{
					Settings: &LinterSettings{
						Format: &FormatSettings{
							IndentWidth:   4,
							MaxLineLength: 120,
						},
					},
				},
			},
			wantIndentWidth:   4,
			wantMaxLineLength: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.cfg.GetFormatSettings()
			if settings.IndentWidth != tt.wantIndentWidth {
				t.Errorf("IndentWidth = %d, want %d", settings.IndentWidth, tt.wantIndentWidth)
			}
			if settings.MaxLineLength != tt.wantMaxLineLength {
				t.Errorf("MaxLineLength = %d, want %d", settings.MaxLineLength, tt.wantMaxLineLength)
			}
		})
	}
}

func TestConfig_GetStyleSettings(t *testing.T) {
	tests := []struct {
		name                 string
		cfg                  *Config
		wantMinNameLength    int
		wantMaxNameLength    int
		wantNamingConvention string
		wantCheckoutFirst    bool
		wantRequireStepNames bool
	}{
		{
			name:              "nil config",
			cfg:               nil,
			wantMinNameLength: 3,
			wantMaxNameLength: 50,
		},
		{
			name:              "empty config",
			cfg:               &Config{},
			wantMinNameLength: 3,
			wantMaxNameLength: 50,
		},
		{
			name: "full settings",
			cfg: &Config{
				Linters: &LinterConfig{
					Settings: &LinterSettings{
						Style: &StyleSettings{
							MinNameLength:    5,
							MaxNameLength:    100,
							NamingConvention: "title",
							CheckoutFirst:    true,
							RequireStepNames: true,
						},
					},
				},
			},
			wantMinNameLength:    5,
			wantMaxNameLength:    100,
			wantNamingConvention: "title",
			wantCheckoutFirst:    true,
			wantRequireStepNames: true,
		},
		{
			name: "partial settings",
			cfg: &Config{
				Linters: &LinterConfig{
					Settings: &LinterSettings{
						Style: &StyleSettings{
							MinNameLength: 3,
							MaxNameLength: 50,
							CheckoutFirst: true,
						},
					},
				},
			},
			wantMinNameLength: 3,
			wantMaxNameLength: 50,
			wantCheckoutFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.cfg.GetStyleSettings()
			if settings.MinNameLength != tt.wantMinNameLength {
				t.Errorf("MinNameLength = %d, want %d", settings.MinNameLength, tt.wantMinNameLength)
			}
			if settings.MaxNameLength != tt.wantMaxNameLength {
				t.Errorf("MaxNameLength = %d, want %d", settings.MaxNameLength, tt.wantMaxNameLength)
			}
			if settings.NamingConvention != tt.wantNamingConvention {
				t.Errorf("NamingConvention = %q, want %q", settings.NamingConvention, tt.wantNamingConvention)
			}
			if settings.CheckoutFirst != tt.wantCheckoutFirst {
				t.Errorf("CheckoutFirst = %v, want %v", settings.CheckoutFirst, tt.wantCheckoutFirst)
			}
			if settings.RequireStepNames != tt.wantRequireStepNames {
				t.Errorf("RequireStepNames = %v, want %v", settings.RequireStepNames, tt.wantRequireStepNames)
			}
		})
	}
}

func TestShouldUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		newVersion     string
		pattern        string
		expected       bool
	}{
		// New version must be greater
		{"same version", "1.0.0", "1.0.0", "", false},
		{"older version", "2.0.0", "1.0.0", "", false},
		{"newer version, empty pattern", "1.0.0", "2.0.0", "", true},

		// ^1.0.0 pattern
		{"^1.0.0 allows v2", "1.0.0", "v2.0.0", "^1.0.0", true},
		{"^1.0.0 allows v5", "1.0.0", "v5.0.0", "^1.0.0", true},

		// ^2.0.0 pattern (same major)
		{"^2.0.0 allows v2.5", "2.0.0", "v2.5.0", "^2.0.0", true},
		{"^2.0.0 rejects v3", "2.0.0", "v3.0.0", "^2.0.0", false},

		// ~2.5.0 pattern (same major.minor)
		{"~2.5.0 allows v2.5.1", "2.5.0", "v2.5.1", "~2.5.0", true},
		{"~2.5.0 rejects v2.6.0", "2.5.0", "v2.6.0", "~2.5.0", false},

		// Invalid pattern
		{"invalid pattern", "1.0.0", "2.0.0", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUpdate(tt.currentVersion, tt.newVersion, tt.pattern)
			if result != tt.expected {
				t.Errorf("ShouldUpdate(%q, %q, %q) = %v, want %v",
					tt.currentVersion, tt.newVersion, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFullDefaultLinterConfig(t *testing.T) {
	cfg := FullDefaultLinterConfig()

	if cfg.Default != "all" {
		t.Errorf("Default = %q, want %q", cfg.Default, "all")
	}

	// Should have all linters enabled
	expectedLinters := []string{
		LinterVersions, LinterPermissions, LinterFormat,
		LinterSecrets, LinterInjection, LinterStyle,
	}
	if len(cfg.Enable) != len(expectedLinters) {
		t.Errorf("Enable has %d linters, want %d", len(cfg.Enable), len(expectedLinters))
	}

	// Should have settings initialized
	if cfg.Settings == nil {
		t.Error("Settings is nil, want non-nil")
	}
	if cfg.Settings.Format == nil {
		t.Error("Settings.Format is nil, want non-nil")
	}
	if cfg.Settings.Style == nil {
		t.Error("Settings.Style is nil, want non-nil")
	}
}

func TestUpgradeConfig_EnsureDefaults(t *testing.T) {
	tests := []struct {
		name    string
		config  *UpgradeConfig
		wantVer string
	}{
		{
			name:    "nil actions and empty version",
			config:  &UpgradeConfig{},
			wantVer: "tag",
		},
		{
			name:    "nil actions with existing version",
			config:  &UpgradeConfig{Version: "hash"},
			wantVer: "hash",
		},
		{
			name: "existing actions with empty version",
			config: &UpgradeConfig{
				Actions: map[string]ActionConfig{"test": {Version: "^1.0.0"}},
			},
			wantVer: "tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.EnsureDefaults()

			if tt.config.Actions == nil {
				t.Error("Actions is nil after EnsureDefaults")
			}
			if tt.config.Version != tt.wantVer {
				t.Errorf("Version = %q, want %q", tt.config.Version, tt.wantVer)
			}
		})
	}
}
