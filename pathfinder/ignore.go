package pathfinder

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// IgnoreMatcher handles .fulmenignore pattern matching
type IgnoreMatcher struct {
	patterns []string
	root     string
}

// NewIgnoreMatcher creates a new ignore matcher for the given root directory
func NewIgnoreMatcher(root string) (*IgnoreMatcher, error) {
	matcher := &IgnoreMatcher{
		root:     root,
		patterns: make([]string, 0),
	}

	// Load .fulmenignore if it exists
	ignoreFile := filepath.Join(root, ".fulmenignore")
	if _, err := os.Stat(ignoreFile); err == nil {
		if err := matcher.loadIgnoreFile(ignoreFile); err != nil {
			return nil, err
		}
	}

	return matcher, nil
}

// loadIgnoreFile reads and parses a .fulmenignore file
func (m *IgnoreMatcher) loadIgnoreFile(path string) error {
	// #nosec G304 -- path is constructed from validated root via filepath.Join in NewIgnoreMatcher
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log but don't fail on close error
			_ = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Add pattern to list
		m.patterns = append(m.patterns, line)
	}

	return scanner.Err()
}

// IsIgnored checks if a relative path should be ignored based on patterns
func (m *IgnoreMatcher) IsIgnored(relPath string) bool {
	// Normalize path separators for cross-platform compatibility
	normalizedPath := filepath.ToSlash(relPath)

	for _, pattern := range m.patterns {
		// Normalize pattern separators
		normalizedPattern := filepath.ToSlash(pattern)

		// Handle directory patterns (ending with /)
		if strings.HasSuffix(normalizedPattern, "/") {
			// Directory pattern - match the directory and everything under it
			dirPattern := strings.TrimSuffix(normalizedPattern, "/")
			if strings.HasPrefix(normalizedPath, dirPattern+"/") || normalizedPath == dirPattern {
				return true
			}
		}

		// Try exact match with doublestar for glob support
		matched, err := doublestar.Match(normalizedPattern, normalizedPath)
		if err == nil && matched {
			return true
		}

		// Gitignore semantics: patterns without / match files in any directory
		// e.g., "*.log" should match "src/debug.log"
		if !strings.Contains(normalizedPattern, "/") {
			// Match just the filename
			filename := filepath.Base(normalizedPath)
			matched, err := doublestar.Match(normalizedPattern, filename)
			if err == nil && matched {
				return true
			}
		}

		// Also try matching with pattern as prefix (for directory-style patterns)
		if strings.HasPrefix(normalizedPath, normalizedPattern+"/") {
			return true
		}
	}

	return false
}

// AddPattern adds a custom ignore pattern
func (m *IgnoreMatcher) AddPattern(pattern string) {
	m.patterns = append(m.patterns, pattern)
}

// GetPatterns returns all loaded patterns
func (m *IgnoreMatcher) GetPatterns() []string {
	return append([]string{}, m.patterns...)
}
