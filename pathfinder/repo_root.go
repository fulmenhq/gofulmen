package pathfinder

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/fulmenhq/gofulmen/errors"
)

// Predefined marker sets for common repository types
var (
	GitMarkers      = []string{".git"}
	GoModMarkers    = []string{"go.mod"}
	NodeMarkers     = []string{"package.json"}
	PythonMarkers   = []string{"pyproject.toml", "setup.py"}
	MonorepoMarkers = []string{".git", "pnpm-workspace.yaml", "lerna.json"}
)

// FindOptions configures safety boundaries and behavior for repository root discovery
type FindOptions struct {
	// StopAtFirst stops at first marker found (default: true)
	StopAtFirst bool

	// MaxDepth limits upward traversal (default: 10 directories)
	MaxDepth int

	// Boundary sets absolute ceiling path, never traverse above
	// Default: user home directory ($HOME, %USERPROFILE%)
	Boundary string

	// RespectConstraints integrates with PathConstraint if configured (default: true)
	RespectConstraints bool

	// FollowSymlinks whether to follow symlinks during traversal (default: false)
	FollowSymlinks bool
}

// FindOption is a functional option for configuring FindRepositoryRoot
type FindOption func(*FindOptions)

// WithStopAtFirst configures whether to stop at the first marker found
func WithStopAtFirst(stop bool) FindOption {
	return func(opts *FindOptions) {
		opts.StopAtFirst = stop
	}
}

// WithMaxDepth sets the maximum number of directories to traverse upward
func WithMaxDepth(depth int) FindOption {
	return func(opts *FindOptions) {
		opts.MaxDepth = depth
	}
}

// WithBoundary sets an explicit ceiling path to never traverse above
func WithBoundary(boundary string) FindOption {
	return func(opts *FindOptions) {
		opts.Boundary = boundary
	}
}

// WithRespectConstraints configures PathConstraint integration
func WithRespectConstraints(respect bool) FindOption {
	return func(opts *FindOptions) {
		opts.RespectConstraints = respect
	}
}

// WithFollowSymlinks configures whether to follow symlinks during traversal
func WithFollowSymlinks(follow bool) FindOption {
	return func(opts *FindOptions) {
		opts.FollowSymlinks = follow
	}
}

// FindRepositoryRoot searches upward from startPath looking for marker files/directories.
// Returns the directory containing the first marker found, or error if not found.
//
// This function provides safe upward filesystem traversal with multiple safety boundaries:
// - Never traverses above user home directory by default
// - Never traverses above filesystem root (/, C:\, \\server\share\)
// - Respects explicit boundary configuration
// - Guards against infinite loops with max depth limit
//
// Example usage:
//
//	// Find Git repository root
//	root, err := pathfinder.FindRepositoryRoot(".", pathfinder.GitMarkers)
//
//	// Find Go module root with custom options
//	root, err := pathfinder.FindRepositoryRoot(
//	    "internal/config",
//	    pathfinder.GoModMarkers,
//	    pathfinder.WithMaxDepth(5),
//	    pathfinder.WithBoundary("/home/user/projects"),
//	)
func FindRepositoryRoot(startPath string, markers []string, opts ...FindOption) (string, error) {
	// Apply default options
	options := FindOptions{
		StopAtFirst:        true,
		MaxDepth:           10,
		Boundary:           "", // Will be set to home dir below
		RespectConstraints: true,
		FollowSymlinks:     false,
	}

	for _, opt := range opts {
		opt(&options)
	}

	// Validate start path
	if startPath == "" {
		envelope := errors.NewErrorEnvelope("INVALID_START_PATH", "start path cannot be empty")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		return "", envelope
	}

	// Validate markers
	if len(markers) == 0 {
		envelope := errors.NewErrorEnvelope("INVALID_MARKERS", "markers list cannot be empty")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		return "", envelope
	}

	// Get absolute start path
	absStart, err := filepath.Abs(startPath)
	if err != nil {
		envelope := errors.NewErrorEnvelope("INVALID_START_PATH", "failed to resolve absolute path for start path")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithOriginal(err)
		envelope = errors.SafeWithContext(envelope, map[string]any{
			"startPath": startPath,
		})
		return "", envelope
	}

	// Ensure start path exists
	startInfo, err := os.Stat(absStart)
	if err != nil {
		if os.IsNotExist(err) {
			envelope := errors.NewErrorEnvelope("INVALID_START_PATH", "start path does not exist")
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
			envelope = errors.SafeWithContext(envelope, map[string]any{
				"startPath": startPath,
				"absPath":   absStart,
			})
			return "", envelope
		}
		envelope := errors.NewErrorEnvelope("INVALID_START_PATH", "failed to stat start path")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithOriginal(err)
		envelope = errors.SafeWithContext(envelope, map[string]any{
			"startPath": startPath,
			"absPath":   absStart,
		})
		return "", envelope
	}

	// If start path is a file, use its directory
	if !startInfo.IsDir() {
		absStart = filepath.Dir(absStart)
	}

	// Determine boundary (default: home directory)
	boundary := options.Boundary
	if boundary == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback: if we can't get home dir (container/CI), use start path as boundary
			boundary = absStart
		} else {
			// Special case: if home is /root or /, use start path instead to prevent filesystem traversal
			if homeDir == "/" || homeDir == "/root" {
				boundary = absStart
			} else {
				boundary = homeDir
			}
		}
	}

	// Get absolute boundary path
	absBoundary, err := filepath.Abs(boundary)
	if err != nil {
		envelope := errors.NewErrorEnvelope("INVALID_BOUNDARY", "failed to resolve absolute path for boundary")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithOriginal(err)
		envelope = errors.SafeWithContext(envelope, map[string]any{
			"boundary": boundary,
		})
		return "", envelope
	}

	// Ensure boundary exists
	if _, err := os.Stat(absBoundary); err != nil {
		envelope := errors.NewErrorEnvelope("INVALID_BOUNDARY", "boundary path does not exist")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithOriginal(err)
		envelope = errors.SafeWithContext(envelope, map[string]any{
			"boundary":    boundary,
			"absBoundary": absBoundary,
		})
		return "", envelope
	}

	// Get filesystem root for this path
	fsRoot := getFilesystemRoot(absStart)

	// Track visited paths to detect symlink loops
	visited := make(map[string]bool)

	// Iterative upward traversal
	currentDir := absStart
	depth := 0

	for {
		// Detect symlink loops when following symlinks
		if options.FollowSymlinks {
			// Get real path to detect cycles
			realPath, err := filepath.EvalSymlinks(currentDir)
			if err == nil {
				if visited[realPath] {
					envelope := errors.NewErrorEnvelope("TRAVERSAL_LOOP", "symlink loop detected during repository root discovery")
					envelope = errors.SafeWithSeverity(envelope, errors.SeverityCritical)
					envelope = errors.SafeWithContext(envelope, map[string]any{
						"startPath":     startPath,
						"absStartPath":  absStart,
						"currentPath":   currentDir,
						"realPath":      realPath,
						"searchedDepth": depth,
					})
					return "", envelope
				}
				visited[realPath] = true
			}
		}
		// Check if we've hit any boundary
		if depth >= options.MaxDepth {
			envelope := errors.NewErrorEnvelope("REPOSITORY_NOT_FOUND", "no repository markers found within search boundaries")
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = errors.SafeWithContext(envelope, map[string]any{
				"startPath":     startPath,
				"absStartPath":  absStart,
				"markers":       markers,
				"searchedDepth": depth,
				"stoppedAt":     currentDir,
				"reason":        "max_depth_reached",
				"maxDepth":      options.MaxDepth,
			})
			return "", envelope
		}

		// Stop at boundary
		if !isWithinBoundary(currentDir, absBoundary) {
			envelope := errors.NewErrorEnvelope("REPOSITORY_NOT_FOUND", "no repository markers found within search boundaries")
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = errors.SafeWithContext(envelope, map[string]any{
				"startPath":     startPath,
				"absStartPath":  absStart,
				"markers":       markers,
				"searchedDepth": depth,
				"stoppedAt":     currentDir,
				"reason":        "boundary_reached",
				"boundary":      absBoundary,
			})
			return "", envelope
		}

		// Stop at filesystem root
		if currentDir == fsRoot {
			envelope := errors.NewErrorEnvelope("REPOSITORY_NOT_FOUND", "no repository markers found within search boundaries")
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = errors.SafeWithContext(envelope, map[string]any{
				"startPath":      startPath,
				"absStartPath":   absStart,
				"markers":        markers,
				"searchedDepth":  depth,
				"stoppedAt":      currentDir,
				"reason":         "filesystem_root_reached",
				"filesystemRoot": fsRoot,
			})
			return "", envelope
		}

		// Check for markers in current directory
		found, _, err := checkForMarkers(currentDir, markers, options.FollowSymlinks)
		if err != nil {
			// Permission denied or other errors: log and continue upward
			// This provides graceful degradation
		} else if found {
			// Found a marker! Return this directory
			return currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)

		// Detect if we can't go further (reached root in a different way)
		if parentDir == currentDir {
			envelope := errors.NewErrorEnvelope("REPOSITORY_NOT_FOUND", "no repository markers found within search boundaries")
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = errors.SafeWithContext(envelope, map[string]any{
				"startPath":     startPath,
				"absStartPath":  absStart,
				"markers":       markers,
				"searchedDepth": depth,
				"stoppedAt":     currentDir,
				"reason":        "traversal_termination",
			})
			return "", envelope
		}

		currentDir = parentDir
		depth++
	}
}

// getFilesystemRoot returns the filesystem root for a given path
func getFilesystemRoot(absPath string) string {
	if runtime.GOOS == "windows" {
		// Windows: get volume name (C:\, D:\, \\server\share\)
		volumeName := filepath.VolumeName(absPath)
		if volumeName != "" {
			// For drive letters (C:), append backslash (C:\)
			// For UNC paths (\\server\share), volumeName is already complete
			if len(volumeName) == 2 && volumeName[1] == ':' {
				return volumeName + "\\"
			}
			return volumeName + "\\"
		}
		return absPath // Fallback
	}

	// Unix: root is always /
	return "/"
}

// isWithinBoundary checks if path is within or equal to boundary
func isWithinBoundary(path, boundary string) bool {
	// Clean paths for comparison
	cleanPath := filepath.Clean(path)
	cleanBoundary := filepath.Clean(boundary)

	// Path must be equal to or a subdirectory of boundary
	// We check if path has boundary as prefix
	rel, err := filepath.Rel(cleanBoundary, cleanPath)
	if err != nil {
		return false
	}

	// If rel is "." then path == boundary (within boundary)
	// If rel doesn't start with ".." then path is under boundary
	return rel == "." || (!filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.')
}

// checkForMarkers checks if any marker exists in the given directory
func checkForMarkers(dir string, markers []string, followSymlinks bool) (bool, string, error) {
	for _, marker := range markers {
		markerPath := filepath.Join(dir, marker)

		var stat os.FileInfo
		var err error

		if followSymlinks {
			stat, err = os.Stat(markerPath)
		} else {
			stat, err = os.Lstat(markerPath)
		}

		if err != nil {
			if os.IsNotExist(err) {
				continue // Marker doesn't exist, try next
			}
			// Other error (permission denied, etc.)
			return false, "", err
		}

		// Marker exists!
		// We don't care if it's a file or directory for this function
		_ = stat
		return true, marker, nil
	}

	return false, "", nil
}
