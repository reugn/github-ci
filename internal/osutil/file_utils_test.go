package osutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with existing file
	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(existingFile) {
		t.Error("FileExists() = false for existing file, want true")
	}

	// Test with non-existing file
	nonExistingFile := filepath.Join(tmpDir, "non-existing.txt")
	if FileExists(nonExistingFile) {
		t.Error("FileExists() = true for non-existing file, want false")
	}

	// Test with directory
	if !FileExists(tmpDir) {
		t.Error("FileExists() = false for existing directory, want true")
	}

	// Test with empty path
	if FileExists("") {
		t.Error("FileExists() = true for empty path, want false")
	}
}
