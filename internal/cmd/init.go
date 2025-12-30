package cmd

import (
	"fmt"
	"os"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/osutil"
	"github.com/reugn/github-ci/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	updateFlag   bool
	defaultsFlag bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a new .github-ci.yaml configuration file with default settings.

If the configuration file already exists:
  - Without --update: fails with an error
  - With --update: adds any new actions found in workflows to the config

Use --defaults to include all linter settings and scan workflows to discover
actions with default version patterns.`,
	RunE:         runInit,
	SilenceUsage: true,
}

func init() {
	addCommonFlags(initCmd)
	initCmd.Flags().BoolVarP(&updateFlag, "update", "u", false,
		"Update existing config with new actions from workflows")
	initCmd.Flags().BoolVarP(&defaultsFlag, "defaults", "d", false,
		"Include all linter settings and discover actions from workflows")
}

func runInit(_ *cobra.Command, _ []string) error {
	configExists := osutil.FileExists(configFlag)

	// Check if config exists and we're not updating
	if configExists && !updateFlag {
		return fmt.Errorf("config file %s already exists (use --update to add new actions)", configFlag)
	}

	// Load or create the configuration file
	cfg, err := resolveConfig(configExists)
	if err != nil {
		return err
	}

	// Discover actions if --defaults or --update is set
	newActions, err := scanActions(cfg, configExists)
	if err != nil {
		return err
	}

	// Save the config
	if err := config.SaveConfig(cfg, configFlag); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print the result
	printResult(configExists, newActions)

	return nil
}

// resolveConfig returns the appropriate config based on flags and file existence.
func resolveConfig(exists bool) (*config.Config, error) {
	switch {
	case exists:
		cfg, err := config.LoadConfig(configFlag)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing config: %w", err)
		}
		return cfg, nil
	case defaultsFlag:
		return config.NewFullDefaultConfig(), nil
	default:
		return config.NewDefaultConfig(), nil
	}
}

// printResult outputs the init command result.
func printResult(configExists bool, newActions []string) {
	switch {
	case configExists && len(newActions) > 0:
		fmt.Printf("✓ Updated %s with %d new action(s):\n", configFlag, len(newActions))
		for _, action := range newActions {
			fmt.Printf("  - %s\n", action)
		}
	case configExists:
		fmt.Printf("✓ No new actions found in workflows\n")
	case len(newActions) > 0:
		fmt.Printf("✓ Created %s with %d action(s)\n", configFlag, len(newActions))
	default:
		fmt.Printf("✓ Created %s\n", configFlag)
	}
}

// scanActions discovers actions from workflows when --defaults or --update is set.
func scanActions(cfg *config.Config, configExists bool) ([]string, error) {
	if !defaultsFlag && !updateFlag {
		return nil, nil
	}

	workflows, err := loadWorkflows(pathFlag)
	if err != nil {
		if configExists {
			return nil, fmt.Errorf("failed to load workflows: %w", err)
		}
		return nil, nil
	}

	return discoverActions(cfg, workflows), nil
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
