package export

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrFileExists is returned when trying to write to an existing file without Overwrite
	ErrFileExists = errors.New("file already exists (use Overwrite option to replace)")

	// ErrPathValidation is returned when the output path fails validation
	ErrPathValidation = errors.New("output path validation failed")

	// ErrSchemaNotFound is returned when the requested schema cannot be loaded
	ErrSchemaNotFound = errors.New("schema not found")

	// ErrSchemaValidation is returned when schema validation fails
	ErrSchemaValidation = errors.New("schema validation failed")

	// ErrFileWrite is returned when writing the output file fails
	ErrFileWrite = errors.New("failed to write file")
)

// validateOutputPath validates and prepares the output path
func validateOutputPath(outPath string, overwrite bool) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("%w: failed to resolve absolute path: %v", ErrPathValidation, err)
	}

	// Check if file already exists
	if _, err := os.Stat(absPath); err == nil {
		// File exists
		if !overwrite {
			return fmt.Errorf("%w: %s", ErrFileExists, absPath)
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred (permission, etc.)
		return fmt.Errorf("%w: failed to stat path: %v", ErrPathValidation, err)
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("%w: failed to create parent directory: %v", ErrPathValidation, err)
	}

	return nil
}

// writeFileSafe writes data to a file with safety checks
func writeFileSafe(path string, data []byte, overwrite bool) error {
	// Validate path
	if err := validateOutputPath(path, overwrite); err != nil {
		return err
	}

	// Write file with appropriate permissions (0644 = rw-r--r--)
	// #nosec G306 -- schema exports are intended to be readable by other users
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
