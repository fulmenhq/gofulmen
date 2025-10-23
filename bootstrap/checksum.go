package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func VerifySHA256(filePath string, expectedHex string) error {
	// #nosec G304 -- filePath is controlled path for checksum verification in bootstrap
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum verification: %w", err)
	}
	defer f.Close() //nolint:errcheck // defer Close() error is commonly ignored in Go

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to read file for checksum: %w", err)
	}

	actualHex := hex.EncodeToString(h.Sum(nil))
	if actualHex != expectedHex {
		return &ChecksumMismatchError{
			FilePath: filePath,
			Expected: expectedHex,
			Actual:   actualHex,
		}
	}

	return nil
}

func ComputeSHA256(filePath string) (string, error) {
	// #nosec G304 -- filePath is controlled path for checksum computation in bootstrap
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close() //nolint:errcheck // defer Close() error is commonly ignored in Go

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
