package upgrader

import (
	"context"
	"fmt"

	"github.com/reugn/github-ci/internal/actions"
	"github.com/reugn/github-ci/internal/config"
	"github.com/reugn/github-ci/internal/version"
	"github.com/reugn/github-ci/internal/workflow"
)

// Upgrader manages the upgrade process for GitHub Actions in workflow files.
type Upgrader struct {
	workflows  []*workflow.Workflow
	configFile string
	client     actions.Resolver
}

// updateInfo holds information about a pending action update.
type updateInfo struct {
	Workflow      *workflow.Workflow
	Action        *workflow.Action
	ActionInfo    *actions.ActionInfo
	CurrentTag    string
	NewTag        string
	NewHash       string
	VersionFormat string // "tag", "hash", or "major"
	Warning       string // Warning message if hash couldn't be resolved
}

// New creates a new Upgrader for the specified workflows directory.
func New(ctx context.Context, workflowsDir string) (*Upgrader, error) {
	workflows, err := workflow.LoadWorkflows(workflowsDir)
	if err != nil {
		return nil, err
	}
	return NewWithWorkflows(ctx, workflows, ""), nil
}

// NewWithWorkflows creates a new Upgrader with the provided workflows.
func NewWithWorkflows(ctx context.Context, workflows []*workflow.Workflow, configFile string) *Upgrader {
	return &Upgrader{
		workflows:  workflows,
		configFile: configFile,
		client:     actions.NewClientWithContext(ctx),
	}
}

// NewWithClient creates a new Upgrader with a custom actions client (for testing).
func NewWithClient(workflows []*workflow.Workflow, configFile string, client actions.Resolver) *Upgrader {
	return &Upgrader{
		workflows:  workflows,
		configFile: configFile,
		client:     client,
	}
}

// Upgrade upgrades GitHub Actions in all workflows to their latest versions.
func (u *Upgrader) Upgrade() error {
	cfg, err := u.loadAndInitConfig()
	if err != nil {
		return err
	}

	updates, err := u.findUpdates(cfg)
	if err != nil {
		return err
	}

	for _, upd := range updates {
		if err := u.applyUpdate(upd); err != nil {
			return err
		}
	}

	// Normalize comment spacing for all workflows
	u.normalizeAllCommentSpacing()

	return nil
}

// DryRun shows what would be updated without modifying files.
func (u *Upgrader) DryRun() error {
	cfg, err := config.LoadConfig(u.configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	updates, err := u.findUpdates(cfg)
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		fmt.Println("✓ No updates available")
		return nil
	}

	fmt.Printf("Would update %d action(s):\n\n", len(updates))
	for _, upd := range updates {
		u.printUpdate(upd)
	}

	return nil
}

// loadAndInitConfig loads the config and initializes missing action entries in memory.
// Does not save the config - use 'init --update' to persist new actions.
func (u *Upgrader) loadAndInitConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(u.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Discover actions and initialize config entries in memory
	for _, wf := range u.workflows {
		wfActions, err := wf.FindActions()
		if err != nil {
			return nil, fmt.Errorf("failed to find actions in %s: %w", wf.File, err)
		}

		for _, action := range wfActions {
			name := config.NormalizeActionName(action.Uses)
			if cfg.Upgrade.Actions[name].Version == "" {
				cfg.SetActionConfig(name, config.DefaultActionConfig)
			}
		}
	}

	return cfg, nil
}

// findUpdates scans all workflows and returns actions that need updating.
func (u *Upgrader) findUpdates(cfg *config.Config) ([]updateInfo, error) {
	var updates []updateInfo

	for _, wf := range u.workflows {
		wfActions, err := wf.FindActions()
		if err != nil {
			return nil, fmt.Errorf("failed to find actions in %s: %w", wf.File, err)
		}

		for _, action := range wfActions {
			upd, err := u.checkForUpdate(cfg, wf, action)
			if err != nil {
				return updates, err // Return partial results with error
			}
			if upd != nil {
				updates = append(updates, *upd)
			}
		}
	}

	return updates, nil
}

// checkForUpdate checks if an action needs updating and returns the update info.
// Returns nil if no update is needed, or an error if the check failed.
func (u *Upgrader) checkForUpdate(cfg *config.Config, wf *workflow.Workflow,
	action *workflow.Action) (*updateInfo, error) {
	actionInfo, err := actions.ParseActionUses(action.Uses)
	if err != nil {
		return nil, nil // Skip unparseable actions
	}

	actionName := config.NormalizeActionName(action.Uses)
	actionCfg := cfg.GetActionConfig(actionName)

	currentVersion, warning := u.resolveCurrentVersion(actionInfo)
	latestTag, latestHash, err := u.getLatestVersion(cfg, actionInfo, actionName,
		currentVersion, actionCfg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to check %s: %w", actionName, err)
	}

	// Skip if current hash already points to the latest version
	if actions.IsCommitHash(actionInfo.Ref) && actionInfo.Ref == latestHash {
		return nil, nil
	}

	// Check if format change is needed (e.g., tag → hash)
	versionFormat := cfg.GetVersionFormat()
	formatNeedsUpdate := needsFormatChange(actionInfo.Ref, versionFormat)

	// Determine version pattern for update check
	pattern := actionCfg.Version
	if _, exists := cfg.Upgrade.Actions[actionName]; !exists {
		pattern = "" // Allow any newer version if not in config
	}

	versionNeedsUpdate := config.ShouldUpdate(currentVersion, latestTag, pattern)

	// Update if either version or format needs changing
	if !versionNeedsUpdate && !formatNeedsUpdate {
		return nil, nil
	}

	return &updateInfo{
		Workflow:      wf,
		Action:        action,
		ActionInfo:    actionInfo,
		CurrentTag:    currentVersion,
		NewTag:        latestTag,
		NewHash:       latestHash,
		VersionFormat: cfg.GetVersionFormat(),
		Warning:       warning,
	}, nil
}

// resolveCurrentVersion resolves a commit hash to its tag if possible.
// Returns the resolved version and a warning message if the hash couldn't be resolved.
func (u *Upgrader) resolveCurrentVersion(info *actions.ActionInfo) (string, string) {
	if !actions.IsCommitHash(info.Ref) {
		return info.Ref, ""
	}

	tag, err := u.client.GetTagForCommit(info.Owner, info.Repo, info.Ref)
	if err != nil || tag == "" {
		refPreview := info.Ref
		if len(refPreview) > 12 {
			refPreview = refPreview[:12]
		}
		warning := fmt.Sprintf("cannot resolve hash %s to a tag (may be unreleased commit)",
			refPreview)
		return info.Ref, warning
	}
	return tag, ""
}

// getLatestVersion fetches the latest version based on config constraints.
func (u *Upgrader) getLatestVersion(cfg *config.Config, info *actions.ActionInfo, actionName,
	currentVersion, pattern string) (string, string, error) {
	if _, exists := cfg.Upgrade.Actions[actionName]; !exists {
		return u.client.GetLatestVersionUnconstrained(info.Owner, info.Repo)
	}
	return u.client.GetLatestVersion(info.Owner, info.Repo, currentVersion, pattern)
}

// applyUpdate applies a single update to the workflow file.
func (u *Upgrader) applyUpdate(upd updateInfo) error {
	newRef, comment := u.formatVersion(upd)

	// Build the new uses string, preserving path for composite actions
	var newUses string
	if upd.ActionInfo.Path != "" {
		newUses = fmt.Sprintf("%s/%s/%s@%s", upd.ActionInfo.Owner, upd.ActionInfo.Repo, upd.ActionInfo.Path, newRef)
	} else {
		newUses = fmt.Sprintf("%s/%s@%s", upd.ActionInfo.Owner, upd.ActionInfo.Repo, newRef)
	}
	if err := upd.Workflow.UpdateActionUses(upd.Action.Uses, newUses, comment); err != nil {
		return fmt.Errorf("failed to update action in %s: %w", upd.Workflow.File, err)
	}

	return nil
}

// formatVersion returns the new reference and optional comment based on version format.
func (u *Upgrader) formatVersion(upd updateInfo) (newRef, comment string) {
	switch upd.VersionFormat {
	case "hash":
		return upd.NewHash, upd.NewTag
	case "major":
		majorTag := version.ToMajorTag(upd.NewTag)
		return majorTag, upd.NewTag
	default: // "tag"
		return upd.NewTag, ""
	}
}

// needsFormatChange checks if the current ref format differs from the desired format.
func needsFormatChange(currentRef, desiredFormat string) bool {
	isHash := actions.IsCommitHash(currentRef)

	switch desiredFormat {
	case "hash":
		return !isHash // Need to change if current is not a hash
	case "major":
		// Need to change if current is a hash or a full version tag (e.g., v1.2.3)
		if isHash {
			return true
		}
		// Check if it's already a major-only tag (e.g., v1, v2)
		return len(currentRef) > 0 && currentRef[0] == 'v' &&
			len(currentRef) > 2 && (currentRef[2] == '.' || (len(currentRef) > 3 && currentRef[3] == '.'))
	default: // "tag"
		return isHash // Need to change if current is a hash
	}
}

// printUpdate prints information about a pending update.
func (u *Upgrader) printUpdate(upd updateInfo) {
	actionName := config.NormalizeActionName(upd.Action.Uses)
	fmt.Printf("  %s:%d\n", upd.Workflow.File, upd.Action.Line)
	fmt.Printf("    %s\n", upd.Action.Uses)

	newRef, comment := u.formatVersion(upd)
	if comment != "" {
		fmt.Printf("    → %s@%s (%s)\n", actionName, newRef, comment)
	} else {
		fmt.Printf("    → %s@%s\n", actionName, newRef)
	}

	if upd.Warning != "" {
		fmt.Printf("    ⚠ Warning: %s\n", upd.Warning)
	}
	fmt.Println()
}

// GetCacheStats returns cache statistics for GitHub API calls.
func (u *Upgrader) GetCacheStats() actions.CacheStats {
	return u.client.GetCacheStats()
}

// normalizeAllCommentSpacing normalizes comment spacing in all workflows.
func (u *Upgrader) normalizeAllCommentSpacing() {
	for _, wf := range u.workflows {
		if wf.NormalizeCommentSpacing() {
			_ = wf.Save()
		}
	}
}
