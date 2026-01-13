package editor

import (
	"os"
	"runtime"
	"testing"
)

func TestDetectEditorWithEDITOR(t *testing.T) {
	// Save and restore original env
	orig := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", orig)

	os.Setenv("EDITOR", "nano")

	got := DetectEditor()
	if got != "nano" {
		t.Errorf("DetectEditor() = %q, want %q", got, "nano")
	}
}

func TestDetectEditorWithVISUAL(t *testing.T) {
	// Save and restore original env
	origEditor := os.Getenv("EDITOR")
	origVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", origEditor)
		os.Setenv("VISUAL", origVisual)
	}()

	os.Unsetenv("EDITOR")
	os.Setenv("VISUAL", "code")

	got := DetectEditor()
	if got != "code" {
		t.Errorf("DetectEditor() = %q, want %q", got, "code")
	}
}

func TestDetectEditorPlatformDefault(t *testing.T) {
	// Save and restore original env
	origEditor := os.Getenv("EDITOR")
	origVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", origEditor)
		os.Setenv("VISUAL", origVisual)
	}()

	os.Unsetenv("EDITOR")
	os.Unsetenv("VISUAL")

	got := DetectEditor()

	var want string
	switch runtime.GOOS {
	case "windows":
		want = "notepad"
	default:
		want = "vim"
	}

	if got != want {
		t.Errorf("DetectEditor() = %q, want %q", got, want)
	}
}
