package pathfinder

import (
	"errors"
	"path/filepath"
	"strings"
)

// Common safety errors
var (
	ErrPathTraversal = errors.New("path traversal detected")
	ErrInvalidPath   = errors.New("invalid path")
)

// ValidatePath checks if a path is safe to access
func ValidatePath(path string) error {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return ErrPathTraversal
	}

	// Check for empty path
	if cleanPath == "" || cleanPath == "." {
		return ErrInvalidPath
	}

	// Check for root path (too broad, but safe)
	if cleanPath == "/" || cleanPath == "\\" {
		return ErrInvalidPath
	}

	return nil
}

// IsSafePath checks if a path is safe without returning an error
func IsSafePath(path string) bool {
	return ValidatePath(path) == nil
}
