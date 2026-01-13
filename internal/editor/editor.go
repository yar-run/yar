// Package editor provides utilities for opening files in the user's preferred editor.
package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// DetectEditor returns the editor command to use.
// Priority: $EDITOR -> $VISUAL -> platform default (vim on Unix, notepad on Windows)
func DetectEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	// Platform default
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vim"
}

// OpenInEditor opens the file at path in the user's preferred editor.
// Blocks until the editor exits. Inherits stdin/stdout/stderr for interactive editing.
func OpenInEditor(path string) error {
	editor := DetectEditor()

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	return nil
}
