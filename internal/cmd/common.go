package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	// Common flags shared across commands
	pathFlag   string
	configFlag string
)

// addCommonFlags adds common flags (path and config) to a command.
func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&pathFlag, "path", "p", ".github/workflows", "Path to workflow directory or file")
	cmd.Flags().StringVarP(&configFlag, "config", "c", ".github-ci.yaml", "Path to configuration file")
}

// createTimeoutContext creates a context with timeout from config.
// Returns the context and a cancel function that must be called to release resources.
func createTimeoutContext(configFile string) (context.Context, context.CancelFunc) {
	cfg, _ := config.LoadConfig(configFile)
	timeout := config.DefaultTimeout
	if cfg != nil {
		timeout = cfg.GetTimeout()
	}
	return context.WithTimeout(context.Background(), timeout)
}

// loadWorkflows loads workflows from the specified path, which can be a directory or a file.
func loadWorkflows(path string) ([]*workflow.Workflow, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if info.IsDir() {
		return workflow.LoadWorkflows(path)
	}

	// Single file
	wf, err := workflow.LoadWorkflow(path)
	if err != nil {
		return nil, err
	}
	return []*workflow.Workflow{wf}, nil
}

// printCacheStats prints GitHub API cache statistics if any calls were made.
func printCacheStats(hits, misses int64) {
	total := hits + misses
	if total > 0 {
		fmt.Printf("\nGitHub API: %d call(s), %d from cache\n", misses, hits)
	}
}
