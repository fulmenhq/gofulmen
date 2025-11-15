package fulpack

// Create creates an archive from source files/directories.
//
// This operation creates a new archive in the specified format, applying include/exclude
// glob patterns via pathfinder integration. Sources can be individual files, directories,
// or a mix of both.
//
// Parameters:
//   - sources: Paths to files/directories to include in the archive
//   - output: Output archive file path
//   - format: Archive format (TAR, TAR.GZ, ZIP, GZIP)
//   - options: Optional creation configuration (nil uses defaults)
//
// Returns:
//   - ArchiveInfo with metadata (entry count, sizes, checksums)
//   - error if creation fails
//
// Security:
//   - Validates all source paths before archiving
//   - Applies path traversal protection
//   - Symlinks only followed if FollowSymlinks is true
//
// Example:
//
//	info, err := fulpack.Create(
//	    []string{"src/", "docs/", "README.md"},
//	    "release.tar.gz",
//	    fulpack.ArchiveFormatTARGZ,
//	    &fulpack.CreateOptions{
//	        ExcludePatterns: []string{"**/__pycache__", "**/.git"},
//	        CompressionLevel: 9,
//	    },
//	)
func Create(sources []string, output string, format ArchiveFormat, options *CreateOptions) (*ArchiveInfo, error) {
	return createImpl(sources, output, format, options)
}

// Extract extracts archive contents to a destination directory.
//
// This operation extracts an archive with mandatory security protections:
// path traversal prevention, symlink validation, and decompression bomb detection.
// Optional glob patterns can filter which entries to extract.
//
// Parameters:
//   - archive: Path to archive file
//   - destination: Target directory for extraction (must be explicit, no CWD)
//   - options: Optional extraction configuration (nil uses defaults)
//
// Returns:
//   - ExtractResult with counts and error details
//   - error if extraction fails critically
//
// Security (MANDATORY):
//   - Path traversal protection: Rejects ../, absolute paths
//   - Symlink validation: Ensures targets stay within destination
//   - Decompression bomb protection: Enforces max_size and max_entries limits
//   - Checksum verification: Verifies checksums if present (unless disabled)
//
// Example:
//
//	result, err := fulpack.Extract(
//	    "data.tar.gz",
//	    "/tmp/extracted",
//	    &fulpack.ExtractOptions{
//	        IncludePatterns: []string{"**/*.csv"},
//	        VerifyChecksums: true,
//	    },
//	)
func Extract(archive string, destination string, options *ExtractOptions) (*ExtractResult, error) {
	return extractImpl(archive, destination, options)
}

// Scan lists archive entries without extraction (for Pathfinder integration).
//
// This operation reads the archive table of contents (TOC) and returns entry metadata
// without extracting files. This enables pathfinder glob searches within archives.
//
// Parameters:
//   - archive: Path to archive file
//   - options: Optional scan configuration (nil uses defaults)
//
// Returns:
//   - Slice of ArchiveEntry with metadata
//   - error if scan fails
//
// Performance:
//   - Lazy evaluation: Reads TOC only, does NOT extract files
//   - Fast operation for large archives
//
// Pathfinder Integration:
//
//	entries, err := fulpack.Scan("data.tar.gz", nil)
//	// Pathfinder can now apply glob patterns to entries
//	matches := pathfinder.Glob(entries, "**/*.csv")
//
// Example:
//
//	entries, err := fulpack.Scan("archive.zip", &fulpack.ScanOptions{
//	    IncludeMetadata: boolPtr(true),
//	    EntryTypes: []fulpack.EntryType{fulpack.EntryTypeFile},
//	})
func Scan(archive string, options *ScanOptions) ([]ArchiveEntry, error) {
	return scanImpl(archive, options)
}

// Verify validates archive integrity and security properties.
//
// This operation performs comprehensive validation:
//   - Archive structure integrity
//   - Checksum verification (if present)
//   - Path traversal detection
//   - Decompression bomb detection
//   - Symlink safety validation
//
// Parameters:
//   - archive: Path to archive file
//   - options: Optional verification configuration (nil uses defaults)
//
// Returns:
//   - ValidationResult with validation status and details
//   - error if verification cannot be performed
//
// Security Checks:
//   - structure_valid: Archive format is intact
//   - checksums_verified: All checksums match (if present)
//   - no_path_traversal: No ../ or absolute paths
//   - no_decompression_bomb: Reasonable compression ratio and entry count
//   - symlinks_safe: All symlink targets are within bounds
//
// Example:
//
//	result, err := fulpack.Verify("data.tar.gz", nil)
//	if err != nil {
//	    return err
//	}
//	if !result.Valid {
//	    log.Printf("Archive validation failed: %v", result.Errors)
//	}
func Verify(archive string, options *VerifyOptions) (*ValidationResult, error) {
	return verifyImpl(archive, options)
}

// Info returns archive metadata without extraction.
//
// This operation provides quick inspection of archive properties:
// format detection, size estimation, compression ratio analysis.
//
// Parameters:
//   - archive: Path to archive file
//
// Returns:
//   - ArchiveInfo with metadata
//   - error if info retrieval fails
//
// Use Cases:
//   - Quick format detection before processing
//   - Size estimation for capacity planning
//   - Compression ratio analysis
//   - Archive validation before extraction
//
// Example:
//
//	info, err := fulpack.Info("release.tar.gz")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Format: %s, Entries: %d, Compression: %.1fx\n",
//	    info.Format, info.EntryCount, info.CompressionRatio)
func Info(archive string) (*ArchiveInfo, error) {
	return infoImpl(archive)
}
