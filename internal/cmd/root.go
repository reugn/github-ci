package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "github-ci",
	Short:         "A CLI tool for managing GitHub Actions workflows",
	Long:          `github-ci is a CLI tool that helps lint and upgrade GitHub Actions workflows.`,
	SilenceErrors: true,
}

// SetVersion sets the version string for the CLI.
func SetVersion(version string) {
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError("%v", err)
		os.Exit(1)
	}
}

// printError prints a formatted error message to stderr.
func printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "✗ Error: "+format+"\n", args...)
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(upgradeCmd)
}
