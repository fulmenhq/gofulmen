// Package docscribe provides lightweight processing capabilities for markdown
// and YAML documentation from any source.
//
// This is a standalone, source-agnostic module that handles common documentation
// tasks like frontmatter extraction, header parsing, format detection, and
// multi-document splitting. It works with content from Crucible, Cosmography,
// local files, or any other documentation source.
//
// # Key Design Principle
//
// This package is intentionally source-agnostic. It processes raw content
// ([]byte, string, io.Reader) without coupling to specific storage or access
// patterns. Source-specific integrations (e.g., Crucible shim) use this package
// but live in their own namespaces.
//
// # Core Capabilities
//
// Frontmatter Processing:
//   - ParseFrontmatter: Extract both metadata and clean content
//   - ExtractMetadata: Get only the YAML frontmatter metadata
//   - StripFrontmatter: Remove frontmatter, return clean markdown
//
// Header Extraction:
//   - ExtractHeaders: Extract all markdown headers with hierarchy, anchors, and line numbers
//
// Format Detection:
//   - DetectFormat: Heuristic-based format detection (markdown, yaml, json, etc.)
//
// Document Inspection:
//   - InspectDocument: Quick analysis without full parsing (<1ms target)
//
// Multi-Document Handling:
//   - SplitDocuments: Split YAML streams and concatenated markdown documents
//
// # Usage Example
//
//	import (
//	    "github.com/fulmenhq/gofulmen/docscribe"
//	    "github.com/fulmenhq/gofulmen/crucible"
//	)
//
//	// Get documentation from any source
//	content, err := crucible.GetDoc("standards/coding/go.md")
//	if err != nil {
//	    return err
//	}
//
//	// Quick inspection
//	info, err := docscribe.InspectDocument([]byte(content))
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Format: %s, Sections: %d\n", info.Format, info.EstimatedSections)
//
//	// Extract frontmatter and content
//	body, metadata, err := docscribe.ParseFrontmatter([]byte(content))
//	if err != nil {
//	    return err
//	}
//	if metadata != nil {
//	    fmt.Printf("Title: %s\n", metadata["title"])
//	    fmt.Printf("Status: %s\n", metadata["status"])
//	}
//
//	// Extract headers for TOC generation
//	headers, err := docscribe.ExtractHeaders([]byte(content))
//	if err != nil {
//	    return err
//	}
//	for _, h := range headers {
//	    fmt.Printf("%s %s (line %d)\n", strings.Repeat("#", h.Level), h.Text, h.LineNumber)
//	}
//
// # Multi-Document Support
//
// The package correctly handles the dual purpose of "---" separators:
//   - Frontmatter delimiter in markdown files
//   - Document separator in YAML streams
//
// This is critical for processing Kubernetes manifests, CI/CD outputs,
// and bundled documentation sets.
//
// # Performance Targets
//
//   - InspectDocument: <1ms for 100KB documents
//   - ParseFrontmatter: <5ms
//   - SplitDocuments: <10ms for 10-document stream
//   - ExtractHeaders: <50ms for 1MB document
//
// # Error Handling
//
// The package uses typed errors for different failure modes:
//   - ParseError: Malformed YAML or content structure issues (includes line numbers)
//   - FormatError: Content doesn't match expected format
//
// All errors implement standard error unwrapping for inspection.
package docscribe
