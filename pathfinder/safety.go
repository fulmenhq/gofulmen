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
	ErrEscapesRoot   = errors.New("path escapes root directory")
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

// ValidatePathWithinRoot ensures a path doesn't escape the given root directory
// This is critical for security - prevents path traversal attacks via glob patterns
func ValidatePathWithinRoot(absPath, absRoot string) error {
	// Both paths must be absolute for reliable comparison
	if !filepath.IsAbs(absPath) || !filepath.IsAbs(absRoot) {
		return ErrInvalidPath
	}

	// Get relative path from root to target
	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return err
	}

	// If relative path starts with "..", it's outside the root
	// This catches cases like: /repo/../etc/passwd -> ../etc/passwd
	if strings.HasPrefix(relPath, "..") {
		return ErrEscapesRoot
	}

	// Additional check: ensure the path doesn't contain .. anywhere
	// This catches cases like: foo/../../../etc/passwd
	if strings.Contains(relPath, "..") {
		return ErrPathTraversal
	}

	return nil
}

// ContainsHiddenSegment checks if any component of a path starts with a dot
// This is used to filter hidden files/directories when IncludeHidden is false
func ContainsHiddenSegment(path string) bool {
	// Normalize path separators for cross-platform compatibility
	normalized := filepath.ToSlash(path)

	// Split into segments and check each one
	segments := strings.Split(normalized, "/")
	for _, segment := range segments {
		if segment != "" && strings.HasPrefix(segment, ".") {
			return true
		}
	}

	return false
}
