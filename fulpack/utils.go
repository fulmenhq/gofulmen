package fulpack

import (
	"path/filepath"
	"strings"
)

// Default values for options.
const (
	DefaultCompressionLevel     = 6
	DefaultMaxSizeBytes         = 1024 * 1024 * 1024 // 1GB
	DefaultMaxEntries           = 10000
	DefaultScanMaxEntries       = 100000
	DefaultChecksumAlgorithm    = "sha256"
	DefaultOverwritePolicy      = OverwritePolicyError
	DefaultPreservePermissions  = true
	DefaultVerifyChecksums      = true
	DefaultIncludeMetadata      = true
	DefaultFollowSymlinks       = false
	DefaultCompressionRatioWarn = 100.0 // Warn if ratio > 100x
)

// boolPtr returns a pointer to a bool value (helper for option defaults).
func boolPtr(b bool) *bool {
	return &b
}

// intPtr returns a pointer to an int value (helper for option defaults).
//
//nolint:unused // Will be used in Create() implementation
func intPtr(i int) *int {
	return &i
}

// applyCreateDefaults applies default values to CreateOptions.
//
//nolint:unused // Will be used in Create() implementation
func applyCreateDefaults(opts *CreateOptions) *CreateOptions {
	if opts == nil {
		opts = &CreateOptions{}
	}
	if opts.CompressionLevel == 0 {
		opts.CompressionLevel = DefaultCompressionLevel
	}
	if opts.ChecksumAlgorithm == "" {
		opts.ChecksumAlgorithm = DefaultChecksumAlgorithm
	}
	if opts.PreservePermissions == nil {
		opts.PreservePermissions = boolPtr(DefaultPreservePermissions)
	}
	return opts
}

// applyExtractDefaults applies default values to ExtractOptions.
//
//nolint:unused // Will be used in Extract() implementation
func applyExtractDefaults(opts *ExtractOptions) *ExtractOptions {
	if opts == nil {
		opts = &ExtractOptions{}
	}
	if opts.Overwrite == "" {
		opts.Overwrite = DefaultOverwritePolicy
	}
	if opts.VerifyChecksums == nil {
		opts.VerifyChecksums = boolPtr(DefaultVerifyChecksums)
	}
	if opts.PreservePermissions == nil {
		opts.PreservePermissions = boolPtr(DefaultPreservePermissions)
	}
	if opts.MaxSize == 0 {
		opts.MaxSize = DefaultMaxSizeBytes
	}
	if opts.MaxEntries == 0 {
		opts.MaxEntries = DefaultMaxEntries
	}
	return opts
}

// applyScanDefaults applies default values to ScanOptions.
func applyScanDefaults(opts *ScanOptions) *ScanOptions {
	if opts == nil {
		opts = &ScanOptions{}
	}
	if opts.IncludeMetadata == nil {
		opts.IncludeMetadata = boolPtr(DefaultIncludeMetadata)
	}
	if opts.MaxEntries == 0 {
		opts.MaxEntries = DefaultScanMaxEntries
	}
	return opts
}

// detectFormat detects archive format from file extension.
func detectFormat(path string) ArchiveFormat {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return ArchiveFormatTARGZ
	case strings.HasSuffix(lower, ".tar"):
		return ArchiveFormatTAR
	case strings.HasSuffix(lower, ".zip"):
		return ArchiveFormatZIP
	case strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".gzip"):
		return ArchiveFormatGZIP
	default:
		return ""
	}
}

// isPathTraversal checks if a path contains traversal attempts.
//
//nolint:unused // Will be used in Extract() implementation
func isPathTraversal(path string) bool {
	// Check for absolute paths
	if filepath.IsAbs(path) {
		return true
	}

	// Check for ../ components
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return true
	}

	return false
}

// sanitizePath sanitizes a path for safe extraction.
// Returns the sanitized path and a boolean indicating if it was modified.
//
//nolint:unused // Will be used in Extract() implementation
func sanitizePath(path string) (string, bool) {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Remove leading slashes (make relative)
	for strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		cleaned = cleaned[1:]
	}

	// Check if modification occurred
	modified := cleaned != path

	return cleaned, modified
}

// isWithinBounds checks if a target path is within the destination bounds.
//
//nolint:unused // Will be used in Extract() implementation
func isWithinBounds(target, destination string) bool {
	// Clean and make absolute
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	absDest, err := filepath.Abs(destination)
	if err != nil {
		return false
	}

	// Check if target is under destination
	rel, err := filepath.Rel(absDest, absTarget)
	if err != nil {
		return false
	}

	// If relative path starts with .., it's outside bounds
	return !strings.HasPrefix(rel, "..")
}

// calculateCompressionRatio calculates compression ratio.
func calculateCompressionRatio(uncompressed, compressed int64) float64 {
	if compressed == 0 {
		return 0
	}
	return float64(uncompressed) / float64(compressed)
}

// isDecompressionBomb checks if archive exhibits decompression bomb characteristics.
//
//nolint:unused // Will be used in Verify() implementation
func isDecompressionBomb(uncompressed, compressed int64, entryCount int, maxEntries int) bool {
	// Check compression ratio (warn if > 100x)
	ratio := calculateCompressionRatio(uncompressed, compressed)
	if ratio > DefaultCompressionRatioWarn {
		return true
	}

	// Check entry count
	if entryCount > maxEntries {
		return true
	}

	return false
}
