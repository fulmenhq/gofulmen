package fulpack

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"time"
)

// infoImpl implements the Info operation.
func infoImpl(archive string) (*ArchiveInfo, error) {
	start := time.Now()
	var err error
	var info *ArchiveInfo

	defer func() {
		duration := time.Since(start)
		var format ArchiveFormat
		var entryCount int
		var bytesProcessed int64
		if info != nil {
			format = info.Format
			entryCount = info.EntryCount
			bytesProcessed = info.TotalSize
		}
		emitOperationMetrics(OperationInfo, format, duration, entryCount, bytesProcessed, err)
	}()
	// Detect format
	format := detectFormat(archive)
	if format == "" {
		err = newError(ErrCodeInvalidFormat, "could not detect archive format", OperationInfo, archive, nil)
		return nil, err
	}

	// Get file info for compressed size
	fileInfo, statErr := os.Stat(archive)
	if statErr != nil {
		err = newErrorf(ErrCodeCorruptArchive, OperationInfo, archive, statErr, "failed to stat archive: %v", statErr)
		return nil, err
	}

	info = &ArchiveInfo{
		Format:         format,
		CompressedSize: fileInfo.Size(),
		Compression:    getCompressionType(format),
		Checksums:      make(map[string]string),
	}

	// Read archive metadata based on format
	switch format {
	case ArchiveFormatTAR:
		if readErr := readTarInfo(archive, info); readErr != nil {
			err = readErr
			return nil, err
		}
	case ArchiveFormatTARGZ:
		if readErr := readTarGzInfo(archive, info); readErr != nil {
			err = readErr
			return nil, err
		}
	case ArchiveFormatZIP:
		if readErr := readZipInfo(archive, info); readErr != nil {
			err = readErr
			return nil, err
		}
	case ArchiveFormatGZIP:
		if readErr := readGzipInfo(archive, info); readErr != nil {
			err = readErr
			return nil, err
		}
	default:
		err = newError(ErrCodeInvalidFormat, "unsupported archive format", OperationInfo, archive, nil)
		return nil, err
	}

	// Calculate compression ratio
	if info.CompressedSize > 0 {
		info.CompressionRatio = calculateCompressionRatio(info.TotalSize, info.CompressedSize)
	}

	return info, nil
}

// readTarInfo reads metadata from uncompressed tar archive.
func readTarInfo(path string, info *ArchiveInfo) error {
	f, err := os.Open(path)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to open tar archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	var totalSize int64
	var entryCount int

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to read tar header: %v", err)
		}

		entryCount++
		totalSize += header.Size
	}

	info.EntryCount = entryCount
	info.TotalSize = totalSize

	return nil
}

// readTarGzInfo reads metadata from tar.gz archive.
func readTarGzInfo(path string, info *ArchiveInfo) error {
	f, err := os.Open(path)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to open tar.gz archive: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	var totalSize int64
	var entryCount int

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to read tar header: %v", err)
		}

		entryCount++
		totalSize += header.Size
	}

	info.EntryCount = entryCount
	info.TotalSize = totalSize

	return nil
}

// readZipInfo reads metadata from zip archive.
func readZipInfo(path string, info *ArchiveInfo) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to open zip archive: %v", err)
	}
	defer func() { _ = zr.Close() }()

	var totalSize int64
	for _, f := range zr.File {
		totalSize += int64(f.UncompressedSize64)
	}

	info.EntryCount = len(zr.File)
	info.TotalSize = totalSize

	return nil
}

// readGzipInfo reads metadata from gzip file.
func readGzipInfo(path string, info *ArchiveInfo) error {
	f, err := os.Open(path)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to open gzip file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	// For gzip (single file), we need to decompress to get uncompressed size
	totalSize, err := io.Copy(io.Discard, gr)
	if err != nil {
		return newErrorf(ErrCodeCorruptArchive, OperationInfo, path, err, "failed to read gzip content: %v", err)
	}

	info.EntryCount = 1 // Single file
	info.TotalSize = totalSize

	return nil
}

// getCompressionType returns compression type string for format.
func getCompressionType(format ArchiveFormat) string {
	switch format {
	case ArchiveFormatTAR:
		return "none"
	case ArchiveFormatTARGZ:
		return "gzip"
	case ArchiveFormatZIP:
		return "deflate"
	case ArchiveFormatGZIP:
		return "gzip"
	default:
		return "unknown"
	}
}
