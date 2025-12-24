package linter

import (
	"context"
	"fmt"

	"github.com/reugn/github-ci/internal/actions"
	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
)

// Issue represents a linting problem found in a workflow file.
// It contains the file name, line number, linter name, and a descriptive message about the issue.
type Issue struct {
	File    string // Name of the workflow file with the issue
	Line    int    // Line number where the issue was found (0 if not applicable)
	Linter  string // Name of the linter that found this issue
	Message string // Description of the linting issue
}

// Key returns a unique identifier for this issue.
func (i *Issue) Key() string {
	return fmt.Sprintf("%s:%d:%s:%s", i.File, i.Line, i.Linter, i.Message)
}

// String implements fmt.Stringer for Issue.
func (i *Issue) String() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s:%d: (%s) %s", i.File, i.Line, i.Linter, i.Message)
	}
	return fmt.Sprintf("%s: (%s) %s", i.File, i.Linter, i.Message)
}

// WorkflowLinter orchestrates multiple individual linters based on configuration.
type WorkflowLinter struct {
	ctx        context.Context      // Context for timeout/cancellation
	workflows  []*workflow.Workflow // Workflows to analyze
	configFile string               // Path to configuration file
	cfg        *config.Config       // Loaded configuration
	linters    map[string]Linter    // Map of linter name to linter implementation
}

// New creates a new WorkflowLinter instance for the specified workflows directory.
// The directory should contain .yml or .yaml workflow files.
func New(ctx context.Context, workflowsDir string) *WorkflowLinter {
	workflows, err := workflow.LoadWorkflows(workflowsDir)
	if err != nil {
		// Return linter with empty workflows - error will be caught during Lint()
		cfg, _ := config.LoadConfig("")
		return &WorkflowLinter{
			ctx:        ctx,
			workflows:  []*workflow.Workflow{},
			configFile: "",
			cfg:        cfg,
			linters:    createLinters(ctx, cfg),
		}
	}
	return NewWithWorkflows(ctx, workflows, "")
}

// NewWithWorkflows creates a new WorkflowLinter instance with the provided workflows.
// This allows linting specific files or pre-loaded workflows.
// configFile specifies the path to the configuration file (empty string uses default).
func NewWithWorkflows(ctx context.Context, workflows []*workflow.Workflow, configFile string) *WorkflowLinter {
	// LoadConfig returns defaults if file doesn't exist; only errors on parse failures
	cfg, _ := config.LoadConfig(configFile)
	return &WorkflowLinter{
		ctx:        ctx,
		workflows:  workflows,
		configFile: configFile,
		cfg:        cfg,
		linters:    createLinters(ctx, cfg),
	}
}

// createLinters creates a map of enabled linters with their settings from config.
// Only linters that are enabled according to the configuration are created.
// If cfg is nil, all linters are created (default behavior).
func createLinters(ctx context.Context, cfg *config.Config) map[string]Linter {
	linters := make(map[string]Linter)

	// If config is nil, create all linters (default behavior)
	if cfg == nil {
		formatSettings := getFormatSettings(cfg)
		linters[LinterVersions] = NewVersionsLinter(ctx)
		linters[LinterPermissions] = NewPermissionsLinter()
		linters[LinterFormat] = NewFormatLinter(formatSettings)
		linters[LinterSecrets] = NewSecretsLinter()
		linters[LinterInjection] = NewInjectionLinter()
		return linters
	}

	// Get format settings from config, with defaults
	formatSettings := getFormatSettings(cfg)

	// Only create linters that are enabled according to the config
	if cfg.IsLinterEnabled(LinterVersions) {
		linters[LinterVersions] = NewVersionsLinter(ctx)
	}
	if cfg.IsLinterEnabled(LinterPermissions) {
		linters[LinterPermissions] = NewPermissionsLinter()
	}
	if cfg.IsLinterEnabled(LinterFormat) {
		linters[LinterFormat] = NewFormatLinter(formatSettings)
	}
	if cfg.IsLinterEnabled(LinterSecrets) {
		linters[LinterSecrets] = NewSecretsLinter()
	}
	if cfg.IsLinterEnabled(LinterInjection) {
		linters[LinterInjection] = NewInjectionLinter()
	}

	return linters
}

// Lint runs all enabled linters on all workflows and collects their issues.
func (l *WorkflowLinter) Lint() ([]*Issue, error) {
	// Initialize config if not already loaded
	if l.cfg == nil {
		var err error
		l.cfg, err = config.LoadConfig(l.configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		// Recreate linters with the loaded config to get updated settings
		l.linters = createLinters(l.ctx, l.cfg)
	}

	var allIssues []*Issue

	// Iterate over workflows once, running all enabled linters on each
	for _, wf := range l.workflows {
		for name, linter := range l.linters {
			if !l.cfg.IsLinterEnabled(name) {
				continue
			}

			issues, err := linter.LintWorkflow(wf)
			if err != nil {
				return nil, fmt.Errorf("linter %s failed on %s: %w", name, wf.File, err)
			}

			// Set the linter name on each issue
			for _, issue := range issues {
				issue.Linter = name
			}
			allIssues = append(allIssues, issues...)
		}
	}

	return allIssues, nil
}

// Fix runs the Fix method on all enabled linters for all workflows.
func (l *WorkflowLinter) Fix() error {
	// Initialize config if not already loaded
	if l.cfg == nil {
		var err error
		l.cfg, err = config.LoadConfig(l.configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		// Recreate linters with the loaded config to get updated settings
		l.linters = createLinters(l.ctx, l.cfg)
	}

	// Iterate over workflows once, running all enabled linter fixes on each
	for _, wf := range l.workflows {
		for name, linter := range l.linters {
			if !l.cfg.IsLinterEnabled(name) {
				continue
			}

			if err := linter.FixWorkflow(wf); err != nil {
				return fmt.Errorf("linter %s fix failed on %s: %w", name, wf.File, err)
			}
		}
	}

	return nil
}

// GetCacheStats returns cache statistics from the versions linter if it's enabled.
// Returns zero stats if the versions linter is not enabled or not available.
func (l *WorkflowLinter) GetCacheStats() actions.CacheStats {
	if l.linters == nil {
		return actions.CacheStats{}
	}
	if vl, ok := l.linters[LinterVersions].(*VersionsLinter); ok {
		return vl.GetCacheStats()
	}
	return actions.CacheStats{}
}
