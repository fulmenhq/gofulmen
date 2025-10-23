package pathfinder_test

import (
	"context"
	"fmt"

	"github.com/fulmenhq/gofulmen/pathfinder"
)

// ExampleValidatePath demonstrates path validation
func ExampleValidatePath() {
	// Valid paths
	err := pathfinder.ValidatePath("src/main.go")
	fmt.Println("Valid path:", err == nil)

	// Invalid path (traversal attempt)
	err = pathfinder.ValidatePath("../etc/passwd")
	fmt.Println("Traversal attempt:", err != nil)

	// Output:
	// Valid path: true
	// Traversal attempt: true
}

// ExampleFinder_FindGoFiles demonstrates finding Go source files
func ExampleFinder_FindGoFiles() {
	ctx := context.Background()
	finder := pathfinder.NewFinder()

	// Find all .go files recursively
	results, err := finder.FindGoFiles(ctx, "testdata/basic")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Found %d Go file(s)\n", len(results))
	// Output:
	// Found 1 Go file(s)
}

// ExampleFinder_FindFiles demonstrates file discovery with patterns
func ExampleFinder_FindFiles() {
	ctx := context.Background()
	finder := pathfinder.NewFinder()

	// Find files matching multiple patterns
	query := pathfinder.FindQuery{
		Root:    "testdata/basic",
		Include: []string{"*.go", "*.md"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Found %d file(s)\n", len(results))
	// Output:
	// Found 2 file(s)
}

// ExampleFinder_FindFiles_withExclude demonstrates excluding files
func ExampleFinder_FindFiles_withExclude() {
	ctx := context.Background()
	finder := pathfinder.NewFinder()

	// Find Go files but exclude test files
	query := pathfinder.FindQuery{
		Root:    "testdata/mixed",
		Include: []string{"**/*.go"},
		Exclude: []string{"**/*_test.go"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, result := range results {
		fmt.Println("Found:", result.RelativePath)
	}
	// Output:
	// Found: src/main.go
}

// ExampleFinder_FindFiles_maxDepth demonstrates depth limiting
func ExampleFinder_FindFiles_maxDepth() {
	ctx := context.Background()
	finder := pathfinder.NewFinder()

	// Find files only in top-level directory
	query := pathfinder.FindQuery{
		Root:     "testdata/nested",
		Include:  []string{"**/*.go"},
		MaxDepth: 1,
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Found %d file(s) at depth 1\n", len(results))
	// Output:
	// Found 1 file(s) at depth 1
}

// ExampleFinder_FindFiles_withChecksums demonstrates checksum calculation
func ExampleFinder_FindFiles_withChecksums() {
	ctx := context.Background()
	finder := pathfinder.NewFinder()

	// Find files with checksum calculation
	query := pathfinder.FindQuery{
		Root:               "testdata/basic",
		Include:            []string{"*.go"},
		CalculateChecksums: true,
		ChecksumAlgorithm:  "xxh3-128",
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(results) > 0 {
		result := results[0]
		checksum := result.Metadata["checksum"]
		algorithm := result.Metadata["checksumAlgorithm"]
		fmt.Printf("File: %s\n", result.RelativePath)
		fmt.Printf("Checksum: %s\n", checksum)
		fmt.Printf("Algorithm: %s\n", algorithm)
	}
	// Output:
	// File: file1.go
	// Checksum: xxh3-128:8754591d28adc9bc021db81a6d87be18
	// Algorithm: xxh3-128
}
