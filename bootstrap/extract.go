package bootstrap

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxExtractionSize = 1024 * 1024 * 1024

func ExtractArchive(archivePath, destDir string) error {
	ext := strings.ToLower(filepath.Ext(archivePath))

	if ext == ".gz" && strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	}

	if ext == ".zip" {
		return extractZip(archivePath, destDir)
	}

	return fmt.Errorf("unsupported archive format: %s (supported: .tar.gz, .zip)", ext)
}

func extractTarGz(archivePath, destDir string) error {
	// #nosec G304 -- archivePath is validated by caller and used in controlled bootstrap context
	f, err := os.Open(archivePath)
	if err != nil {
		return &ExtractionError{Archive: archivePath, Err: err}
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return &ExtractionError{Archive: archivePath, Err: err}
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var totalSize int64

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			continue
		}

		// #nosec G305 -- path traversal prevented by validateExtractedPath check below
		target := filepath.Join(destDir, header.Name)
		if err := validateExtractedPath(target, destDir); err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0750); err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}

		case tar.TypeReg:
			totalSize += header.Size
			if totalSize > maxExtractionSize {
				return &ExtractionError{
					Archive: archivePath,
					Err:     fmt.Errorf("extraction size exceeds limit (%d bytes)", maxExtractionSize),
				}
			}

			if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}
			// #nosec G115 G304
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}

			// #nosec G110 -- decompression bomb protected by totalSize check above
			if _, err := io.Copy(outFile, tr); err != nil {
				// #nosec G104 -- Close() error ignored during error cleanup
				outFile.Close()
				return &ExtractionError{Archive: archivePath, Err: err}
			}
			// #nosec G104 -- defer Close() error is commonly ignored in Go
			outFile.Close()
		}
	}

	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return &ExtractionError{Archive: archivePath, Err: err}
	}
	defer r.Close()

	var totalSize int64

	for _, f := range r.File {
		// #nosec G305 -- path traversal prevented by validateExtractedPath check below
		target := filepath.Join(destDir, f.Name)
		if err := validateExtractedPath(target, destDir); err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0750); err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}
			continue
		}

		// #nosec G115 -- bounds checked conversion from uint64 to int64 with overflow validation
		fileSize := int64(f.UncompressedSize64)
		if fileSize < 0 {
			return &ExtractionError{
				Archive: archivePath,
				Err:     fmt.Errorf("invalid file size: %d", f.UncompressedSize64),
			}
		}
		totalSize += fileSize
		if totalSize > maxExtractionSize {
			return &ExtractionError{
				Archive: archivePath,
				Err:     fmt.Errorf("extraction size exceeds limit (%d bytes)", maxExtractionSize),
			}
		}

		if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		rc, err := f.Open()
		if err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		// #nosec G304 -- target path is validated above to prevent traversal
		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, f.Mode())
		if err != nil {
			// #nosec G104 -- Close() error ignored during error cleanup
			rc.Close()
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		// #nosec G110 -- decompression bomb protected by totalSize check above
		_, err = io.Copy(outFile, rc)
		// #nosec G104 -- defer Close() error is commonly ignored in Go
		rc.Close()
		// #nosec G104 -- defer Close() error is commonly ignored in Go
		outFile.Close()

		if err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}
	}

	return nil
}

func validatePath(p string) error {
	if strings.Contains(p, "..") {
		return &UnsafePath{Path: p}
	}

	if filepath.IsAbs(p) {
		return &UnsafePath{Path: p}
	}

	return nil
}

func validateExtractedPath(extractedPath, destDir string) error {
	// Clean the extracted path to resolve any .. or . components
	cleaned := filepath.Clean(extractedPath)

	// Clean the destination directory
	cleanDest := filepath.Clean(destDir)

	// Ensure the cleaned path is within the destination directory
	rel, err := filepath.Rel(cleanDest, cleaned)
	if err != nil {
		return &UnsafePath{Path: extractedPath}
	}

	// Check if the relative path tries to escape (starts with ..)
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return &UnsafePath{Path: extractedPath}
	}

	return nil
}
