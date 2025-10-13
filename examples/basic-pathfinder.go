package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
	// Create a new finder with default configuration
	finder := pathfinder.NewFinder()

	// Find all Go files in the current directory
	ctx := context.Background()
	results, err := finder.FindByExtension(ctx, ".", []string{"go"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d Go files:\n", len(results))
	for _, result := range results {
		fmt.Printf("- %s\n", result.SourcePath)
	}

	// Find files by extension
	results, err = finder.FindByExtension(ctx, ".", []string{"md", "json"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d documentation/config files:\n", len(results))
	for _, result := range results {
		fmt.Printf("- %s\n", result.SourcePath)
	}

	// Custom query example
	query := pathfinder.FindQuery{
		Root:    ".",
		Include: []string{"*.go", "*.md"},
		Exclude: []string{"*_test.go"},
	}

	results, err = finder.FindFiles(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d files (excluding tests):\n", len(results))
	for _, result := range results {
		fmt.Printf("- %s\n", result.SourcePath)
	}
}
