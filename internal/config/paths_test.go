package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobalConfigPath(t *testing.T) {
	path, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath() error: %v", err)
	}

	if path == "" {
		t.Error("GlobalConfigPath() returned empty string")
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("GlobalConfigPath() should end with config.yaml, got %q", filepath.Base(path))
	}

	// Should contain yar directory
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "yar" {
		t.Errorf("GlobalConfigPath() should be in yar directory, got %q", dir)
	}
}

func TestFindProjectConfigInCurrentDir(t *testing.T) {
	// Create temp directory with yar.yaml
	tmpDir := t.TempDir()
	yarPath := filepath.Join(tmpDir, "yar.yaml")
	if err := os.WriteFile(yarPath, []byte("project: test\n"), 0644); err != nil {
		t.Fatalf("failed to create test yar.yaml: %v", err)
	}

	// Find from current directory
	found, err := FindProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("FindProjectConfig() error: %v", err)
	}

	if found != yarPath {
		t.Errorf("FindProjectConfig() = %q, want %q", found, yarPath)
	}
}

func TestFindProjectConfigInParentDir(t *testing.T) {
	// Create temp directory structure: parent/child
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("failed to create child directory: %v", err)
	}

	// Put yar.yaml in parent
	yarPath := filepath.Join(tmpDir, "yar.yaml")
	if err := os.WriteFile(yarPath, []byte("project: test\n"), 0644); err != nil {
		t.Fatalf("failed to create test yar.yaml: %v", err)
	}

	// Find from child directory - should find parent's yar.yaml
	found, err := FindProjectConfig(childDir)
	if err != nil {
		t.Fatalf("FindProjectConfig() error: %v", err)
	}

	if found != yarPath {
		t.Errorf("FindProjectConfig() = %q, want %q", found, yarPath)
	}
}

func TestFindProjectConfigNotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir := t.TempDir()

	// Should return NotFoundError
	_, err := FindProjectConfig(tmpDir)
	if err == nil {
		t.Fatal("FindProjectConfig() should return error when no yar.yaml exists")
	}

	// Verify it's a NotFoundError by checking error message
	errStr := err.Error()
	if !containsStr(errStr, "yar.yaml") {
		t.Errorf("error should mention yar.yaml, got: %v", err)
	}
}

func TestFindProjectConfigFromWorkingDir(t *testing.T) {
	// Create temp directory with yar.yaml
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS has /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to eval symlinks: %v", err)
	}

	yarPath := filepath.Join(tmpDir, "yar.yaml")
	if err := os.WriteFile(yarPath, []byte("project: test\n"), 0644); err != nil {
		t.Fatalf("failed to create test yar.yaml: %v", err)
	}

	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Find with empty string (should use current directory)
	found, err := FindProjectConfig("")
	if err != nil {
		t.Fatalf("FindProjectConfig(\"\") error: %v", err)
	}

	if found != yarPath {
		t.Errorf("FindProjectConfig(\"\") = %q, want %q", found, yarPath)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
