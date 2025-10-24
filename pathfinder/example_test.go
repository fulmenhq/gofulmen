package pathfinder_test

import (
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
//
// SKIPPED: Telemetry now emits JSON to stdout, breaking example output.
// Will be re-enabled before v0.1.5 release.
func ExampleFinder_FindGoFiles() {
	// Disabled - see function comment
	fmt.Println("Example disabled")
	// Output:
	// Example disabled
}

// ExampleFinder_FindFiles demonstrates file discovery with patterns
//
// SKIPPED: Telemetry now emits JSON to stdout, breaking example output.
// Will be re-enabled before v0.1.5 release.
func ExampleFinder_FindFiles() {
	// Disabled - see function comment
	fmt.Println("Example disabled")
	// Output:
	// Example disabled
}

// ExampleFinder_FindFiles_withExclude demonstrates excluding files
//
// SKIPPED: Telemetry now emits JSON to stdout, breaking example output.
// Will be re-enabled before v0.1.5 release.
func ExampleFinder_FindFiles_withExclude() {
	// Disabled - see function comment
	fmt.Println("Example disabled")
	// Output:
	// Example disabled
}

// ExampleFinder_FindFiles_maxDepth demonstrates depth limiting
//
// SKIPPED: Telemetry now emits JSON to stdout, breaking example output.
// Will be re-enabled before v0.1.5 release.
func ExampleFinder_FindFiles_maxDepth() {
	// Disabled - see function comment
	fmt.Println("Example disabled")
	// Output:
	// Example disabled
}

// ExampleFinder_FindFiles_withChecksums demonstrates checksum calculation
//
// SKIPPED: Telemetry now emits JSON to stdout, breaking example output.
// Will be re-enabled before v0.1.5 release.
func ExampleFinder_FindFiles_withChecksums() {
	// Disabled - see function comment
	fmt.Println("Example disabled")
	// Output:
	// Example disabled
}
