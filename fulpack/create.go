package fulpack

import (
	"archive/tar"
	"archive/zip"
	"compress/flate"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fulmenhq/gofulmen/fulhash"
	"github.com/fulmenhq/gofulmen/pathfinder"
)

// createImpl implements the Create operation.
func createImpl(sources []string, output string, format ArchiveFormat, options *CreateOptions) (*ArchiveInfo, error) {
	start := time.Now()
	var err error
	var info *ArchiveInfo

	defer func() {
		duration := time.Since(start)
		var entryCount int
		var bytesProcessed int64
		if info != nil {
			entryCount = info.EntryCount
			bytesProcessed = info.TotalSize
		}
		emitOperationMetrics(OperationCreate, format, duration, entryCount, bytesProcessed, err)
	}()

	// Apply defaults
	opts := applyCreateDefaults(options)

	// Validate sources
	if len(sources) == 0 {
		err = newError(ErrCodeInvalidFormat, "no source files specified", OperationCreate, "", nil)
		return nil, err
	}

	// Validate output path
	if output == "" {
		err = newError(ErrCodeInvalidFormat, "output path cannot be empty", OperationCreate, "", nil)
		return nil, err
	}

	// Initialize archive info
	info = &ArchiveInfo{
		Format:      format,
		Compression: getCompressionType(format),
		EntryCount:  0,
		TotalSize:   0,
		Checksums:   make(map[string]string),
	}

	// Discover source files using pathfinder
	filesToArchive, discoverErr := discoverSourceFiles(sources, opts)
	if discoverErr != nil {
		err = discoverErr
		return nil, err
	}

	// Create archive based on format
	switch format {
	case ArchiveFormatTAR:
		err = createTar(output, filesToArchive, opts, info)
	case ArchiveFormatTARGZ:
		err = createTarGz(output, filesToArchive, opts, info)
	case ArchiveFormatZIP:
		err = createZip(output, filesToArchive, opts, info)
	case ArchiveFormatGZIP:
		err = createGzip(output, filesToArchive, opts, info)
	default:
		err = newError(ErrCodeInvalidFormat, "unsupported archive format", OperationCreate, output, nil)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// Get compressed size
	if fileInfo, statErr := os.Stat(output); statErr == nil {
		info.CompressedSize = fileInfo.Size()
		info.CompressionRatio = calculateCompressionRatio(info.TotalSize, info.CompressedSize)
	}

	// Generate archive checksum using fulhash
	outFile, openErr := os.Open(output)
	if openErr == nil {
		defer func() { _ = outFile.Close() }()

		// Map checksum algorithm to fulhash Algorithm
		// Note: fulhash currently supports SHA256 and XXH3_128
		// If unsupported algorithm requested, use SHA256 and update the label
		var algorithm fulhash.Algorithm
		actualAlgorithm := opts.ChecksumAlgorithm

		switch opts.ChecksumAlgorithm {
		case "sha256":
			algorithm = fulhash.SHA256
		case "xxh3-128":
			algorithm = fulhash.XXH3_128
		case "sha512", "sha1", "md5":
			// Unsupported by fulhash - fallback to SHA256 and update label
			algorithm = fulhash.SHA256
			actualAlgorithm = "sha256"
		default:
			// Unknown algorithm - fallback to SHA256
			algorithm = fulhash.SHA256
			actualAlgorithm = "sha256"
		}

		if digest, hashErr := fulhash.HashReader(outFile, fulhash.WithAlgorithm(algorithm)); hashErr == nil {
			// Store under the ACTUAL algorithm used, not the requested one
			info.Checksums[actualAlgorithm] = fulhash.FormatDigest(digest)
			info.HasChecksums = true
			info.ChecksumAlgorithm = actualAlgorithm
		}
	}

	// Set created timestamp
	now := time.Now()
	info.Created = &now

	return info, nil
}

// discoverSourceFiles uses pathfinder to discover files to archive.
func discoverSourceFiles(sources []string, opts *CreateOptions) ([]string, error) {
	var allFiles []string
	seen := make(map[string]bool) // Deduplicate files

	ctx := context.Background()
	finder := pathfinder.NewFinder()

	for _, source := range sources {
		// Check if source exists
		sourceInfo, err := os.Stat(source)
		if err != nil {
			return nil, newErrorf(ErrCodeCorruptArchive, OperationCreate, source, err,
				"source not found: %v", err)
		}

		if sourceInfo.IsDir() {
			// Use pathfinder to discover files in directory
			// Pathfinder requires Include patterns - default to all files
			includePatterns := opts.IncludePatterns
			if len(includePatterns) == 0 {
				includePatterns = []string{"**/*"}
			}

			query := pathfinder.FindQuery{
				Root:           source,
				Include:        includePatterns,
				Exclude:        opts.ExcludePatterns,
				FollowSymlinks: opts.FollowSymlinks,
			}

			results, findErr := finder.FindFiles(ctx, query)
			if findErr != nil {
				return nil, newErrorf(ErrCodeCorruptArchive, OperationCreate, source, findErr,
					"failed to discover files: %v", findErr)
			}

			for _, result := range results {
				// Use SourcePath which is the actual filesystem path
				if !seen[result.SourcePath] {
					allFiles = append(allFiles, result.SourcePath)
					seen[result.SourcePath] = true
				}
			}
		} else {
			// Single file - check against patterns
			normalizedPath := filepath.ToSlash(source)
			if shouldIncludeFile(normalizedPath, opts.IncludePatterns, opts.ExcludePatterns) {
				if !seen[source] {
					allFiles = append(allFiles, source)
					seen[source] = true
				}
			}
		}
	}

	return allFiles, nil
}

// shouldIncludeFile checks if a file should be included based on patterns.
func shouldIncludeFile(normalizedPath string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, pattern := range excludePatterns {
		if matched, _ := doublestar.Match(pattern, normalizedPath); matched {
			return false
		}
	}

	// If no include patterns, include by default
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

// createTar creates an uncompressed tar archive.
func createTar(output string, files []string, opts *CreateOptions, info *ArchiveInfo) error {
	outFile, err := os.Create(output)
	if err != nil {
		return newErrorf(ErrCodeFileExists, OperationCreate, output, err,
			"failed to create tar archive: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	tw := tar.NewWriter(outFile)
	defer func() { _ = tw.Close() }()

	return writeTarEntries(tw, files, opts, info, output)
}

// createTarGz creates a tar.gz archive.
func createTarGz(output string, files []string, opts *CreateOptions, info *ArchiveInfo) error {
	outFile, err := os.Create(output)
	if err != nil {
		return newErrorf(ErrCodeFileExists, OperationCreate, output, err,
			"failed to create tar.gz archive: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	// Create gzip writer with compression level
	gw, err := gzip.NewWriterLevel(outFile, opts.CompressionLevel)
	if err != nil {
		return newErrorf(ErrCodeUnsupportedCompression, OperationCreate, output, err,
			"failed to create gzip writer: %v", err)
	}
	defer func() { _ = gw.Close() }()

	tw := tar.NewWriter(gw)
	defer func() { _ = tw.Close() }()

	return writeTarEntries(tw, files, opts, info, output)
}

// writeTarEntries writes files to a tar writer.
func writeTarEntries(tw *tar.Writer, files []string, opts *CreateOptions, info *ArchiveInfo, archivePath string) error {
	for _, filePath := range files {
		fileInfo, err := os.Lstat(filePath)
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
				"failed to stat file %s: %v", filePath, err)
		}

		// Handle symlinks
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			if !opts.FollowSymlinks {
				// Add symlink as-is
				linkTarget, err := os.Readlink(filePath)
				if err != nil {
					return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
						"failed to read symlink %s: %v", filePath, err)
				}

				header := &tar.Header{
					Name:     filePath,
					Linkname: linkTarget,
					Typeflag: tar.TypeSymlink,
					Mode:     int64(fileInfo.Mode()),
					ModTime:  fileInfo.ModTime(),
				}

				if !*opts.PreservePermissions {
					header.Mode = 0777
				}

				if err := tw.WriteHeader(header); err != nil {
					return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
						"failed to write symlink header: %v", err)
				}

				info.EntryCount++
				continue
			}

			// Follow symlink - get target file info
			targetInfo, err := os.Stat(filePath)
			if err != nil {
				return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
					"failed to stat symlink target %s: %v", filePath, err)
			}
			fileInfo = targetInfo
		}

		// Handle directories
		if fileInfo.IsDir() {
			header := &tar.Header{
				Name:     filePath + "/",
				Typeflag: tar.TypeDir,
				Mode:     int64(fileInfo.Mode()),
				ModTime:  fileInfo.ModTime(),
			}

			if !*opts.PreservePermissions {
				header.Mode = 0755
			}

			if err := tw.WriteHeader(header); err != nil {
				return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
					"failed to write directory header: %v", err)
			}

			info.EntryCount++
			continue
		}

		// Handle regular files
		file, err := os.Open(filePath)
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
				"failed to open file %s: %v", filePath, err)
		}

		header := &tar.Header{
			Name:    filePath,
			Size:    fileInfo.Size(),
			Mode:    int64(fileInfo.Mode()),
			ModTime: fileInfo.ModTime(),
		}

		if !*opts.PreservePermissions {
			header.Mode = 0644
		}

		if err := tw.WriteHeader(header); err != nil {
			_ = file.Close()
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
				"failed to write file header: %v", err)
		}

		bytesWritten, err := io.Copy(tw, file)
		_ = file.Close()

		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, archivePath, err,
				"failed to write file data: %v", err)
		}

		info.EntryCount++
		info.TotalSize += bytesWritten
	}

	return nil
}

// createZip creates a zip archive.
func createZip(output string, files []string, opts *CreateOptions, info *ArchiveInfo) error {
	outFile, err := os.Create(output)
	if err != nil {
		return newErrorf(ErrCodeFileExists, OperationCreate, output, err,
			"failed to create zip archive: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	zw := zip.NewWriter(outFile)
	defer func() { _ = zw.Close() }()

	// Set compression level
	zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, opts.CompressionLevel)
	})

	for _, filePath := range files {
		fileInfo, err := os.Lstat(filePath)
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
				"failed to stat file %s: %v", filePath, err)
		}

		// Handle symlinks
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			if opts.FollowSymlinks {
				targetInfo, err := os.Stat(filePath)
				if err != nil {
					return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
						"failed to stat symlink target %s: %v", filePath, err)
				}
				fileInfo = targetInfo
			} else {
				// ZIP doesn't have native symlink support - skip
				continue
			}
		}

		// Handle directories
		if fileInfo.IsDir() {
			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
					"failed to create zip header: %v", err)
			}
			header.Name = filePath + "/"
			header.Method = zip.Deflate

			if _, err := zw.CreateHeader(header); err != nil {
				return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
					"failed to write directory header: %v", err)
			}

			info.EntryCount++
			continue
		}

		// Handle regular files
		file, err := os.Open(filePath)
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
				"failed to open file %s: %v", filePath, err)
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			_ = file.Close()
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
				"failed to create zip header: %v", err)
		}
		header.Name = filePath
		header.Method = zip.Deflate

		if !*opts.PreservePermissions {
			header.SetMode(0644)
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			_ = file.Close()
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
				"failed to create zip entry: %v", err)
		}

		bytesWritten, err := io.Copy(writer, file)
		_ = file.Close()

		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
				"failed to write file data: %v", err)
		}

		info.EntryCount++
		info.TotalSize += bytesWritten
	}

	return nil
}

// createGzip creates a gzip file (single file only).
func createGzip(output string, files []string, opts *CreateOptions, info *ArchiveInfo) error {
	// GZIP format only supports single file
	if len(files) == 0 {
		return newError(ErrCodeInvalidFormat, "no files to compress", OperationCreate, output, nil)
	}
	if len(files) > 1 {
		return newError(ErrCodeInvalidFormat, "gzip format only supports single file compression", OperationCreate, output, nil)
	}

	inputPath := files[0]

	// Verify it's a file (not directory)
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
			"failed to stat input file: %v", err)
	}
	if fileInfo.IsDir() {
		return newError(ErrCodeInvalidFormat, "cannot compress directory with gzip format (use tar.gz)", OperationCreate, output, nil)
	}

	// Open input file
	inFile, err := os.Open(inputPath)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
			"failed to open input file: %v", err)
	}
	defer func() { _ = inFile.Close() }()

	// Create output file
	outFile, err := os.Create(output)
	if err != nil {
		return newErrorf(ErrCodeFileExists, OperationCreate, output, err,
			"failed to create gzip file: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	// Create gzip writer with compression level
	gw, err := gzip.NewWriterLevel(outFile, opts.CompressionLevel)
	if err != nil {
		return newErrorf(ErrCodeUnsupportedCompression, OperationCreate, output, err,
			"failed to create gzip writer: %v", err)
	}
	defer func() { _ = gw.Close() }()

	// Set gzip header name to original filename
	gw.Name = filepath.Base(inputPath)

	// Compress file
	bytesWritten, err := io.Copy(gw, inFile)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationCreate, output, err,
			"failed to compress file: %v", err)
	}

	info.EntryCount = 1
	info.TotalSize = bytesWritten

	return nil
}
