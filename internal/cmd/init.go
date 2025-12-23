package cmd

import (
	"fmt"
	"os"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/osutil"
	"github.com/reugn/github-ci/internal/workflow"
	"github.com/spf13/cobra"
)

var updateFlag bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a new .github-ci.yaml configuration file with default settings.

If the configuration file already exists:
  - Without --update: fails with an error
  - With --update: adds any new actions found in workflows to the config

The command scans workflows to discover actions and adds them to the config
with default version patterns.`,
	RunE:         runInit,
	SilenceUsage: true,
}

func init() {
	addCommonFlags(initCmd)
	initCmd.Flags().BoolVarP(&updateFlag, "update", "u", false,
		"Update existing config with new actions from workflows")
}

func runInit(_ *cobra.Command, _ []string) error {
	configExists := osutil.FileExists(configFlag)

	// Check if config exists and we're not updating
	if configExists && !updateFlag {
		return fmt.Errorf("config file %s already exists (use --update to add new actions)", configFlag)
	}

	// Load existing config or create new one
	var cfg *config.Config
	var err error
	if configExists {
		cfg, err = config.LoadConfig(configFlag)
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
	} else {
		cfg = config.NewDefaultConfig()
	}

	// Load workflows and discover actions
	workflows, err := loadWorkflows(pathFlag)
	if err != nil {
		// If no workflows found, just create config with defaults
		if !configExists {
			if err := config.SaveConfig(cfg, configFlag); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Created %s\n", configFlag)
			return nil
		}
		return fmt.Errorf("failed to load workflows: %w", err)
	}

	// Discover actions from workflows
	newActions := discoverActions(cfg, workflows)

	// Save config
	if err := config.SaveConfig(cfg, configFlag); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print result
	if configExists {
		if len(newActions) > 0 {
			fmt.Printf("✓ Updated %s with %d new action(s):\n", configFlag, len(newActions))
			for _, action := range newActions {
				fmt.Printf("  - %s\n", action)
			}
		} else {
			fmt.Printf("✓ No new actions found in workflows\n")
		}
	} else {
		fmt.Printf("✓ Created %s", configFlag)
		if len(newActions) > 0 {
			fmt.Printf(" with %d action(s)", len(newActions))
		}
		fmt.Println()
	}

	return nil
}

// discoverActions finds all actions in workflows and adds missing ones to config.
// Returns the list of newly added action names.
func discoverActions(cfg *config.Config, workflows []*workflow.Workflow) []string {
	var newActions []string

	for _, wf := range workflows {
		wfActions, err := wf.FindActions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse actions in %s: %v\n", wf.File, err)
			continue
		}

		for _, action := range wfActions {
			name := config.NormalizeActionName(action.Uses)
			if name == "" {
				continue
			}

			// Check if action already exists in config
			if cfg.Upgrade.Actions[name].Version == "" {
				cfg.SetActionConfig(name, config.DefaultActionConfig)
				newActions = append(newActions, name)
			}
		}
	}

	return newActions
}
