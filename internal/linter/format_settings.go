package linter

import "github.com/reugn/github-ci/internal/config"

const (
	defaultIndentWidth   = 2
	defaultMaxLineLength = 120
)

// getFormatSettings extracts format linter settings from config with defaults applied.
func getFormatSettings(cfg *config.Config) *config.FormatSettings {
	settings := &config.FormatSettings{
		IndentWidth:   defaultIndentWidth,
		MaxLineLength: defaultMaxLineLength,
	}

	formatMap := getFormatMap(cfg)
	if formatMap == nil {
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

// getFormatMap extracts the format settings map from config.
func getFormatMap(cfg *config.Config) map[string]any {
	if cfg == nil || cfg.Linters == nil || cfg.Linters.Settings == nil {
		return nil
	}
	formatMap, _ := cfg.Linters.Settings["format"].(map[string]any)
	return formatMap
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
