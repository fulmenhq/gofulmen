package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
)

func installLink(tool *Tool) error {
	source := tool.Install.Source

	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file not found: %s", source)
		}
		return fmt.Errorf("failed to access source: %w", err)
	}

	destDir := tool.Install.Destination
	if destDir == "" {
		destDir = "./bin"
	}

	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destPath := filepath.Join(destDir, tool.Install.BinName)

	// Remove existing file or symlink
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing binary: %w", err)
	}

	// Convert source to absolute path for symlink
	absSource, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}

	// Create symlink instead of copying
	if err := os.Symlink(absSource, destPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}
