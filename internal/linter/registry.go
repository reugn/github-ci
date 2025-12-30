package linter

import (
	"context"

	"github.com/reugn/github-ci/internal/config"
)

// linterFactory creates a linter instance with the given context and config.
type linterFactory func(ctx context.Context, cfg *config.Config) Linter

// linterFactories maps linter names to their factory functions.
var linterFactories = map[string]linterFactory{
	config.LinterVersions: func(ctx context.Context, _ *config.Config) Linter {
		return NewVersionsLinter(ctx)
	},
	config.LinterPermissions: func(_ context.Context, _ *config.Config) Linter {
		return NewPermissionsLinter()
	},
	config.LinterFormat: func(_ context.Context, cfg *config.Config) Linter {
		return NewFormatLinter(cfg.GetFormatSettings())
	},
	config.LinterSecrets: func(_ context.Context, _ *config.Config) Linter {
		return NewSecretsLinter()
	},
	config.LinterInjection: func(_ context.Context, _ *config.Config) Linter {
		return NewInjectionLinter()
	},
	config.LinterStyle: func(_ context.Context, cfg *config.Config) Linter {
		return NewStyleLinter(cfg.GetStyleSettings())
	},
}

// lintersWithAutoFix lists linters that support automatic fixing.
var lintersWithAutoFix = map[string]bool{
	config.LinterVersions: true,
	config.LinterFormat:   true,
}

// SupportsAutoFix returns true if the linter supports automatic fixing.
func SupportsAutoFix(linterName string) bool {
	return lintersWithAutoFix[linterName]
}
