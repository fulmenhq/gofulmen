package fulpack

import (
	"time"
)

// verifyImpl implements the Verify operation.
func verifyImpl(archive string, options *VerifyOptions) (*ValidationResult, error) {
	start := time.Now()
	var err error
	var result *ValidationResult

	defer func() {
		duration := time.Since(start)
		format := detectFormat(archive)
		var entryCount int
		var bytesProcessed int64
		if result != nil {
			entryCount = result.EntryCount
		}
		emitOperationMetrics(OperationVerify, format, duration, entryCount, bytesProcessed, err)
	}()

	// Initialize result with default checks
	result = &ValidationResult{
		Valid: true,
		ChecksPerformed: []string{
			"structure_valid",
			"no_path_traversal",
			"no_decompression_bomb",
			"symlinks_safe",
		},
	}

	// Step 1: Verify archive structure by scanning entries
	entries, scanErr := scanImpl(archive, nil)
	if scanErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrCodeCorruptArchive,
			Message: "Failed to scan archive structure",
			Details: map[string]any{"error": scanErr.Error()},
		})
		// Return result, not error - Verify() should return validation results
		return result, nil
	}

	result.EntryCount = len(entries)

	// Step 2: Check each entry for security issues
	var totalUncompressedSize int64
	for _, entry := range entries {
		totalUncompressedSize += entry.Size

		// Check for path traversal
		if isPathTraversal(entry.Path) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodePathTraversal,
				Message: "Path traversal detected",
				Path:    entry.Path,
			})
		}

		// Check symlink safety
		if entry.Type == EntryTypeSymlink {
			if entry.LinkTarget == "" {
				result.Warnings = append(result.Warnings, "Symlink missing target: "+entry.Path)
				continue
			}

			// For symlinks, we need to check if the target would escape bounds
			// We can't fully validate without extracting, but we can check for obvious issues
			if isPathTraversal(entry.LinkTarget) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrCodeSymlinkEscape,
					Message: "Symlink target contains path traversal",
					Path:    entry.Path,
					Details: map[string]any{"target": entry.LinkTarget},
				})
			}
		}
	}

	// Step 3: Check for decompression bomb characteristics
	info, infoErr := infoImpl(archive)
	if infoErr != nil {
		result.Warnings = append(result.Warnings, "Could not retrieve archive info for compression ratio check")
	} else {
		if isDecompressionBomb(info.TotalSize, info.CompressedSize, info.EntryCount, DefaultMaxEntries) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodeDecompressionBomb,
				Message: "Archive exhibits decompression bomb characteristics",
				Details: map[string]any{
					"compression_ratio": info.CompressionRatio,
					"entry_count":       info.EntryCount,
					"total_size":        info.TotalSize,
					"compressed_size":   info.CompressedSize,
				},
			})
		}
	}

	// Step 4: Check for checksums (if present, we mark as verified - actual verification happens during Extract)
	if info != nil && info.HasChecksums {
		result.ChecksPerformed = append(result.ChecksPerformed, "checksums_present")
		result.Warnings = append(result.Warnings, "Archive contains checksums - use Extract with VerifyChecksums=true for full validation")
	}

	return result, nil
}
