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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(upgradeCmd)
}
