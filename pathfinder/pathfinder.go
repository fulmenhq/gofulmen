package pathfinder

import (
	"context"
	"io/fs"
)

// Pathfinder provides safe filesystem discovery operations
type Pathfinder interface {
	// Discover finds files matching patterns with safety checks
	Discover(ctx context.Context, root string, patterns []string) ([]string, error)

	// Walk safely traverses a directory tree
	Walk(ctx context.Context, root string, walkFn fs.WalkDirFunc) error

	// ValidatePath checks if a path is safe to access
	ValidatePath(path string) error
}

// DiscoveryResult represents a discovered file or directory
type DiscoveryResult struct {
	Path  string
	Type  fs.FileMode
	Size  int64
	Error error
}

// DiscoveryOptions configures discovery behavior
type DiscoveryOptions struct {
	MaxDepth        int
	FollowLinks     bool
	ExcludePatterns []string
	IncludeHidden   bool
}

// PathConstraint defines boundaries for path operations
type PathConstraint interface {
	Contains(path string) bool
	Root() string
	Type() ConstraintType
	EnforcementLevel() EnforcementLevel
}

// ConstraintType represents different types of path constraints
type ConstraintType string

const (
	ConstraintRepository ConstraintType = "repository"
	ConstraintWorkspace  ConstraintType = "workspace"
	ConstraintCloud      ConstraintType = "cloud"
)

// EnforcementLevel defines how strictly constraints are enforced
type EnforcementLevel string

const (
	EnforcementStrict     EnforcementLevel = "strict"
	EnforcementWarn       EnforcementLevel = "warn"
	EnforcementPermissive EnforcementLevel = "permissive"
)
