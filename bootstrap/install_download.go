package bootstrap

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func installDownload(tool *Tool, platform Platform) error {
	url := InterpolateURL(tool.Install.URL, platform)

	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("only HTTPS URLs are allowed, got: %s", url)
	}

	expectedChecksum, ok := tool.Install.Checksum[platform.String()]
	if !ok {
		return fmt.Errorf("no checksum found for platform %s", platform)
	}

	tempDir, err := os.MkdirTemp("", "bootstrap-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archiveName := filepath.Base(url)
	archivePath := filepath.Join(tempDir, archiveName)
	if err := downloadFile(url, archivePath); err != nil {
		return &DownloadError{URL: url, Platform: platform, Err: err}
	}

	if err := VerifySHA256(archivePath, expectedChecksum); err != nil {
		return err
	}

	extractDir := filepath.Join(tempDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := ExtractArchive(archivePath, extractDir); err != nil {
		return err
	}

	binPath, err := findBinary(extractDir, tool.Install.BinName)
	if err != nil {
		return fmt.Errorf("failed to find binary %s in extracted archive: %w", tool.Install.BinName, err)
	}

	destDir := tool.Install.Destination
	if destDir == "" {
		destDir = "./bin"
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	destPath := filepath.Join(destDir, tool.Install.BinName)

	if err := copyFile(binPath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary to %s: %w", destPath, err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return nil
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func findBinary(dir, binName string) (string, error) {
	var found string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Base(path) == binName {
			found = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if found == "" {
		return "", fmt.Errorf("binary %s not found in archive", binName)
	}

	return found, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
