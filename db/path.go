package db

import (
	"os"
	"path/filepath"
	"strings"
)

// GetPointerFilePath returns the path to the global file that stores the workspace location.
func GetPointerFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gotrack_workspace"), nil
}

// GetWorkspacePath reads the global pointer file to find where the DB and Config should live.
// Returns an empty string if it hasn't been set up yet.
func GetWorkspacePath() (string, error) {
	ptr, err := GetPointerFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(ptr)
	if os.IsNotExist(err) {
		return "", nil // Not setup yet
	} else if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// SetWorkspacePath creates the workspace directory and saves its location to the global pointer file.
func SetWorkspacePath(path string) error {
	// Expand ~ if necessary
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	path = filepath.Clean(path)

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	ptr, err := GetPointerFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(ptr, []byte(path), 0644)
}
