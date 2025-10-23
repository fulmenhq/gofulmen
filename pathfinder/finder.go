package pathfinder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fulmenhq/crucible"
	"github.com/fulmenhq/gofulmen/fulhash"
	"github.com/fulmenhq/gofulmen/schema"
)

// FinderConfig holds default settings for the FinderFacade
type FinderConfig struct {
	// TODO: Future enhancement - implement concurrent file discovery
	MaxWorkers int `json:"maxWorkers"` // Currently unused - single-threaded implementation

	// TODO: Future enhancement - implement result caching
	CacheEnabled bool `json:"cacheEnabled"` // Currently unused - no caching layer
	CacheTTL     int  `json:"cacheTTL"`     // Currently unused - cache TTL in seconds

	// TODO: Future enhancement - implement PathConstraint enforcement
	Constraint PathConstraint `json:"constraint"` // Currently unused - no constraint validation

	// Implemented fields
	LoaderType      string `json:"loaderType"`      // Type of loader (default: "local")
	ValidateInputs  bool   `json:"validateInputs"`  // Validate FindQuery inputs against schema
	ValidateOutputs bool   `json:"validateOutputs"` // Validate PathResult outputs against schema
}

// FindQuery specifies the parameters for discovery
type FindQuery struct {
	Root               string                                             `json:"root"`
	Include            []string                                           `json:"include"`
	Exclude            []string                                           `json:"exclude,omitempty"`
	MaxDepth           int                                                `json:"maxDepth,omitempty"`
	FollowSymlinks     bool                                               `json:"followSymlinks,omitempty"`
	IncludeHidden      bool                                               `json:"includeHidden,omitempty"`
	CalculateChecksums bool                                               `json:"calculateChecksums,omitempty"`
	ChecksumAlgorithm  string                                             `json:"checksumAlgorithm,omitempty"`
	ErrorHandler       func(path string, err error) error                 `json:"-"`
	ProgressCallback   func(processed int, total int, currentPath string) `json:"-"`
}

// PathResult represents a discovered path along with logical mapping information
type PathResult struct {
	RelativePath string         `json:"relativePath"`
	SourcePath   string         `json:"sourcePath"`
	LogicalPath  string         `json:"logicalPath"`
	LoaderType   string         `json:"loaderType"`
	Metadata     map[string]any `json:"metadata"`
}

// Finder provides high-level path discovery operations
type Finder struct {
	config FinderConfig
}

// NewFinder creates a new finder with default config
func NewFinder() *Finder {
	return &Finder{
		config: FinderConfig{
			MaxWorkers:      4,
			CacheEnabled:    false,
			LoaderType:      "local",
			ValidateInputs:  false, // disabled by default for performance
			ValidateOutputs: false, // disabled by default for performance
		},
	}
}

// FindFiles performs file discovery based on the query
func (f *Finder) FindFiles(ctx context.Context, query FindQuery) ([]PathResult, error) {
	// Validate input if enabled
	if f.config.ValidateInputs {
		if err := ValidateFindQuery(query); err != nil {
			return nil, fmt.Errorf("input validation failed: %w", err)
		}
	}

	// Convert root to absolute path for relative path calculations
	absRoot, err := filepath.Abs(query.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute root path: %w", err)
	}

	// Load .fulmenignore patterns from root directory
	ignoreMatcher, err := NewIgnoreMatcher(absRoot)
	if err != nil {
		// Non-fatal - continue without ignore patterns
		if query.ErrorHandler != nil {
			// Error handler call failure is non-critical in pathfinder context
			_ = query.ErrorHandler(".fulmenignore", err)
		}
	}

	var results []PathResult

	// Collect all matches from include patterns
	for _, pattern := range query.Include {
		// Use doublestar for recursive ** support - always use absolute root
		globPattern := filepath.Join(absRoot, pattern)

		// SECURITY: Validate the glob pattern base doesn't escape root
		// Extract the base directory (part before any wildcard characters)
		basePattern := globPattern
		for _, wildcard := range []string{"*", "?", "[", "]"} {
			if idx := strings.Index(basePattern, wildcard); idx != -1 {
				basePattern = basePattern[:idx]
			}
		}
		// Clean the base pattern
		basePattern = filepath.Clean(basePattern)

		// Ensure the base pattern is within or starts at absRoot
		// This prevents patterns like ../../**/*.go from escaping
		if basePattern != absRoot && !strings.HasPrefix(basePattern, absRoot+string(filepath.Separator)) {
			// Pattern base escapes root - reject it
			if query.ErrorHandler != nil {
				// Error handler call failure is non-critical in pathfinder context
				_ = query.ErrorHandler(pattern, ErrEscapesRoot)
			}
			continue
		}

		matches, err := doublestar.FilepathGlob(globPattern)
		if err != nil {
			if query.ErrorHandler != nil {
				if handlerErr := query.ErrorHandler(pattern, err); handlerErr != nil {
					return nil, handlerErr
				}
			}
			continue
		}

		for _, match := range matches {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			// Convert to absolute path
			absMatch, err := filepath.Abs(match)
			if err != nil {
				continue
			}

			// Validate path safety
			if err := ValidatePath(absMatch); err != nil {
				if query.ErrorHandler != nil {
					// Error handler call failure is non-critical in pathfinder context
					_ = query.ErrorHandler(absMatch, err)
				}
				continue
			}

			// SECURITY: Ensure the matched path doesn't escape the root directory
			// This prevents path traversal attacks via glob patterns like ../**/*.go
			if err := ValidatePathWithinRoot(absMatch, absRoot); err != nil {
				if query.ErrorHandler != nil {
					// Error handler call failure is non-critical in pathfinder context
					_ = query.ErrorHandler(absMatch, err)
				}
				continue
			}

			// Get file info
			info, err := os.Lstat(absMatch)
			if err != nil {
				if query.ErrorHandler != nil {
					// Error handler call failure is non-critical in pathfinder context
					_ = query.ErrorHandler(absMatch, err)
				}
				continue
			}

			// Skip directories (glob returns both files and dirs)
			if info.IsDir() {
				continue
			}

			// Handle symlinks
			if !query.FollowSymlinks && info.Mode()&os.ModeSymlink != 0 {
				continue
			}

			// Get relative path
			relPath, err := filepath.Rel(absRoot, absMatch)
			if err != nil {
				continue
			}

			// Check MaxDepth
			if query.MaxDepth > 0 {
				depth := strings.Count(relPath, string(filepath.Separator)) + 1
				if depth > query.MaxDepth {
					continue
				}
			}

			// Check hidden files/directories - check ALL path segments, not just the base
			// This correctly filters files under hidden directories like .secrets/key.pem
			if !query.IncludeHidden && ContainsHiddenSegment(relPath) {
				continue
			}

			// Check .fulmenignore patterns if matcher is loaded
			if ignoreMatcher != nil && ignoreMatcher.IsIgnored(relPath) {
				continue
			}

			// Populate metadata per Pathfinder spec (size, mtime, checksum)
			metadata := make(map[string]any)
			metadata["size"] = info.Size()
			metadata["mtime"] = info.ModTime().Format("2006-01-02T15:04:05.000000000Z07:00") // RFC3339Nano

			// Optional checksum calculation using FulHash
			if query.CalculateChecksums {
				algorithm := query.ChecksumAlgorithm
				if algorithm == "" {
					algorithm = "xxh3-128" // default
				}

				var alg fulhash.Algorithm
				switch algorithm {
				case "xxh3-128":
					alg = fulhash.XXH3_128
				case "sha256":
					alg = fulhash.SHA256
				default:
					// This should be caught by validation, but handle gracefully
					metadata["checksumError"] = fmt.Sprintf("unsupported algorithm: %s", algorithm)
				}

				if metadata["checksumError"] == nil {
					file, err := os.Open(absMatch) // #nosec G304 -- absMatch is validated with ValidatePathWithinRoot to prevent path traversal
					if err != nil {
						metadata["checksumError"] = fmt.Sprintf("failed to open file: %v", err)
					} else {
						digest, err := fulhash.HashReader(file, fulhash.WithAlgorithm(alg))
						if err != nil {
							metadata["checksumError"] = fmt.Sprintf("checksum calculation failed: %v", err)
						} else {
							metadata["checksum"] = digest.String()
							metadata["checksumAlgorithm"] = string(digest.Algorithm())
						}
						_ = file.Close()
					}
				}
			}

			result := PathResult{
				RelativePath: relPath,
				SourcePath:   absMatch,
				LogicalPath:  relPath,
				LoaderType:   f.config.LoaderType,
				Metadata:     metadata,
			}

			results = append(results, result)

			// Progress callback
			if query.ProgressCallback != nil {
				query.ProgressCallback(len(results), -1, absMatch) // -1 for unknown total
			}
		}
	}

	// Filter by exclude patterns
	if len(query.Exclude) > 0 {
		filtered := make([]PathResult, 0, len(results))
		for _, result := range results {
			excluded := false
			for _, excludePattern := range query.Exclude {
				matched, _ := doublestar.Match(excludePattern, result.RelativePath)
				if matched {
					excluded = true
					break
				}
			}
			if !excluded {
				filtered = append(filtered, result)
			}
		}
		results = filtered
	}

	// Validate outputs if enabled
	if f.config.ValidateOutputs {
		if err := ValidatePathResults(results); err != nil {
			return nil, fmt.Errorf("output validation failed: %w", err)
		}
	}

	return results, nil
}

// FindGoFiles finds Go source files
func (f *Finder) FindGoFiles(ctx context.Context, root string) ([]PathResult, error) {
	query := FindQuery{
		Root:    root,
		Include: []string{"**/*.go"},
	}
	return f.FindFiles(ctx, query)
}

// FindConfigFiles finds common configuration files
func (f *Finder) FindConfigFiles(ctx context.Context, root string) ([]PathResult, error) {
	query := FindQuery{
		Root:    root,
		Include: []string{"**/*.json", "**/*.yaml", "**/*.yml", "**/*.toml", "**/*.config", "**/*.conf"},
	}
	return f.FindFiles(ctx, query)
}

// FindSchemaFiles finds JSON Schema files
func (f *Finder) FindSchemaFiles(ctx context.Context, root string) ([]PathResult, error) {
	query := FindQuery{
		Root:    root,
		Include: []string{"**/*.schema.json", "**/schema.json"},
	}
	return f.FindFiles(ctx, query)
}

// FindByExtension finds files with specific extensions
func (f *Finder) FindByExtension(ctx context.Context, root string, exts []string) ([]PathResult, error) {
	patterns := make([]string, len(exts))
	for i, ext := range exts {
		patterns[i] = "**/*." + ext
	}

	query := FindQuery{
		Root:    root,
		Include: patterns,
	}
	return f.FindFiles(ctx, query)
}

// FindGoFilesWithChecksums finds Go source files with optional checksum calculation
func (f *Finder) FindGoFilesWithChecksums(ctx context.Context, root string, calculateChecksums bool, algorithm string) ([]PathResult, error) {
	query := FindQuery{
		Root:               root,
		Include:            []string{"**/*.go"},
		CalculateChecksums: calculateChecksums,
		ChecksumAlgorithm:  algorithm,
	}
	return f.FindFiles(ctx, query)
}

// FindConfigFilesWithChecksums finds common configuration files with optional checksum calculation
func (f *Finder) FindConfigFilesWithChecksums(ctx context.Context, root string, calculateChecksums bool, algorithm string) ([]PathResult, error) {
	query := FindQuery{
		Root:               root,
		Include:            []string{"**/*.json", "**/*.yaml", "**/*.yml", "**/*.toml", "**/*.config", "**/*.conf"},
		CalculateChecksums: calculateChecksums,
		ChecksumAlgorithm:  algorithm,
	}
	return f.FindFiles(ctx, query)
}

// ValidateFindQuery validates a FindQuery against the schema
func ValidateFindQuery(query FindQuery) error {
	// Validate checksum algorithm if checksums are enabled
	if query.CalculateChecksums {
		switch query.ChecksumAlgorithm {
		case "", "xxh3-128", "sha256":
			// Valid algorithms
		default:
			return fmt.Errorf("invalid checksumAlgorithm %q: must be one of 'xxh3-128', 'sha256', or empty (defaults to 'xxh3-128')", query.ChecksumAlgorithm)
		}
	}

	pathfinderSchemas, err := crucible.SchemaRegistry.Pathfinder().V1_0_0()
	if err != nil {
		return fmt.Errorf("failed to get pathfinder schemas: %w", err)
	}

	schemaData, err := pathfinderSchemas.FindQuery()
	if err != nil {
		return fmt.Errorf("failed to load find-query schema: %w", err)
	}

	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	diags, err := validator.ValidateData(query)
	if err != nil {
		return fmt.Errorf("failed to validate query: %w", err)
	}
	if verrs := schema.DiagnosticsToValidationErrors(diags); len(verrs) > 0 {
		return verrs
	}
	return nil
}

// ValidatePathResult validates a PathResult against the schema
func ValidatePathResult(result PathResult) error {
	pathfinderSchemas, err := crucible.SchemaRegistry.Pathfinder().V1_0_0()
	if err != nil {
		return fmt.Errorf("failed to get pathfinder schemas: %w", err)
	}

	schemaData, err := pathfinderSchemas.PathResult()
	if err != nil {
		return fmt.Errorf("failed to load path-result schema: %w", err)
	}

	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	diags, err := validator.ValidateData(result)
	if err != nil {
		return fmt.Errorf("failed to validate path result: %w", err)
	}
	if verrs := schema.DiagnosticsToValidationErrors(diags); len(verrs) > 0 {
		return verrs
	}
	return nil
}

// ValidatePathResults validates multiple PathResult objects
func ValidatePathResults(results []PathResult) error {
	for i, result := range results {
		if err := ValidatePathResult(result); err != nil {
			return fmt.Errorf("result %d validation failed: %w", i, err)
		}
	}
	return nil
}
