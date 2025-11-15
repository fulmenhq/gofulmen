package fulpack

import "time"

// ArchiveFormat represents supported archive format identifiers.
// Generated from: schemas/crucible-go/taxonomy/library/fulpack/archive-formats/v1.0.0/formats.yaml
type ArchiveFormat string

const (
	// ArchiveFormatTAR is uncompressed POSIX tar format.
	// Use cases: Maximum speed, pre-compressed data, streaming scenarios.
	ArchiveFormatTAR ArchiveFormat = "tar"

	// ArchiveFormatTARGZ is POSIX tar with gzip compression.
	// Use cases: General purpose, best compatibility, balanced compression/speed.
	ArchiveFormatTARGZ ArchiveFormat = "tar.gz"

	// ArchiveFormatZIP is ZIP archive with deflate compression.
	// Use cases: Windows compatibility, random access, wide tool support.
	ArchiveFormatZIP ArchiveFormat = "zip"

	// ArchiveFormatGZIP is single-file gzip compression.
	// Use cases: Single file compression, streaming, log files.
	ArchiveFormatGZIP ArchiveFormat = "gzip"
)

// EntryType represents archive entry type identifiers.
// Generated from: schemas/crucible-go/taxonomy/library/fulpack/entry-types/v1.0.0/types.yaml
type EntryType string

const (
	// EntryTypeFile represents a regular file with data.
	EntryTypeFile EntryType = "file"

	// EntryTypeDirectory represents a directory/folder entry.
	EntryTypeDirectory EntryType = "directory"

	// EntryTypeSymlink represents a symbolic link.
	// Security note: Must validate target is within archive bounds.
	EntryTypeSymlink EntryType = "symlink"
)

// Operation represents canonical archive operation identifiers.
// Generated from: schemas/crucible-go/taxonomy/library/fulpack/operations/v1.0.0/operations.yaml
type Operation string

const (
	// OperationCreate represents creating a new archive.
	OperationCreate Operation = "create"

	// OperationExtract represents extracting archive contents.
	OperationExtract Operation = "extract"

	// OperationScan represents listing archive entries.
	OperationScan Operation = "scan"

	// OperationVerify represents validating archive integrity.
	OperationVerify Operation = "verify"

	// OperationInfo represents getting archive metadata.
	OperationInfo Operation = "info"
)

// OverwritePolicy defines behavior when extracting over existing files.
type OverwritePolicy string

const (
	// OverwritePolicyError fails extraction if target file exists.
	OverwritePolicyError OverwritePolicy = "error"

	// OverwritePolicySkip skips existing files without error.
	OverwritePolicySkip OverwritePolicy = "skip"

	// OverwritePolicyOverwrite replaces existing files.
	OverwritePolicyOverwrite OverwritePolicy = "overwrite"
)

// CreateOptions configures archive creation behavior.
type CreateOptions struct {
	// CompressionLevel specifies compression level (1-9, default: 6).
	// Ignored for ArchiveFormatTAR (uncompressed).
	CompressionLevel int `json:"compression_level,omitempty"`

	// IncludePatterns specifies glob patterns to include (e.g., ["**/*.py", "**/*.md"]).
	IncludePatterns []string `json:"include_patterns,omitempty"`

	// ExcludePatterns specifies glob patterns to exclude (e.g., ["**/__pycache__", "**/.git"]).
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`

	// ChecksumAlgorithm specifies checksum algorithm ("xxh3-128", "sha256", "sha512", "sha1", "md5").
	// Default: "sha256"
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	// PreservePermissions preserves file permissions (default: true).
	PreservePermissions *bool `json:"preserve_permissions,omitempty"`

	// FollowSymlinks follows symbolic links (default: false).
	FollowSymlinks bool `json:"follow_symlinks,omitempty"`
}

// ExtractOptions configures archive extraction behavior.
type ExtractOptions struct {
	// Overwrite specifies overwrite policy for existing files (default: "error").
	Overwrite OverwritePolicy `json:"overwrite,omitempty"`

	// VerifyChecksums enables checksum verification (default: true).
	VerifyChecksums *bool `json:"verify_checksums,omitempty"`

	// PreservePermissions preserves file permissions (default: true).
	PreservePermissions *bool `json:"preserve_permissions,omitempty"`

	// IncludePatterns specifies glob patterns to extract (e.g., ["**/*.csv"]).
	IncludePatterns []string `json:"include_patterns,omitempty"`

	// MaxSize specifies maximum total uncompressed size in bytes (default: 1GB, bomb protection).
	MaxSize int64 `json:"max_size,omitempty"`

	// MaxEntries specifies maximum number of entries (default: 10000, bomb protection).
	MaxEntries int `json:"max_entries,omitempty"`
}

// ScanOptions configures archive scanning behavior.
type ScanOptions struct {
	// IncludeMetadata includes detailed metadata in results (default: true).
	IncludeMetadata *bool `json:"include_metadata,omitempty"`

	// EntryTypes filters by entry types (["file", "directory", "symlink"]).
	EntryTypes []EntryType `json:"entry_types,omitempty"`

	// MaxDepth limits directory traversal depth (nil = unlimited).
	MaxDepth *int `json:"max_depth,omitempty"`

	// MaxEntries limits total entries returned (default: 100000, safety limit).
	MaxEntries int `json:"max_entries,omitempty"`

	// IncludePatterns specifies glob patterns to include (e.g., ["**/*.csv"]).
	// Pathfinder integration for archive filtering.
	IncludePatterns []string `json:"include_patterns,omitempty"`

	// ExcludePatterns specifies glob patterns to exclude (e.g., ["**/__pycache__"]).
	// Pathfinder integration for archive filtering.
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
}

// VerifyOptions configures archive verification behavior.
type VerifyOptions struct {
	// Future use - placeholder for verification configuration.
}

// ArchiveInfo contains archive metadata and statistics.
type ArchiveInfo struct {
	// Format is the detected archive format.
	Format ArchiveFormat `json:"format"`

	// Compression is the compression algorithm used.
	Compression string `json:"compression"`

	// EntryCount is the total number of entries in the archive.
	EntryCount int `json:"entry_count"`

	// TotalSize is the total uncompressed size in bytes.
	TotalSize int64 `json:"total_size"`

	// CompressedSize is the archive file size in bytes.
	CompressedSize int64 `json:"compressed_size"`

	// CompressionRatio is the compression ratio (uncompressed / compressed).
	// For uncompressed tar, this will be 1.0.
	CompressionRatio float64 `json:"compression_ratio"`

	// HasChecksums indicates if archive contains checksums.
	HasChecksums bool `json:"has_checksums"`

	// ChecksumAlgorithm is the algorithm used (if HasChecksums is true).
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	// Created is the archive creation timestamp (if available).
	Created *time.Time `json:"created,omitempty"`

	// Checksums maps checksum algorithm to digest value.
	Checksums map[string]string `json:"checksums,omitempty"`
}

// ArchiveEntry represents a single entry within an archive.
type ArchiveEntry struct {
	// Path is the normalized entry path within the archive.
	Path string `json:"path"`

	// Type is the entry type (file, directory, symlink).
	Type EntryType `json:"type"`

	// Size is the uncompressed size in bytes.
	Size int64 `json:"size"`

	// CompressedSize is the compressed size in bytes (if available).
	CompressedSize int64 `json:"compressed_size,omitempty"`

	// Modified is the modification timestamp.
	Modified time.Time `json:"modified"`

	// Checksum is the entry checksum (if available).
	Checksum string `json:"checksum,omitempty"`

	// Mode is the Unix file permissions (if available).
	Mode uint32 `json:"mode,omitempty"`

	// LinkTarget is the symlink target path (for symlinks only).
	LinkTarget string `json:"link_target,omitempty"`
}

// ExtractResult contains extraction operation results.
type ExtractResult struct {
	// ExtractedCount is the number of successfully extracted entries.
	ExtractedCount int `json:"extracted_count"`

	// SkippedCount is the number of skipped entries (e.g., due to overwrite policy).
	SkippedCount int `json:"skipped_count"`

	// ErrorCount is the number of entries that failed extraction.
	ErrorCount int `json:"error_count"`

	// Errors contains extraction error details.
	Errors []ExtractionError `json:"errors,omitempty"`

	// BytesWritten is the total bytes written to disk.
	BytesWritten int64 `json:"bytes_written"`
}

// ExtractionError represents an error during extraction of a specific entry.
type ExtractionError struct {
	// Path is the entry path that failed.
	Path string `json:"path"`

	// Error is the error message.
	Error string `json:"error"`

	// Code is the error code (e.g., "PATH_TRAVERSAL", "CHECKSUM_MISMATCH").
	Code string `json:"code,omitempty"`
}

// ValidationResult contains archive verification results.
type ValidationResult struct {
	// Valid indicates if the archive is intact and safe.
	Valid bool `json:"valid"`

	// Errors contains validation errors (e.g., corruption, security issues).
	Errors []ValidationError `json:"errors,omitempty"`

	// Warnings contains non-fatal warnings (e.g., missing checksums).
	Warnings []string `json:"warnings,omitempty"`

	// EntryCount is the number of entries validated.
	EntryCount int `json:"entry_count"`

	// ChecksumsVerified is the count of successful checksum validations.
	ChecksumsVerified int `json:"checksums_verified"`

	// ChecksPerformed lists the checks that were performed.
	ChecksPerformed []string `json:"checks_performed"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	// Message is the error description.
	Message string `json:"message"`

	// Code is the error code (e.g., "PATH_TRAVERSAL", "DECOMPRESSION_BOMB").
	Code string `json:"code"`

	// Path is the affected entry path (if applicable).
	Path string `json:"path,omitempty"`

	// Details provides additional context.
	Details map[string]any `json:"details,omitempty"`
}
