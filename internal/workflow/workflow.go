package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workflow represents a GitHub Actions workflow file.
type Workflow struct {
	File     string   // Path to the workflow file
	Content  *Content // Parsed workflow structure
	RawBytes []byte   // Raw YAML bytes for manipulation
	node     *yaml.Node
}

// Content represents the parsed structure of a GitHub Actions workflow.
type Content struct {
	Name        string         `yaml:"name"`
	On          any            `yaml:"on"`
	Jobs        map[string]any `yaml:"jobs"`
	Permissions any            `yaml:"permissions"`
}

// Action represents a GitHub Action usage in a workflow file.
type Action struct {
	Uses string     // Action reference (e.g., "actions/checkout@v3")
	Line int        // Line number in the YAML file
	Node *yaml.Node // YAML node reference for updates
}

// Lines returns the workflow content as individual lines.
func (w *Workflow) Lines() []string {
	return strings.Split(string(w.RawBytes), "\n")
}

// BaseName returns the base name of the workflow file.
func (w *Workflow) BaseName() string {
	return filepath.Base(w.File)
}

// LoadWorkflows loads all workflow files from the specified directory.
func LoadWorkflows(dir string) ([]*Workflow, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows directory: %w", err)
	}

	workflows := make([]*Workflow, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		wf, err := LoadWorkflow(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load workflow %s: %w", entry.Name(), err)
		}
		workflows = append(workflows, wf)
	}

	return workflows, nil
}

// isYAMLFile checks if a filename has a YAML extension.
func isYAMLFile(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// LoadWorkflow loads a single workflow file.
func LoadWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var content Content
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &Workflow{
		File:     path,
		Content:  &content,
		RawBytes: data,
	}, nil
}

// getNode returns the cached YAML node, parsing if necessary.
func (w *Workflow) getNode() (*yaml.Node, error) {
	if w.node != nil {
		return w.node, nil
	}

	w.node = &yaml.Node{}
	if err := yaml.Unmarshal(w.RawBytes, w.node); err != nil {
		w.node = nil
		return nil, fmt.Errorf("failed to parse YAML node: %w", err)
	}
	return w.node, nil
}

// invalidateNode clears the cached node after modifications.
func (w *Workflow) invalidateNode() {
	w.node = nil
}

// FindActions finds all GitHub Actions used in the workflow.
func (w *Workflow) FindActions() ([]*Action, error) {
	node, err := w.getNode()
	if err != nil {
		return nil, err
	}

	var actions []*Action
	findActionsInNode(node, &actions)
	return actions, nil
}

// findActionsInNode recursively finds all "uses" keys in a YAML node tree.
func findActionsInNode(node *yaml.Node, actions *[]*Action) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, item := range node.Content {
			findActionsInNode(item, actions)
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "uses" && valueNode.Kind == yaml.ScalarNode {
				*actions = append(*actions, &Action{
					Uses: valueNode.Value,
					Line: valueNode.Line,
					Node: valueNode,
				})
			} else {
				findActionsInNode(valueNode, actions)
			}
		}

	case yaml.SequenceNode:
		for _, item := range node.Content {
			findActionsInNode(item, actions)
		}
	}
}

// HasPermissions returns true if the workflow has permissions configured.
func (w *Workflow) HasPermissions() bool {
	return w.Content.Permissions != nil
}

// Save writes the workflow to disk using the current RawBytes.
// This preserves original formatting including empty lines.
func (w *Workflow) Save() error {
	return os.WriteFile(w.File, w.RawBytes, 0600)
}

// UpdateActionUses updates an action reference and optionally adds a comment.
// Uses line-based replacement to preserve original formatting including empty lines.
func (w *Workflow) UpdateActionUses(oldUses, newUses, comment string) error {
	lines := strings.Split(string(w.RawBytes), "\n")
	updated := false

	for i, line := range lines {
		// Find lines containing the old uses value
		if strings.Contains(line, oldUses) {
			// Replace the uses value
			newLine := strings.Replace(line, oldUses, newUses, 1)

			// Remove any existing line comment and trailing whitespace
			if idx := strings.Index(newLine, " #"); idx != -1 {
				newLine = newLine[:idx]
			}
			newLine = strings.TrimRight(newLine, " \t")
			if comment != "" {
				newLine += " # " + comment
			}

			lines[i] = newLine
			updated = true
		}
	}

	if !updated {
		return fmt.Errorf("action %s not found", oldUses)
	}

	w.RawBytes = []byte(strings.Join(lines, "\n"))
	w.invalidateNode()

	return os.WriteFile(w.File, w.RawBytes, 0600)
}

// NormalizeCommentSpacing normalizes spacing before version tag comments on uses: lines.
// Only affects comments that look like version tags (e.g., "# v1.0.0").
// Ensures exactly 1 space before the # character.
func (w *Workflow) NormalizeCommentSpacing() bool {
	lines := strings.Split(string(w.RawBytes), "\n")
	modified := false

	for i, line := range lines {
		// Only process lines containing "uses:" with a comment
		if !strings.Contains(line, "uses:") {
			continue
		}
		idx := strings.Index(line, " #")
		if idx == -1 {
			continue
		}

		// Extract the comment text (after #)
		commentStart := strings.Index(line[idx:], "#")
		comment := strings.TrimSpace(line[idx+commentStart+1:])

		// Only normalize if comment looks like a version tag (e.g., v1.0.0, v2, 1.0.0)
		if !isVersionTag(comment) {
			continue
		}

		// Extract content before comment
		beforeComment := strings.TrimRight(line[:idx], " \t")
		fullComment := line[idx+commentStart:]

		// Reconstruct with exactly 1 space
		newLine := beforeComment + " " + fullComment
		if newLine != line {
			lines[i] = newLine
			modified = true
		}
	}

	if modified {
		w.RawBytes = []byte(strings.Join(lines, "\n"))
		w.invalidateNode()
	}

	return modified
}

// isVersionTag checks if a string looks like a version tag (e.g., v1.0.0, v2, 1.0.0).
func isVersionTag(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Handle v-prefixed versions (v1, v1.0, v1.0.0)
	if s[0] == 'v' || s[0] == 'V' {
		if len(s) > 1 && s[1] >= '0' && s[1] <= '9' {
			return true
		}
	}
	// Handle non-prefixed versions (1.0.0)
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	return false
}

// FindJobLine finds the line number where a job is defined.
func (w *Workflow) FindJobLine(jobID string) int {
	lines := w.Lines()
	prefix := "  " + jobID + ":"
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return i + 1
		}
	}
	return 0
}

// FindStepLine finds the line number where a step is defined within a job.
func (w *Workflow) FindStepLine(jobID string, stepIndex int) int {
	lines := w.Lines()
	inJob := false
	inSteps := false
	stepCount := 0
	jobPrefix := "  " + jobID + ":"

	for i, line := range lines {
		if strings.HasPrefix(line, jobPrefix) {
			inJob = true
			continue
		}

		if inJob {
			trimmed := strings.TrimSpace(line)
			// Check if we've exited the job
			if len(line) > 0 && !strings.HasPrefix(line, " ") {
				break
			}
			if strings.HasPrefix(trimmed, "steps:") {
				inSteps = true
				continue
			}
			if inSteps && strings.HasPrefix(trimmed, "- ") {
				if stepCount == stepIndex {
					return i + 1
				}
				stepCount++
			}
		}
	}

	return 0
}

// ExtractWorkflowEnv extracts workflow-level env variable names.
func (w *Workflow) ExtractWorkflowEnv() map[string]bool {
	result := make(map[string]bool)
	lines := w.Lines()
	inEnv := false
	envIndent := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))

		// Check for top-level env:
		if trimmed == "env:" && indent == 0 {
			inEnv = true
			envIndent = indent
			continue
		}

		if inEnv {
			// Exit env block if we hit a top-level key
			if indent == 0 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				break
			}

			// Extract env var name
			if indent > envIndent && strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) > 0 {
					result[strings.TrimSpace(parts[0])] = true
				}
			}
		}
	}

	return result
}
