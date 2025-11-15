package fulpack

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

// scanImpl implements the Scan operation.
func scanImpl(archive string, options *ScanOptions) ([]ArchiveEntry, error) {
	start := time.Now()
	var err error
	var entries []ArchiveEntry

	defer func() {
		duration := time.Since(start)
		format := detectFormat(archive)
		var bytesProcessed int64
		for _, entry := range entries {
			bytesProcessed += entry.Size
		}
		emitOperationMetrics(OperationScan, format, duration, len(entries), bytesProcessed, err)
	}()
	// Apply defaults
	opts := applyScanDefaults(options)

	// Detect format
	format := detectFormat(archive)
	if format == "" {
		err = newError(ErrCodeInvalidFormat, "could not detect archive format", OperationScan, archive, nil)
		return nil, err
	}

	// Scan based on format
	switch format {
	case ArchiveFormatTAR:
		entries, err = scanTar(archive, opts)
	case ArchiveFormatTARGZ:
		entries, err = scanTarGz(archive, opts)
	case ArchiveFormatZIP:
		entries, err = scanZip(archive, opts)
	case ArchiveFormatGZIP:
		entries, err = scanGzip(archive, opts)
	default:
		err = newError(ErrCodeInvalidFormat, "unsupported archive format", OperationScan, archive, nil)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// Apply filters
	entries = filterEntries(entries, opts)

	// Enforce max entries limit
	if len(entries) > opts.MaxEntries {
		err = newErrorf(ErrCodeMaxEntriesExceeded, OperationScan, archive, nil,
			"archive contains %d entries, exceeds limit of %d", len(entries), opts.MaxEntries)
		return nil, err
	}

	return entries, nil
}

// scanTar scans an uncompressed tar archive.
func scanTar(path string, opts *ScanOptions) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to open tar archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	var entries []ArchiveEntry

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to read tar header: %v", err)
		}

		entry := convertTarHeader(header, opts)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}

// scanTarGz scans a tar.gz archive.
func scanTarGz(path string, opts *ScanOptions) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to open tar.gz archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	var entries []ArchiveEntry

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to read tar header: %v", err)
		}

		entry := convertTarHeader(header, opts)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}

// scanZip scans a zip archive.
func scanZip(path string, opts *ScanOptions) ([]ArchiveEntry, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to open zip archive: %v", err)
	}
	defer func() { _ = zr.Close() }()

	var entries []ArchiveEntry
	for _, f := range zr.File {
		entry := convertZipFileHeader(f, opts)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}

// scanGzip scans a gzip file (single file).
func scanGzip(path string, opts *ScanOptions) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to open gzip file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	// For gzip, we can only get the original filename from the header
	name := gr.Name
	if name == "" {
		// Use archive name without .gz extension
		name = filepath.Base(path)
		if ext := filepath.Ext(name); ext == ".gz" || ext == ".gzip" {
			name = name[:len(name)-len(ext)]
		}
	}

	// We need to read the entire file to get uncompressed size
	size, err := io.Copy(io.Discard, gr)
	if err != nil {
		return nil, newErrorf(ErrCodeCorruptArchive, OperationScan, path, err, "failed to read gzip content: %v", err)
	}

	entry := ArchiveEntry{
		Path: name,
		Type: EntryTypeFile,
		Size: size,
	}

	if *opts.IncludeMetadata {
		// Get file info for compressed size
		fileInfo, err := f.Stat()
		if err == nil {
			entry.CompressedSize = fileInfo.Size()
			entry.Modified = fileInfo.ModTime()
		}
	}

	return []ArchiveEntry{entry}, nil
}

// convertTarHeader converts a tar header to ArchiveEntry.
func convertTarHeader(header *tar.Header, opts *ScanOptions) *ArchiveEntry {
	// Determine entry type
	var entryType EntryType
	switch header.Typeflag {
	case tar.TypeReg:
		entryType = EntryTypeFile
	case tar.TypeDir:
		entryType = EntryTypeDirectory
	case tar.TypeSymlink, tar.TypeLink:
		entryType = EntryTypeSymlink
	default:
		// Skip unsupported types
		return nil
	}

	entry := ArchiveEntry{
		Path: filepath.Clean(header.Name),
		Type: entryType,
		Size: header.Size,
	}

	if *opts.IncludeMetadata {
		entry.Modified = header.ModTime
		entry.Mode = uint32(header.Mode)
		if entryType == EntryTypeSymlink {
			entry.LinkTarget = header.Linkname
		}
	}

	return &entry
}

// convertZipFileHeader converts a zip file header to ArchiveEntry.
func convertZipFileHeader(f *zip.File, opts *ScanOptions) *ArchiveEntry {
	// Determine entry type
	var entryType EntryType
	if f.FileInfo().IsDir() {
		entryType = EntryTypeDirectory
	} else {
		entryType = EntryTypeFile
	}

	entry := ArchiveEntry{
		Path:           filepath.Clean(f.Name),
		Type:           entryType,
		Size:           int64(f.UncompressedSize64),
		CompressedSize: int64(f.CompressedSize64),
	}

	if *opts.IncludeMetadata {
		entry.Modified = f.Modified
		entry.Mode = uint32(f.Mode())
	}

	return &entry
}

// filterEntries applies filters from ScanOptions.
func filterEntries(entries []ArchiveEntry, opts *ScanOptions) []ArchiveEntry {
	if len(opts.EntryTypes) == 0 && opts.MaxDepth == nil &&
		len(opts.IncludePatterns) == 0 && len(opts.ExcludePatterns) == 0 {
		return entries
	}

	var filtered []ArchiveEntry
	for _, entry := range entries {
		// Normalize path for glob matching (use forward slashes)
		normalizedPath := filepath.ToSlash(entry.Path)

		// Filter by include patterns (if specified, must match at least one)
		if len(opts.IncludePatterns) > 0 {
			matched := false
			for _, pattern := range opts.IncludePatterns {
				if m, _ := doublestar.Match(pattern, normalizedPath); m {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Filter by exclude patterns (if matches any, exclude)
		if len(opts.ExcludePatterns) > 0 {
			excluded := false
			for _, pattern := range opts.ExcludePatterns {
				if m, _ := doublestar.Match(pattern, normalizedPath); m {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		// Filter by entry type
		if len(opts.EntryTypes) > 0 {
			found := false
			for _, t := range opts.EntryTypes {
				if entry.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by max depth
		if opts.MaxDepth != nil {
			depth := countPathDepth(entry.Path)
			if depth > *opts.MaxDepth {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// countPathDepth counts the depth of a path (number of directory components).
func countPathDepth(path string) int {
	if path == "" || path == "." {
		return 0
	}
	// Clean and convert to slash separators
	clean := filepath.ToSlash(filepath.Clean(path))
	// Count components by splitting on /
	parts := 0
	for _, c := range clean {
		if c == '/' {
			parts++
		}
	}
	return parts + 1 // Add 1 for the base component
}
