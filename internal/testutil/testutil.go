package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateWorkflow creates a test workflow file with the given content.
func CreateWorkflow(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test workflow: %v", err)
	}
	return path
}

// CreateConfig creates a test config file with the given content.
func CreateConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".github-ci.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	return path
}
