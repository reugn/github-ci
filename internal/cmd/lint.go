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
	printIssues("Issues:", issues)

	// Only suggest --fix if at least one issue can be auto-fixed
	if hasFixableIssues(issues) {
		fmt.Println("\nRun with --fix to automatically fix some issues")
	}

	printIssueSummary(len(issues))
	return issuesExitCode
}

// doLintWithFix applies fixes and prints results in two sections.
// Returns exit code 0 if all issues are fixed, issuesExitCode if some remain.
func doLintWithFix(l *linter.WorkflowLinter, issues []*linter.Issue, issuesExitCode int) int {
	// Apply fixes
	if err := l.Fix(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to fix workflows: %v\n", err)
		return 1
	}

	// Re-lint to see what issues remain after fixing
	remainingIssues, err := l.Lint()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to re-lint workflows: %v\n", err)
		return 1
	}

	fixed, unfixed := classifyIssues(issues, remainingIssues)

	printIssues("Fixed:", fixed)
	printIssuesSeparator(fixed, unfixed)
	printIssues("Issues:", unfixed)

	stats := l.GetCacheStats()
	printCacheStats(stats.Hits, stats.Misses)
	printIssueSummary(len(unfixed))

	if len(unfixed) > 0 {
		return issuesExitCode
	}
	return 0
}

// printIssue prints a single issue with indentation.
func printIssue(issue *linter.Issue) {
	fmt.Printf("  %s\n", issue)
}

// printIssues prints a labeled section of issues.
func printIssues(header string, issues []*linter.Issue) {
	if len(issues) == 0 {
		return
	}

	fmt.Println(header)
	for _, issue := range issues {
		printIssue(issue)
	}
}

// printIssuesSeparator prints a blank line if all provided slices are non-empty.
func printIssuesSeparator(slices ...[]*linter.Issue) {
	for _, s := range slices {
		if len(s) == 0 {
			return
		}
	}
	fmt.Println()
}

// printIssueSummary prints the total issue count.
func printIssueSummary(count int) {
	fmt.Printf("\n%d issue(s).\n", count)
}

// classifyIssues separates issues into fixed and unfixed based on what remains after fixing.
func classifyIssues(original, remaining []*linter.Issue) (fixed, unfixed []*linter.Issue) {
	remainingKeys := make(map[string]bool)
	for _, issue := range remaining {
		remainingKeys[issue.Key()] = true
	}

	for _, issue := range original {
		if remainingKeys[issue.Key()] {
			unfixed = append(unfixed, issue)
		} else {
			fixed = append(fixed, issue)
		}
	}

	return fixed, unfixed
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
