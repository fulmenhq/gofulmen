package pathfinder

import (
	goerrors "errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fulmenhq/gofulmen/errors"
)

// Common safety errors (sentinel errors for errors.Is compatibility)
var (
	ErrPathTraversal = goerrors.New("path traversal detected")
	ErrInvalidPath   = goerrors.New("invalid path")
	ErrEscapesRoot   = goerrors.New("path escapes root directory")
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
func ValidatePathWithinRoot(absPath, absRoot string) error {
	return ValidatePathWithinRootWithEnvelope(absPath, absRoot, "")
}

// ValidatePathWithinRootWithEnvelope ensures a path doesn't escape the given root directory with structured error reporting
// This is critical for security - prevents path traversal attacks via glob patterns
func ValidatePathWithinRootWithEnvelope(absPath, absRoot, correlationID string) error {
	// Both paths must be absolute for reliable comparison
	if !filepath.IsAbs(absPath) || !filepath.IsAbs(absRoot) {
		envelope := errors.NewErrorEnvelope("PATHFINDER_VALIDATION_ERROR", "Path validation requires absolute paths")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":        "pathfinder",
			"operation":        "validate_path_within_root",
			"error_type":       "invalid_path",
			"abs_path":         absPath,
			"abs_root":         absRoot,
			"path_is_absolute": filepath.IsAbs(absPath),
			"root_is_absolute": filepath.IsAbs(absRoot),
		})
		envelope = envelope.WithOriginal(ErrInvalidPath)
		return envelope
	}

	// Get relative path from root to target
	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		envelope := errors.NewErrorEnvelope("PATHFINDER_VALIDATION_ERROR", fmt.Sprintf("Failed to compute relative path from %s to %s", absRoot, absPath))
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":  "pathfinder",
			"operation":  "validate_path_within_root",
			"error_type": "path_resolution_error",
			"abs_path":   absPath,
			"abs_root":   absRoot,
		})
		envelope = envelope.WithOriginal(err)
		return envelope
	}

	// If relative path starts with "..", it's outside the root
	// This catches cases like: /repo/../etc/passwd -> ../etc/passwd
	if strings.HasPrefix(relPath, "..") {
		envelope := errors.NewErrorEnvelope("PATHFINDER_SECURITY_ERROR", fmt.Sprintf("Path %s escapes root directory %s", absPath, absRoot))
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityCritical)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":     "pathfinder",
			"operation":     "validate_path_within_root",
			"error_type":    "path_escape",
			"abs_path":      absPath,
			"abs_root":      absRoot,
			"relative_path": relPath,
		})
		envelope = envelope.WithOriginal(ErrEscapesRoot)
		return envelope
	}

	// Additional check: ensure the path doesn't contain .. anywhere
	// This catches cases like: foo/../../../etc/passwd
	if strings.Contains(relPath, "..") {
		envelope := errors.NewErrorEnvelope("PATHFINDER_SECURITY_ERROR", fmt.Sprintf("Path traversal detected in %s", absPath))
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityCritical)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":     "pathfinder",
			"operation":     "validate_path_within_root",
			"error_type":    "path_traversal",
			"abs_path":      absPath,
			"abs_root":      absRoot,
			"relative_path": relPath,
		})
		envelope = envelope.WithOriginal(ErrPathTraversal)
		return envelope
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
