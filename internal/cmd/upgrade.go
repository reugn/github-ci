package cmd

import (
	"fmt"

	"github.com/reugn/github-ci/internal/upgrader"
	"github.com/spf13/cobra"
)

var dryRunFlag bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [path]",
	Short: "Upgrade GitHub Actions in workflows",
	Long: `Check for newer versions of actions in all workflows and update them.
Creates .github-ci.yaml config file if it doesn't exist.

The path can be a directory (e.g., .github/workflows) or a specific workflow file.
If no path is provided, defaults to .github/workflows.`,
	RunE:         runUpgrade,
	SilenceUsage: true,
}

func init() {
	addCommonFlags(upgradeCmd)
	upgradeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be updated without making changes")
}

func runUpgrade(_ *cobra.Command, args []string) error {
	workflowsPath := pathFlag
	if len(args) > 0 {
		workflowsPath = args[0]
	}

	workflows, err := loadWorkflows(workflowsPath)
	if err != nil {
		return fmt.Errorf("failed to load workflows: %w", err)
	}

	ctx, cancel := createTimeoutContext(configFlag)
	defer cancel()

	upgrader := upgrader.NewWithWorkflows(ctx, workflows, configFlag)

	if dryRunFlag {
		if err := upgrader.DryRun(); err != nil {
			return fmt.Errorf("failed to check for upgrades: %w", err)
		}
	} else {
		if err := upgrader.Upgrade(); err != nil {
			return fmt.Errorf("failed to upgrade workflows: %w", err)
		}
		fmt.Println("✓ Upgrade completed successfully")
	}

	stats := upgrader.GetCacheStats()
	printCacheStats(stats.Hits, stats.Misses)

	return nil
}
