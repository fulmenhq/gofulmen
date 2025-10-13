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

		if err := validatePath(header.Name); err != nil {
			return err
		}

		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			continue
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
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

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return &ExtractionError{Archive: archivePath, Err: err}
			}
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
		if err := validatePath(f.Name); err != nil {
			return err
		}

		target := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return &ExtractionError{Archive: archivePath, Err: err}
			}
			continue
		}

		totalSize += int64(f.UncompressedSize64)
		if totalSize > maxExtractionSize {
			return &ExtractionError{
				Archive: archivePath,
				Err:     fmt.Errorf("extraction size exceeds limit (%d bytes)", maxExtractionSize),
			}
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		rc, err := f.Open()
		if err != nil {
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return &ExtractionError{Archive: archivePath, Err: err}
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
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
