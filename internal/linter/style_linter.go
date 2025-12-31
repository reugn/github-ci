package linter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/stringutil"
	"github.com/reugn/github-ci/internal/workflow"
)

// Context labels for style issue messages.
const (
	ctxWorkflow = "Workflow"
	ctxJob      = "Job"
	ctxStep     = "Step"
)

// StyleLinter checks for style and naming convention issues in workflow files.
type StyleLinter struct {
	noOpFixer
	settings *config.StyleSettings
}

// NewStyleLinter creates a new StyleLinter instance.
func NewStyleLinter(settings *config.StyleSettings) *StyleLinter {
	if settings == nil {
		settings = config.DefaultStyleSettings()
	}
	return &StyleLinter{settings: settings}
}

// LintWorkflow checks a single workflow for style issues.
func (l *StyleLinter) LintWorkflow(wf *workflow.Workflow) ([]*Issue, error) {
	var issues []*Issue
	file := wf.BaseName()

	// Check workflow-level issues
	issues = append(issues, l.checkWorkflowName(wf, file)...)

	// Check job-level issues
	issues = append(issues, l.checkJobs(wf, file)...)

	// Check env consistency
	issues = append(issues, l.checkEnvConsistency(wf, file)...)

	return issues, nil
}

// checkWorkflowName checks for missing or invalid workflow name.
func (l *StyleLinter) checkWorkflowName(wf *workflow.Workflow, file string) []*Issue {
	var issues []*Issue

	if wf.Content == nil || wf.Content.Name == "" {
		issues = append(issues, newIssue(file, 1, "Workflow is missing a name"))
	} else {
		if issue := l.checkNameLength(wf.Content.Name, file, 1, ctxWorkflow); issue != nil {
			issues = append(issues, issue)
		}
		if issue := l.checkNamingConvention(wf.Content.Name, file, 1, ctxWorkflow); issue != nil {
			issues = append(issues, issue)
		}
	}

	return issues
}

// checkJobs checks job-level and step-level style issues.
func (l *StyleLinter) checkJobs(wf *workflow.Workflow, file string) []*Issue {
	var issues []*Issue

	if wf.Content == nil || wf.Content.Jobs == nil {
		return issues
	}

	for jobID, jobData := range wf.Content.Jobs {
		job, ok := jobData.(map[string]any)
		if !ok {
			continue
		}

		jobLine := wf.FindJobLine(jobID)

		// Get job name
		jobName, _ := job["name"].(string)

		// Check for cryptic job ID without explicit name
		if jobName == "" && stringutil.IsCrypticName(jobID) {
			message := fmt.Sprintf("Job '%s' has cryptic ID and is missing a name", jobID)
			issues = append(issues, newIssue(file, jobLine, message))
		}

		// Check job name length and convention
		if jobName != "" {
			if issue := l.checkNameLength(jobName, file, jobLine, ctxJob); issue != nil {
				issues = append(issues, issue)
			}
			if issue := l.checkNamingConvention(jobName, file, jobLine, ctxJob); issue != nil {
				issues = append(issues, issue)
			}
		}

		// Check steps
		issues = append(issues, l.checkSteps(wf, job, file, jobID)...)
	}

	return issues
}

// checkSteps checks step-level style issues.
func (l *StyleLinter) checkSteps(wf *workflow.Workflow, job map[string]any, file string, jobID string) []*Issue {
	var issues []*Issue

	stepsData, ok := job["steps"].([]any)
	if !ok {
		return issues
	}

	lines := wf.Lines()
	checkoutFound := false
	for i, stepData := range stepsData {
		step, ok := stepData.(map[string]any)
		if !ok {
			continue
		}

		stepLine := wf.FindStepLine(jobID, i)

		// Get step properties
		stepName, _ := step["name"].(string)
		stepUses, _ := step["uses"].(string)

		// Check missing step name (only if configured)
		if stepName == "" {
			if l.settings.RequireStepNames {
				issues = append(issues, newIssue(file, stepLine, "Step is missing a name"))
			}
		} else {
			if issue := l.checkNameLength(stepName, file, stepLine, ctxStep); issue != nil {
				issues = append(issues, issue)
			}
			if issue := l.checkNamingConvention(stepName, file, stepLine, ctxStep); issue != nil {
				issues = append(issues, issue)
			}
		}

		// Check if checkout should be first (configurable)
		if l.settings.CheckoutFirst {
			isCheckout := strings.Contains(stepUses, "actions/checkout")
			if isCheckout && !checkoutFound && i > 0 {
				message := "Checkout action should typically be the first step"
				issues = append(issues, newIssue(file, stepLine, message))
			}
			if isCheckout {
				checkoutFound = true
			}
		}

		// Check if name comes first in step
		if issue := l.checkNameFirst(lines, stepLine, file); issue != nil {
			issues = append(issues, issue)
		}

		// Check run script length
		if issue := l.checkRunLength(step, file, stepLine); issue != nil {
			issues = append(issues, issue)
		}
	}

	return issues
}

// checkEnvConsistency checks for env variable placement issues.
func (l *StyleLinter) checkEnvConsistency(wf *workflow.Workflow, file string) []*Issue {
	var issues []*Issue

	// Extract workflow-level env
	workflowEnv := wf.ExtractWorkflowEnv()

	if wf.Content == nil || wf.Content.Jobs == nil {
		return issues
	}

	for jobID, jobData := range wf.Content.Jobs {
		job, ok := jobData.(map[string]any)
		if !ok {
			continue
		}

		jobLine := wf.FindJobLine(jobID)
		jobEnv := extractJobEnv(job)

		// Check for shadowed variables
		for varName := range jobEnv {
			if workflowEnv[varName] {
				message := fmt.Sprintf("Job env var '%s' shadows workflow-level env var", varName)
				issues = append(issues, newIssue(file, jobLine, message))
			}
		}
	}

	return issues
}

// checkNameLength validates name length.
func (l *StyleLinter) checkNameLength(name, file string, line int, context string) *Issue {
	var msg string
	if l.settings.MinNameLength > 0 && len(name) < l.settings.MinNameLength {
		msg = fmt.Sprintf("%s name '%s' is too short (min %d chars)", context, name, l.settings.MinNameLength)
	} else if l.settings.MaxNameLength > 0 && len(name) > l.settings.MaxNameLength {
		msg = fmt.Sprintf("%s name exceeds maximum length of %d characters", context, l.settings.MaxNameLength)
	}

	return newIssue(file, line, msg)
}

// checkNamingConvention validates naming convention.
func (l *StyleLinter) checkNamingConvention(name, file string, line int, context string) *Issue {
	if l.settings.NamingConvention == "" {
		return nil
	}

	words := strings.Fields(name)
	if len(words) == 0 {
		return nil
	}

	var msg string
	switch l.settings.NamingConvention {
	case "title":
		// Title Case: Every Word Starts With Uppercase
		for _, word := range words {
			if len(word) > 0 && !unicode.IsUpper(rune(word[0])) {
				msg = fmt.Sprintf("%s name should use Title Case", context)
				break
			}
		}
	case "sentence":
		// Sentence case: First word must start with uppercase
		if len(words[0]) > 0 && !unicode.IsUpper(rune(words[0][0])) {
			msg = fmt.Sprintf("%s name should start with uppercase (sentence case)", context)
		}
	}

	return newIssue(file, line, msg)
}

// checkNameFirst checks if 'name:' comes first in a step definition.
func (l *StyleLinter) checkNameFirst(lines []string, stepLine int, file string) *Issue {
	if stepLine <= 0 || stepLine > len(lines) {
		return nil
	}

	// Find the step start (line with '- ')
	idx := stepLine - 1
	if idx >= len(lines) {
		return nil
	}

	line := lines[idx]
	trimmed := strings.TrimSpace(line)

	// If step starts with '- name:', it's correct
	if strings.HasPrefix(trimmed, "- name:") {
		return nil
	}

	// If step starts with something else and has a name somewhere
	if strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "- name:") {
		// Look ahead for 'name:' in this step
		indent := stringutil.CountLeadingSpaces(line)
		for i := idx + 1; i < len(lines) && i < idx+10; i++ {
			nextLine := lines[i]
			nextIndent := stringutil.CountLeadingSpaces(nextLine)
			nextTrimmed := strings.TrimSpace(nextLine)

			// Stop if we hit another step or less indented line
			if strings.HasPrefix(nextTrimmed, "- ") || nextIndent <= indent {
				break
			}

			if strings.HasPrefix(nextTrimmed, "name:") {
				return newIssue(file, stepLine, "Step 'name' should come first before other fields")
			}
		}
	}

	return nil
}

// checkRunLength checks if a run script exceeds the maximum line count.
func (l *StyleLinter) checkRunLength(step map[string]any, file string, line int) *Issue {
	if l.settings.MaxRunLines <= 0 {
		return nil
	}

	runScript, ok := step["run"].(string)
	if !ok || runScript == "" {
		return nil
	}

	lineCount := strings.Count(strings.TrimSpace(runScript), "\n") + 1
	if lineCount > l.settings.MaxRunLines {
		msg := fmt.Sprintf("Run script has %d lines (max %d); consider extracting to a script file",
			lineCount, l.settings.MaxRunLines)
		return newIssue(file, line, msg)
	}

	return nil
}

// extractJobEnv extracts env variable names from a job map.
func extractJobEnv(job map[string]any) map[string]bool {
	result := make(map[string]bool)
	env, ok := job["env"].(map[string]any)
	if !ok {
		return result
	}
	for key := range env {
		result[key] = true
	}
	return result
}
