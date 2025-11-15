package fulpack

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

// extractImpl implements the Extract operation.
func extractImpl(archive string, destination string, options *ExtractOptions) (*ExtractResult, error) {
	start := time.Now()
	var err error
	var result *ExtractResult

	defer func() {
		duration := time.Since(start)
		format := detectFormat(archive)
		var entryCount int
		var bytesProcessed int64
		if result != nil {
			entryCount = result.ExtractedCount
			bytesProcessed = result.BytesWritten
		}
		emitOperationMetrics(OperationExtract, format, duration, entryCount, bytesProcessed, err)
	}()

	// Apply defaults
	opts := applyExtractDefaults(options)

	// Initialize result
	result = &ExtractResult{
		ExtractedCount: 0,
		SkippedCount:   0,
		ErrorCount:     0,
		BytesWritten:   0,
	}

	// Validate destination is not empty
	if destination == "" {
		err = newError(ErrCodeInvalidFormat, "destination cannot be empty", OperationExtract, "", nil)
		return nil, err
	}

	// Create destination directory if it doesn't exist
	if mkdirErr := os.MkdirAll(destination, 0755); mkdirErr != nil {
		err = newErrorf(ErrCodeFileExists, OperationExtract, destination, mkdirErr,
			"failed to create destination directory: %v", mkdirErr)
		return nil, err
	}

	// Detect format
	format := detectFormat(archive)
	if format == "" {
		err = newError(ErrCodeInvalidFormat, "could not detect archive format", OperationExtract, archive, nil)
		return nil, err
	}

	// Extract based on format
	switch format {
	case ArchiveFormatTAR:
		err = extractTar(archive, destination, opts, result)
	case ArchiveFormatTARGZ:
		err = extractTarGz(archive, destination, opts, result)
	case ArchiveFormatZIP:
		err = extractZip(archive, destination, opts, result)
	case ArchiveFormatGZIP:
		err = extractGzip(archive, destination, opts, result)
	default:
		err = newError(ErrCodeInvalidFormat, "unsupported archive format", OperationExtract, archive, nil)
		return nil, err
	}

	if err != nil {
		return result, err
	}

	return result, nil
}

// extractTar extracts an uncompressed tar archive.
func extractTar(archivePath string, destination string, opts *ExtractOptions, result *ExtractResult) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to open tar archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	return extractTarReader(tr, destination, opts, result, archivePath)
}

// extractTarGz extracts a tar.gz archive.
func extractTarGz(archivePath string, destination string, opts *ExtractOptions, result *ExtractResult) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to open tar.gz archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	return extractTarReader(tr, destination, opts, result, archivePath)
}

// extractTarReader extracts entries from a tar reader.
func extractTarReader(tr *tar.Reader, destination string, opts *ExtractOptions, result *ExtractResult, archivePath string) error {
	var totalUncompressedSize int64
	var entryCount int

	// Get compressed size for decompression bomb detection
	var compressedSize int64
	if fileInfo, err := os.Stat(archivePath); err == nil {
		compressedSize = fileInfo.Size()
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
				"failed to read tar header: %v", err)
		}

		entryCount++

		// Security: Check max entries limit
		if entryCount > opts.MaxEntries {
			return newErrorf(ErrCodeMaxEntriesExceeded, OperationExtract, archivePath, nil,
				"archive contains more than %d entries", opts.MaxEntries)
		}

		// Security: Check for path traversal
		if isPathTraversal(header.Name) {
			result.ErrorCount++
			result.Errors = append(result.Errors, ExtractionError{
				Path:  header.Name,
				Error: "path traversal detected",
				Code:  ErrCodePathTraversal,
			})
			continue
		}

		// Normalize path (convert to slash separators)
		normalizedPath := filepath.ToSlash(filepath.Clean(header.Name))

		// Apply include/exclude filters
		if !shouldExtract(normalizedPath, opts.IncludePatterns, opts.ExcludePatterns) {
			result.SkippedCount++
			continue
		}

		// Build target path
		targetPath := filepath.Join(destination, header.Name)

		// Security: Verify target is within destination bounds
		if !isWithinBounds(targetPath, destination) {
			result.ErrorCount++
			result.Errors = append(result.Errors, ExtractionError{
				Path:  header.Name,
				Error: "path escapes destination bounds",
				Code:  ErrCodePathTraversal,
			})
			continue
		}

		// Extract based on type
		switch header.Typeflag {
		case tar.TypeDir:
			if extractErr := extractDirectory(targetPath, header.Mode, opts); extractErr != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  header.Name,
					Error: extractErr.Error(),
				})
				continue
			}
			result.ExtractedCount++

		case tar.TypeReg:
			// Security: Check max size limit
			totalUncompressedSize += header.Size
			if totalUncompressedSize > opts.MaxSize {
				return newErrorf(ErrCodeMaxSizeExceeded, OperationExtract, archivePath, nil,
					"total uncompressed size exceeds limit of %d bytes", opts.MaxSize)
			}

			// Security: Check for decompression bomb via compression ratio
			if compressedSize > 0 && isDecompressionBomb(totalUncompressedSize, compressedSize, entryCount, opts.MaxEntries) {
				return newErrorf(ErrCodeDecompressionBomb, OperationExtract, archivePath, nil,
					"decompression bomb detected: ratio %.1fx, %d entries",
					calculateCompressionRatio(totalUncompressedSize, compressedSize), entryCount)
			}

			bytesWritten, extractErr := extractFile(tr, targetPath, header.Mode, header.Size, opts)
			if extractErr != nil {
				if extractErr == errSkipFile {
					result.SkippedCount++
					continue
				}
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  header.Name,
					Error: extractErr.Error(),
				})
				continue
			}
			result.ExtractedCount++
			result.BytesWritten += bytesWritten

		case tar.TypeSymlink, tar.TypeLink:
			// Security: Validate symlink target
			linkTarget := filepath.Join(filepath.Dir(targetPath), header.Linkname)
			if !isWithinBounds(linkTarget, destination) {
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  header.Name,
					Error: "symlink target escapes destination bounds",
					Code:  ErrCodeSymlinkEscape,
				})
				continue
			}

			if extractErr := extractSymlink(targetPath, header.Linkname, opts); extractErr != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  header.Name,
					Error: extractErr.Error(),
				})
				continue
			}
			result.ExtractedCount++

		default:
			// Skip unsupported types
			result.SkippedCount++
		}
	}

	return nil
}

// extractZip extracts a zip archive.
func extractZip(archivePath string, destination string, opts *ExtractOptions, result *ExtractResult) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to open zip archive: %v", err)
	}
	defer func() { _ = zr.Close() }()

	var totalUncompressedSize int64

	// Get compressed size for decompression bomb detection
	var compressedSize int64
	if fileInfo, err := os.Stat(archivePath); err == nil {
		compressedSize = fileInfo.Size()
	}

	for i, f := range zr.File {
		// Security: Check max entries limit
		if i+1 > opts.MaxEntries {
			return newErrorf(ErrCodeMaxEntriesExceeded, OperationExtract, archivePath, nil,
				"archive contains more than %d entries", opts.MaxEntries)
		}

		// Security: Check for path traversal
		if isPathTraversal(f.Name) {
			result.ErrorCount++
			result.Errors = append(result.Errors, ExtractionError{
				Path:  f.Name,
				Error: "path traversal detected",
				Code:  ErrCodePathTraversal,
			})
			continue
		}

		// Normalize path
		normalizedPath := filepath.ToSlash(filepath.Clean(f.Name))

		// Apply include/exclude filters
		if !shouldExtract(normalizedPath, opts.IncludePatterns, opts.ExcludePatterns) {
			result.SkippedCount++
			continue
		}

		// Build target path
		targetPath := filepath.Join(destination, f.Name)

		// Security: Verify target is within destination bounds
		if !isWithinBounds(targetPath, destination) {
			result.ErrorCount++
			result.Errors = append(result.Errors, ExtractionError{
				Path:  f.Name,
				Error: "path escapes destination bounds",
				Code:  ErrCodePathTraversal,
			})
			continue
		}

		// Extract based on type
		if f.FileInfo().IsDir() {
			if extractErr := extractDirectory(targetPath, int64(f.Mode()), opts); extractErr != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  f.Name,
					Error: extractErr.Error(),
				})
				continue
			}
			result.ExtractedCount++
		} else {
			// Security: Check max size limit
			totalUncompressedSize += int64(f.UncompressedSize64)
			if totalUncompressedSize > opts.MaxSize {
				return newErrorf(ErrCodeMaxSizeExceeded, OperationExtract, archivePath, nil,
					"total uncompressed size exceeds limit of %d bytes", opts.MaxSize)
			}

			// Security: Check for decompression bomb via compression ratio
			if compressedSize > 0 && isDecompressionBomb(totalUncompressedSize, compressedSize, i+1, opts.MaxEntries) {
				return newErrorf(ErrCodeDecompressionBomb, OperationExtract, archivePath, nil,
					"decompression bomb detected: ratio %.1fx, %d entries",
					calculateCompressionRatio(totalUncompressedSize, compressedSize), i+1)
			}

			rc, openErr := f.Open()
			if openErr != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  f.Name,
					Error: openErr.Error(),
				})
				continue
			}

			bytesWritten, extractErr := extractFile(rc, targetPath, int64(f.Mode()), int64(f.UncompressedSize64), opts)
			_ = rc.Close()

			if extractErr != nil {
				if extractErr == errSkipFile {
					result.SkippedCount++
					continue
				}
				result.ErrorCount++
				result.Errors = append(result.Errors, ExtractionError{
					Path:  f.Name,
					Error: extractErr.Error(),
				})
				continue
			}
			result.ExtractedCount++
			result.BytesWritten += bytesWritten
		}
	}

	return nil
}

// extractGzip extracts a gzip file (single file).
func extractGzip(archivePath string, destination string, opts *ExtractOptions, result *ExtractResult) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to open gzip file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationExtract, archivePath, err,
			"failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	// Get original filename from gzip header
	name := gr.Name
	if name == "" {
		// Use archive name without .gz extension
		name = filepath.Base(archivePath)
		if ext := filepath.Ext(name); ext == ".gz" || ext == ".gzip" {
			name = name[:len(name)-len(ext)]
		}
	}

	targetPath := filepath.Join(destination, name)

	// Security: Verify target is within destination bounds
	if !isWithinBounds(targetPath, destination) {
		return newErrorf(ErrCodePathTraversal, OperationExtract, archivePath, nil,
			"extracted file would escape destination bounds")
	}

	// Extract the single file
	bytesWritten, extractErr := extractFile(gr, targetPath, 0644, -1, opts)
	if extractErr != nil {
		result.ErrorCount++
		result.Errors = append(result.Errors, ExtractionError{
			Path:  name,
			Error: extractErr.Error(),
		})
		return extractErr
	}

	result.ExtractedCount++
	result.BytesWritten += bytesWritten

	// Check max size limit after extraction
	if result.BytesWritten > opts.MaxSize {
		// Clean up extracted file
		_ = os.Remove(targetPath)
		return newErrorf(ErrCodeMaxSizeExceeded, OperationExtract, archivePath, nil,
			"uncompressed size (%d bytes) exceeds limit of %d bytes", result.BytesWritten, opts.MaxSize)
	}

	return nil
}

// extractDirectory creates a directory with proper permissions.
func extractDirectory(targetPath string, mode int64, opts *ExtractOptions) error {
	// Check if directory already exists
	if info, err := os.Stat(targetPath); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("target exists and is not a directory: %s", targetPath)
		}
		// Directory exists, skip
		return nil
	}

	// Create directory
	perm := os.FileMode(0755)
	if *opts.PreservePermissions && mode != 0 {
		perm = os.FileMode(mode)
	}

	if err := os.MkdirAll(targetPath, perm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return nil
}

// errSkipFile is returned when a file is skipped due to overwrite policy
var errSkipFile = fmt.Errorf("file skipped")

// extractFile extracts a file from a reader to target path.
func extractFile(reader io.Reader, targetPath string, mode int64, expectedSize int64, opts *ExtractOptions) (int64, error) {
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Check overwrite policy
	if _, err := os.Stat(targetPath); err == nil {
		switch opts.Overwrite {
		case OverwritePolicyError:
			return 0, fmt.Errorf("file already exists: %s", targetPath)
		case OverwritePolicySkip:
			return 0, errSkipFile
		case OverwritePolicyOverwrite:
			// Continue to overwrite
		}
	}

	// Create file
	perm := os.FileMode(0644)
	if *opts.PreservePermissions && mode != 0 {
		perm = os.FileMode(mode)
	}

	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	// Copy data
	bytesWritten, err := io.Copy(outFile, reader)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed to write file: %v", err)
	}

	// Verify expected size if provided
	if expectedSize >= 0 && bytesWritten != expectedSize {
		return bytesWritten, fmt.Errorf("size mismatch: expected %d bytes, wrote %d bytes", expectedSize, bytesWritten)
	}

	return bytesWritten, nil
}

// extractSymlink creates a symbolic link.
func extractSymlink(targetPath string, linkTarget string, opts *ExtractOptions) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Check overwrite policy
	if _, err := os.Lstat(targetPath); err == nil {
		switch opts.Overwrite {
		case OverwritePolicyError:
			return fmt.Errorf("symlink already exists: %s", targetPath)
		case OverwritePolicySkip:
			return nil
		case OverwritePolicyOverwrite:
			// Remove existing symlink
			if removeErr := os.Remove(targetPath); removeErr != nil {
				return fmt.Errorf("failed to remove existing symlink: %v", removeErr)
			}
		}
	}

	// Create symlink
	if err := os.Symlink(linkTarget, targetPath); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

// shouldExtract checks if an entry should be extracted based on include/exclude patterns.
func shouldExtract(normalizedPath string, includePatterns []string, excludePatterns []string) bool {
	// Check exclude patterns first - if matches any, exclude
	for _, pattern := range excludePatterns {
		if matched, _ := doublestar.Match(pattern, normalizedPath); matched {
			return false
		}
	}

	// If no include patterns specified, extract everything (that wasn't excluded)
	if len(includePatterns) == 0 {
		return true
	}

	// Must match at least one include pattern
	for _, pattern := range includePatterns {
		if matched, _ := doublestar.Match(pattern, normalizedPath); matched {
			return true
		}
	}

	return false
}
