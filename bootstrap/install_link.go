package bootstrap

import (
	"fmt"
	"io"
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

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destPath := filepath.Join(destDir, tool.Install.BinName)

	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing binary: %w", err)
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	return nil
}
