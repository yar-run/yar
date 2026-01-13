package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPlatform(t *testing.T) {
	t.Run("returns valid OS constant", func(t *testing.T) {
		got := Platform()
		valid := got == Darwin || got == Linux || got == Windows
		if !valid {
			t.Errorf("Platform() = %q, want one of darwin, linux, windows", got)
		}
	})

	t.Run("matches runtime.GOOS", func(t *testing.T) {
		got := Platform()
		want := OS(runtime.GOOS)
		if got != want {
			t.Errorf("Platform() = %q, want %q", got, want)
		}
	})
}

func TestHomeDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		got, err := HomeDir()
		if err != nil {
			t.Fatalf("HomeDir() error = %v", err)
		}
		if got == "" {
			t.Error("HomeDir() returned empty string")
		}
	})

	t.Run("returns absolute path", func(t *testing.T) {
		got, err := HomeDir()
		if err != nil {
			t.Fatalf("HomeDir() error = %v", err)
		}
		if !filepath.IsAbs(got) {
			t.Errorf("HomeDir() = %q, want absolute path", got)
		}
	})
}

func TestConfigDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		if got == "" {
			t.Error("ConfigDir() returned empty string")
		}
	})

	t.Run("ends with yar", func(t *testing.T) {
		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		if filepath.Base(got) != "yar" {
			t.Errorf("ConfigDir() = %q, want path ending in 'yar'", got)
		}
	})

	t.Run("respects XDG_CONFIG_HOME", func(t *testing.T) {
		// Skip on Windows where XDG is not standard
		if runtime.GOOS == "windows" {
			t.Skip("XDG_CONFIG_HOME not used on Windows")
		}

		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		want := filepath.Join(tmpDir, "yar")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})
}

func TestCacheDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		if got == "" {
			t.Error("CacheDir() returned empty string")
		}
	})

	t.Run("contains yar", func(t *testing.T) {
		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		if !strings.Contains(got, "yar") {
			t.Errorf("CacheDir() = %q, want path containing 'yar'", got)
		}
	})

	t.Run("respects XDG_CACHE_HOME", func(t *testing.T) {
		// Skip on Windows where XDG is not standard
		if runtime.GOOS == "windows" {
			t.Skip("XDG_CACHE_HOME not used on Windows")
		}

		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)

		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		want := filepath.Join(tmpDir, "yar")
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})
}

func TestDataDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		if got == "" {
			t.Error("DataDir() returned empty string")
		}
	})

	t.Run("contains yar", func(t *testing.T) {
		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		if !strings.Contains(got, "yar") {
			t.Errorf("DataDir() = %q, want path containing 'yar'", got)
		}
	})

	t.Run("respects XDG_DATA_HOME", func(t *testing.T) {
		// Skip on Windows where XDG is not standard
		if runtime.GOOS == "windows" {
			t.Skip("XDG_DATA_HOME not used on Windows")
		}

		tmpDir := t.TempDir()
		t.Setenv("XDG_DATA_HOME", tmpDir)

		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		want := filepath.Join(tmpDir, "yar")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})
}

func TestExpandPath(t *testing.T) {
	t.Run("expands tilde to home directory", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() error = %v", err)
		}

		got, err := ExpandPath("~/config/yar")
		if err != nil {
			t.Fatalf("ExpandPath() error = %v", err)
		}
		want := filepath.Join(home, "config", "yar")
		if got != want {
			t.Errorf("ExpandPath(~/config/yar) = %q, want %q", got, want)
		}
	})

	t.Run("expands environment variables", func(t *testing.T) {
		t.Setenv("TEST_YAR_VAR", "testvalue")

		got, err := ExpandPath("/path/$TEST_YAR_VAR/config")
		if err != nil {
			t.Fatalf("ExpandPath() error = %v", err)
		}
		want := "/path/testvalue/config"
		if got != want {
			t.Errorf("ExpandPath() = %q, want %q", got, want)
		}
	})

	t.Run("expands braced environment variables", func(t *testing.T) {
		t.Setenv("TEST_YAR_BRACED", "bracedvalue")

		got, err := ExpandPath("/path/${TEST_YAR_BRACED}/config")
		if err != nil {
			t.Fatalf("ExpandPath() error = %v", err)
		}
		want := "/path/bracedvalue/config"
		if got != want {
			t.Errorf("ExpandPath() = %q, want %q", got, want)
		}
	})

	t.Run("handles paths without expansion", func(t *testing.T) {
		got, err := ExpandPath("/absolute/path/no/expansion")
		if err != nil {
			t.Fatalf("ExpandPath() error = %v", err)
		}
		want := "/absolute/path/no/expansion"
		if got != want {
			t.Errorf("ExpandPath() = %q, want %q", got, want)
		}
	})

	t.Run("handles empty path", func(t *testing.T) {
		got, err := ExpandPath("")
		if err != nil {
			t.Fatalf("ExpandPath() error = %v", err)
		}
		if got != "" {
			t.Errorf("ExpandPath(\"\") = %q, want \"\"", got)
		}
	})
}
