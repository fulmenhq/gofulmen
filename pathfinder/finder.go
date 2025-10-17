package pathfinder

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fulmenhq/crucible"
	"github.com/fulmenhq/gofulmen/schema"
)

// FinderConfig holds default settings for the FinderFacade
type FinderConfig struct {
	MaxWorkers      int            `json:"maxWorkers"`
	CacheEnabled    bool           `json:"cacheEnabled"`
	CacheTTL        int            `json:"cacheTTL"` // in seconds
	Constraint      PathConstraint `json:"constraint"`
	LoaderType      string         `json:"loaderType"`
	ValidateInputs  bool           `json:"validateInputs"`  // validate FindQuery inputs
	ValidateOutputs bool           `json:"validateOutputs"` // validate PathResult outputs
}

// FindQuery specifies the parameters for discovery
type FindQuery struct {
	Root             string                                             `json:"root"`
	Include          []string                                           `json:"include"`
	Exclude          []string                                           `json:"exclude,omitempty"`
	MaxDepth         int                                                `json:"maxDepth,omitempty"`
	FollowSymlinks   bool                                               `json:"followSymlinks,omitempty"`
	IncludeHidden    bool                                               `json:"includeHidden,omitempty"`
	ErrorHandler     func(path string, err error) error                 `json:"-"`
	ProgressCallback func(processed int, total int, currentPath string) `json:"-"`
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

	var results []PathResult

	// For each include pattern, find matching files
	for _, pattern := range query.Include {
		matches, err := filepath.Glob(filepath.Join(query.Root, "**", pattern))
		if err != nil {
			if query.ErrorHandler != nil {
				if err := query.ErrorHandler(pattern, err); err != nil {
					return nil, err
				}
			}
			continue
		}

		for _, match := range matches {
			// Validate path safety
			if err := ValidatePath(match); err != nil {
				if query.ErrorHandler != nil {
					// #nosec G104 -- error handler call failure is non-critical in pathfinder context
					query.ErrorHandler(match, err)
				}
				continue
			}

			// Get relative path
			relPath, err := filepath.Rel(query.Root, match)
			if err != nil {
				continue
			}

			result := PathResult{
				RelativePath: relPath,
				SourcePath:   match,
				LogicalPath:  relPath, // Same as relative for now
				LoaderType:   f.config.LoaderType,
				Metadata:     make(map[string]any),
			}

			results = append(results, result)

			// Progress callback
			if query.ProgressCallback != nil {
				query.ProgressCallback(len(results), -1, match) // -1 for unknown total
			}
		}
	}

	// Validate outputs if enabled
	if f.config.ValidateOutputs {
		if err := ValidatePathResults(results); err != nil {
			return nil, fmt.Errorf("output validation failed: %w", err)
		}
	}

	return results, nil
}

// FindConfigFiles finds common configuration files
func (f *Finder) FindConfigFiles(ctx context.Context, root string) ([]PathResult, error) {
	query := FindQuery{
		Root:    root,
		Include: []string{"*.json", "*.yaml", "*.yml", "*.toml", "*.config", "*.conf"},
	}
	return f.FindFiles(ctx, query)
}

// FindSchemaFiles finds JSON Schema files
func (f *Finder) FindSchemaFiles(ctx context.Context, root string) ([]PathResult, error) {
	query := FindQuery{
		Root:    root,
		Include: []string{"*.schema.json", "schema.json"},
	}
	return f.FindFiles(ctx, query)
}

// FindByExtension finds files with specific extensions
func (f *Finder) FindByExtension(ctx context.Context, root string, exts []string) ([]PathResult, error) {
	patterns := make([]string, len(exts))
	for i, ext := range exts {
		patterns[i] = "*." + ext
	}

	query := FindQuery{
		Root:    root,
		Include: patterns,
	}
	return f.FindFiles(ctx, query)
}

// ValidateFindQuery validates a FindQuery against the schema
func ValidateFindQuery(query FindQuery) error {
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

	return validator.Validate(query)
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

	return validator.Validate(result)
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
