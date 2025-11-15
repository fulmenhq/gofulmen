// Package fulpack provides canonical archive operations for tar, tar.gz, zip, and gzip formats.
//
// This package implements the Fulpack Archive Module Standard from the Crucible specification,
// providing a unified façade over Go's standard library archive packages (archive/tar, archive/zip,
// compress/gzip) with security by default, Pathfinder integration, and comprehensive observability.
//
// # Overview
//
// Fulpack is a Common-tier helper library module designed to replace ad-hoc archive handling
// with a consistent, secure, and well-instrumented API. All operations follow the Canonical
// Façade Principle, ensuring cross-language consistency with pyfulmen and tsfulmen.
//
// # Supported Formats
//
// The following archive formats are supported (defined by ArchiveFormat enum):
//
//   - ArchiveFormatTAR:    Uncompressed POSIX tar (maximum speed, pre-compressed data)
//   - ArchiveFormatTARGZ:  Tar with gzip compression (general purpose, best compatibility)
//   - ArchiveFormatZIP:    ZIP with deflate compression (Windows compatibility, random access)
//   - ArchiveFormatGZIP:   Single-file gzip compression (streaming compression, log files)
//
// # Core Operations
//
// Five canonical operations provide complete archive lifecycle management:
//
//   - Create():  Create archives from source files/directories with glob filtering
//   - Extract(): Extract archives with security protections and pattern filtering
//   - Scan():    List archive entries for Pathfinder integration (no extraction)
//   - Verify():  Validate archive integrity, checksums, and security properties
//   - Info():    Get archive metadata (format, size, compression ratio)
//
// # Security by Default
//
// All operations include mandatory security protections:
//
//   - Path traversal protection: Rejects ../, absolute paths, escapes outside bounds
//   - Symlink validation: Ensures symlink targets stay within archive bounds
//   - Decompression bomb detection: Enforces max_size and max_entries limits
//   - Checksum verification: Optional cryptographic integrity validation
//
// # Pathfinder Integration
//
// Fulpack integrates with the pathfinder module for unified glob-based file discovery
// across filesystems and archives. The Scan() operation returns entries compatible with
// pathfinder's matching engine.
//
// # Observability
//
// All operations emit structured telemetry via the telemetry module:
//
//   - Operation counters (fulpack.operations.total)
//   - Duration histograms (fulpack.operation.duration_ms)
//   - Bytes processed metrics (fulpack.bytes.processed)
//   - Security violation counters (fulpack.security.violations_total)
//
// # Dependencies
//
//   - pathfinder (required): Glob-based file discovery and pattern matching
//   - fulhash (required):    Checksum generation and verification (xxh3-128, SHA-256)
//   - telemetry (required):  Observability and metrics
//   - errors (required):     Error envelopes and foundry exit codes
//
// # Basic Usage
//
//	// Create a compressed tar archive
//	info, err := fulpack.Create(
//	    []string{"src/", "docs/", "README.md"},
//	    "release.tar.gz",
//	    fulpack.ArchiveFormatTARGZ,
//	    &fulpack.CreateOptions{
//	        ExcludePatterns: []string{"**/__pycache__", "**/.git"},
//	        CompressionLevel: 9,
//	    },
//	)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Created archive: %d entries, %.1fx compression\n",
//	    info.EntryCount, info.CompressionRatio)
//
//	// Extract with security validation
//	result, err := fulpack.Extract(
//	    "data.tar.gz",
//	    "/tmp/extracted",
//	    &fulpack.ExtractOptions{
//	        IncludePatterns: []string{"**/*.csv"},
//	        VerifyChecksums: true,
//	    },
//	)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Extracted %d files, skipped %d\n",
//	    result.ExtractedCount, result.SkippedCount)
//
//	// Scan archive without extraction (Pathfinder integration)
//	entries, err := fulpack.Scan("archive.zip", nil)
//	if err != nil {
//	    return err
//	}
//	for _, entry := range entries {
//	    fmt.Printf("%s (%s, %d bytes)\n", entry.Path, entry.Type, entry.Size)
//	}
//
//	// Get archive metadata
//	info, err = fulpack.Info("release.tar.gz")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Format: %s, %d entries, %.1fx compression\n",
//	    info.Format, info.EntryCount, info.CompressionRatio)
//
// # Specification
//
// This package implements:
//   - Fulpack Archive Module Standard (docs/crucible-go/standards/library/modules/fulpack.md)
//   - Archive Formats Taxonomy v1.0.0 (schemas/crucible-go/taxonomy/library/fulpack/archive-formats/)
//   - Operations Taxonomy v1.0.0 (schemas/crucible-go/taxonomy/library/fulpack/operations/)
//   - Entry Types Taxonomy v1.0.0 (schemas/crucible-go/taxonomy/library/fulpack/entry-types/)
package fulpack
