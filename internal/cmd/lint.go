package cmd

import (
	"fmt"
	"os"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/linter"
	"github.com/reugn/github-ci/internal/workflow"
	"github.com/spf13/cobra"
)

var fixFlag bool

var lintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Lint GitHub Actions workflows",
	Long: `Analyze workflows for common issues using configurable linters:
- permissions: Missing permissions configuration
- versions: Actions using version tags instead of commit hashes
- format: Formatting issues (indentation, line length, trailing whitespace)
- secrets: Hardcoded secrets and sensitive information
- injection: Shell injection vulnerabilities from untrusted input

The path can be a directory (e.g., .github/workflows) or a specific workflow file.
If no path is provided, defaults to .github/workflows.

Configure enabled linters in .github-ci.yaml.`,
	RunE:         runLint,
	SilenceUsage: true,
}

func init() {
	addCommonFlags(lintCmd)
	lintCmd.Flags().BoolVar(&fixFlag, "fix", false,
		"Automatically fix issues by replacing version tags with commit hashes")
}

func runLint(_ *cobra.Command, args []string) error {
	workflowsPath := pathFlag
	if len(args) > 0 {
		workflowsPath = args[0]
	}

	workflows, err := loadWorkflows(workflowsPath)
	if err != nil {
		return fmt.Errorf("failed to load workflows: %w", err)
	}

	exitCode := doLint(workflows, configFlag)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

// doLint performs linting and returns the exit code.
func doLint(workflows []*workflow.Workflow, configFile string) int {
	ctx, cancel := createTimeoutContext(configFile)
	defer cancel()

	// Load config to get exit code setting
	cfg, _ := config.LoadConfig(configFile)
	issuesExitCode := cfg.GetIssuesExitCode()

	l := linter.NewWithWorkflows(ctx, workflows, configFile)

	issues, err := l.Lint()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to lint workflows: %v\n", err)
		return 1
	}

	if len(issues) == 0 {
		fmt.Println("0 issues.")
		return 0
	}

	if fixFlag {
		return doLintWithFix(l, issues, issuesExitCode)
	}

	// Print all issues
	fmt.Println("Issues:")
	for _, issue := range issues {
		printIssue(issue)
	}

	// Only suggest --fix if at least one issue can be auto-fixed
	if hasFixableIssues(issues) {
		fmt.Println("\nRun with --fix to automatically fix some issues")
	}

	fmt.Printf("\n%d issue(s).\n", len(issues))
	return issuesExitCode
}

// doLintWithFix applies fixes and prints results in two sections.
// Returns exit code 0 if all issues are fixed, issuesExitCode if some remain.
func doLintWithFix(l *linter.WorkflowLinter, issues []*linter.Issue, issuesExitCode int) int {
	// Separate fixable and unfixable issues
	var fixable, unfixable []*linter.Issue
	for _, issue := range issues {
		if linter.SupportsAutoFix(issue.Linter) {
			fixable = append(fixable, issue)
		} else {
			unfixable = append(unfixable, issue)
		}
	}

	// Apply fixes
	if err := l.Fix(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to fix workflows: %v\n", err)
		return 1
	}

	// Print fixed issues
	if len(fixable) > 0 {
		fmt.Println("Fixed:")
		for _, issue := range fixable {
			printIssue(issue)
		}
	}

	// Print remaining issues
	if len(unfixable) > 0 {
		if len(fixable) > 0 {
			fmt.Println()
		}
		fmt.Println("Issues:")
		for _, issue := range unfixable {
			printIssue(issue)
		}
	}

	stats := l.GetCacheStats()
	printCacheStats(stats.Hits, stats.Misses)
	fmt.Printf("\n%d issue(s).\n", len(unfixable))

	if len(unfixable) > 0 {
		return issuesExitCode
	}
	return 0
}

// printIssue prints a single issue.
func printIssue(issue *linter.Issue) {
	if issue.Line > 0 {
		fmt.Printf("  %s:%d: (%s) %s\n", issue.File, issue.Line, issue.Linter, issue.Message)
	} else {
		fmt.Printf("  %s: (%s) %s\n", issue.File, issue.Linter, issue.Message)
	}
}

// hasFixableIssues returns true if any issue can be auto-fixed.
func hasFixableIssues(issues []*linter.Issue) bool {
	for _, issue := range issues {
		if linter.SupportsAutoFix(issue.Linter) {
			return true
		}
	}
	return false
}
