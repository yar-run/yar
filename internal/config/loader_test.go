package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yar-run/yar/internal/errors"
)

func TestNewLoader(t *testing.T) {
	l := NewLoader()
	if l == nil {
		t.Fatal("NewLoader() returned nil")
	}
}

func TestNewLoaderWithGlobalPath(t *testing.T) {
	customPath := "/custom/config.yaml"
	l := NewLoader(WithGlobalPath(customPath))

	path, err := l.GlobalPath()
	if err != nil {
		t.Fatalf("GlobalPath() error: %v", err)
	}
	if path != customPath {
		t.Errorf("GlobalPath() = %q, want %q", path, customPath)
	}
}

func TestNewLoaderWithProjectPath(t *testing.T) {
	customPath := "/custom/yar.yaml"
	l := NewLoader(WithProjectPath(customPath))

	path, err := l.ProjectPath()
	if err != nil {
		t.Fatalf("ProjectPath() error: %v", err)
	}
	if path != customPath {
		t.Errorf("ProjectPath() = %q, want %q", path, customPath)
	}
}

func TestLoadGlobalMissingFile(t *testing.T) {
	// Use non-existent path
	l := NewLoader(WithGlobalPath("/non/existent/config.yaml"))

	cfg, err := l.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() should return defaults, got error: %v", err)
	}

	// Should return defaults
	if cfg.Container != "colima" {
		t.Errorf("Container = %q, want %q (default)", cfg.Container, "colima")
	}
}

func TestLoadGlobalValidFile(t *testing.T) {
	l := NewLoader(WithGlobalPath("testdata/valid/config.yaml"))

	cfg, err := l.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}

	if cfg.Container != "colima" {
		t.Errorf("Container = %q, want %q", cfg.Container, "colima")
	}
	if cfg.VPN == nil || cfg.VPN.Provider != "openvpn" {
		t.Errorf("VPN.Provider = %v, want openvpn", cfg.VPN)
	}
	if cfg.Network == nil || cfg.Network.CIDR != "172.16.34.0/23" {
		t.Errorf("Network.CIDR = %v, want 172.16.34.0/23", cfg.Network)
	}
}

func TestLoadGlobalMinimalFile(t *testing.T) {
	l := NewLoader(WithGlobalPath("testdata/valid/config-minimal.yaml"))

	cfg, err := l.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}

	if cfg.Container != "docker" {
		t.Errorf("Container = %q, want %q", cfg.Container, "docker")
	}
}

func TestLoadGlobalBadYAML(t *testing.T) {
	l := NewLoader(WithGlobalPath("testdata/invalid/config-bad-yaml.yaml"))

	_, err := l.LoadGlobal()
	if err == nil {
		t.Fatal("LoadGlobal() should return error for invalid YAML")
	}

	// Should be a ConfigError
	var cfgErr *errors.ConfigError
	if !asConfigError(err, &cfgErr) {
		t.Errorf("expected ConfigError, got %T: %v", err, err)
	}
}

func TestLoadGlobalBadSchema(t *testing.T) {
	l := NewLoader(WithGlobalPath("testdata/invalid/config-bad-schema.yaml"))

	_, err := l.LoadGlobal()
	if err == nil {
		t.Fatal("LoadGlobal() should return error for invalid schema")
	}

	// Should be a ValidationError
	var valErr *errors.ValidationError
	if !asValidationError(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestLoadProjectNotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir := t.TempDir()
	l := NewLoader(WithProjectPath(filepath.Join(tmpDir, "yar.yaml")))

	_, err := l.LoadProject()
	if err == nil {
		t.Fatal("LoadProject() should return error when file doesn't exist")
	}

	// Should be a NotFoundError
	var nfErr *errors.NotFoundError
	if !asNotFoundError(err, &nfErr) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestLoadProjectValidFile(t *testing.T) {
	l := NewLoader(WithProjectPath("testdata/valid/project.yaml"))

	proj, err := l.LoadProject()
	if err != nil {
		t.Fatalf("LoadProject() error: %v", err)
	}

	if proj.Project != "my-backend" {
		t.Errorf("Project = %q, want %q", proj.Project, "my-backend")
	}
	if len(proj.Services) != 3 {
		t.Errorf("len(Services) = %d, want 3", len(proj.Services))
	}
	if proj.Environments["prod"].Secrets != "azure" {
		t.Errorf("Environments[prod].Secrets = %q, want azure", proj.Environments["prod"].Secrets)
	}
}

func TestLoadProjectMinimalFile(t *testing.T) {
	l := NewLoader(WithProjectPath("testdata/valid/project-minimal.yaml"))

	proj, err := l.LoadProject()
	if err != nil {
		t.Fatalf("LoadProject() error: %v", err)
	}

	if proj.Project != "minimal" {
		t.Errorf("Project = %q, want %q", proj.Project, "minimal")
	}
}

func TestLoadProjectMissingRequiredField(t *testing.T) {
	l := NewLoader(WithProjectPath("testdata/invalid/project-missing-name.yaml"))

	_, err := l.LoadProject()
	if err == nil {
		t.Fatal("LoadProject() should return error for missing required field")
	}

	// Should be a ValidationError
	var valErr *errors.ValidationError
	if !asValidationError(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestLoadProjectSearchesParentDirs(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child", "grandchild")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	// Put yar.yaml in root
	yarPath := filepath.Join(tmpDir, "yar.yaml")
	content := `project: test-parent
environments:
  local:
    cluster: local
    secrets: pass
services:
  - name: redis
    pack: redis
`
	if err := os.WriteFile(yarPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create yar.yaml: %v", err)
	}

	// Save current dir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(origDir)

	// Change to grandchild
	if err := os.Chdir(childDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Load without explicit path - should find parent's yar.yaml
	l := NewLoader()
	proj, err := l.LoadProject()
	if err != nil {
		t.Fatalf("LoadProject() error: %v", err)
	}

	if proj.Project != "test-parent" {
		t.Errorf("Project = %q, want %q", proj.Project, "test-parent")
	}
}

// Helper functions for type assertions
func asConfigError(err error, target **errors.ConfigError) bool {
	if e, ok := err.(*errors.ConfigError); ok {
		*target = e
		return true
	}
	return false
}

func asValidationError(err error, target **errors.ValidationError) bool {
	if e, ok := err.(*errors.ValidationError); ok {
		*target = e
		return true
	}
	return false
}

func asNotFoundError(err error, target **errors.NotFoundError) bool {
	if e, ok := err.(*errors.NotFoundError); ok {
		*target = e
		return true
	}
	return false
}
