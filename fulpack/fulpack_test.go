package fulpack_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/fulpack"
)

// Fixture paths relative to gofulmen root
const fixturesDir = "../config/crucible-go/library/fulpack/fixtures"

func TestInfo_BasicTar(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")

	info, err := fulpack.Info(archive)
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatTAR {
		t.Errorf("Expected format TAR, got %s", info.Format)
	}

	if info.Compression != "none" {
		t.Errorf("Expected compression 'none', got %s", info.Compression)
	}

	if info.EntryCount == 0 {
		t.Errorf("Expected non-zero entry count")
	}

	if info.TotalSize == 0 {
		t.Errorf("Expected non-zero total size")
	}

	// Note: For uncompressed tar, compression ratio compares actual data size to tar file size
	// (which includes 512-byte block padding). This will be < 1.0, not 1.0 as might be expected.
	// The spec notes this edge case for tar format due to block padding overhead.
	if info.CompressionRatio <= 0 {
		t.Errorf("Expected positive compression ratio, got %.2f", info.CompressionRatio)
	}

	t.Logf("Archive info: format=%s, entries=%d, size=%d bytes, ratio=%.2fx",
		info.Format, info.EntryCount, info.TotalSize, info.CompressionRatio)
}

func TestInfo_BasicTarGz(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar.gz")

	info, err := fulpack.Info(archive)
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatTARGZ {
		t.Errorf("Expected format TAR.GZ, got %s", info.Format)
	}

	if info.Compression != "gzip" {
		t.Errorf("Expected compression 'gzip', got %s", info.Compression)
	}

	if info.EntryCount == 0 {
		t.Errorf("Expected non-zero entry count")
	}

	// Compressed archive should have ratio > 1.0
	if info.CompressionRatio <= 1.0 {
		t.Errorf("Expected compression ratio > 1.0 for tar.gz, got %.2f", info.CompressionRatio)
	}

	t.Logf("Archive info: format=%s, entries=%d, compressed=%d bytes, uncompressed=%d bytes, ratio=%.2fx",
		info.Format, info.EntryCount, info.CompressedSize, info.TotalSize, info.CompressionRatio)
}

func TestInfo_NestedZip(t *testing.T) {
	archive := filepath.Join(fixturesDir, "nested.zip")

	info, err := fulpack.Info(archive)
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatZIP {
		t.Errorf("Expected format ZIP, got %s", info.Format)
	}

	if info.Compression != "deflate" {
		t.Errorf("Expected compression 'deflate', got %s", info.Compression)
	}

	if info.EntryCount == 0 {
		t.Errorf("Expected non-zero entry count")
	}

	t.Logf("Archive info: format=%s, entries=%d, compressed=%d bytes, uncompressed=%d bytes, ratio=%.2fx",
		info.Format, info.EntryCount, info.CompressedSize, info.TotalSize, info.CompressionRatio)
}

func TestScan_BasicTar(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")

	entries, err := fulpack.Scan(archive, nil)
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Errorf("Expected non-zero entries")
	}

	// Verify we have files and directories
	hasFile := false
	hasDir := false
	for _, entry := range entries {
		if entry.Type == fulpack.EntryTypeFile {
			hasFile = true
		}
		if entry.Type == fulpack.EntryTypeDirectory {
			hasDir = true
		}
	}

	if !hasFile {
		t.Errorf("Expected at least one file entry")
	}

	t.Logf("Scanned %d entries (hasFile=%v, hasDir=%v)", len(entries), hasFile, hasDir)
	for i, entry := range entries {
		t.Logf("  [%d] %s (%s, %d bytes)", i, entry.Path, entry.Type, entry.Size)
	}
}

func TestScan_NestedZip(t *testing.T) {
	archive := filepath.Join(fixturesDir, "nested.zip")

	entries, err := fulpack.Scan(archive, nil)
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Errorf("Expected non-zero entries")
	}

	// nested.zip should have 3-level directory nesting
	// Count path components (e.g., "level1/level2/level3/file.txt" has 4 components)
	maxDepth := 0
	for _, entry := range entries {
		parts := 1
		for _, c := range filepath.ToSlash(entry.Path) {
			if c == '/' {
				parts++
			}
		}
		if parts > maxDepth {
			maxDepth = parts
		}
	}

	// Should have at least 4 components deep (level1/level2/level3/file)
	if maxDepth < 4 {
		t.Errorf("Expected at least 4-component paths, got %d", maxDepth)
	}

	t.Logf("Scanned %d entries, max depth=%d", len(entries), maxDepth)
	for i, entry := range entries {
		t.Logf("  [%d] %s (%s, %d bytes)", i, entry.Path, entry.Type, entry.Size)
	}
}

func TestScan_WithFilters(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar.gz")

	// Scan only files (no directories)
	entries, err := fulpack.Scan(archive, &fulpack.ScanOptions{
		EntryTypes: []fulpack.EntryType{fulpack.EntryTypeFile},
	})
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	// Verify all entries are files
	for _, entry := range entries {
		if entry.Type != fulpack.EntryTypeFile {
			t.Errorf("Expected only file entries, got %s", entry.Type)
		}
	}

	t.Logf("Scanned %d file entries (directories filtered out)", len(entries))
}

func TestScan_WithGlobPatterns(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")

	// Test include pattern - only .txt files
	entries, err := fulpack.Scan(archive, &fulpack.ScanOptions{
		IncludePatterns: []string{"**/*.txt"},
	})
	if err != nil {
		t.Fatalf("Scan() with include pattern failed: %v", err)
	}

	// Verify all entries match the pattern
	for _, entry := range entries {
		matched, _ := filepath.Match("*.txt", filepath.Base(entry.Path))
		if !matched {
			t.Errorf("Entry %s doesn't match *.txt pattern", entry.Path)
		}
	}

	t.Logf("Scanned %d .txt files with include pattern", len(entries))

	// Test exclude pattern - exclude .txt files
	allEntries, err := fulpack.Scan(archive, &fulpack.ScanOptions{
		EntryTypes: []fulpack.EntryType{fulpack.EntryTypeFile},
	})
	if err != nil {
		t.Fatalf("Scan() all files failed: %v", err)
	}

	excludeEntries, err := fulpack.Scan(archive, &fulpack.ScanOptions{
		EntryTypes:      []fulpack.EntryType{fulpack.EntryTypeFile},
		ExcludePatterns: []string{"**/*.txt"},
	})
	if err != nil {
		t.Fatalf("Scan() with exclude pattern failed: %v", err)
	}

	if len(excludeEntries) >= len(allEntries) {
		t.Errorf("Expected fewer entries with exclude pattern, got %d vs %d", len(excludeEntries), len(allEntries))
	}

	// Verify no .txt files in excluded results
	for _, entry := range excludeEntries {
		if match, _ := filepath.Match("*.txt", filepath.Base(entry.Path)); match {
			t.Errorf("Entry %s shouldn't be in excluded results (matches *.txt)", entry.Path)
		}
	}

	t.Logf("Excluded .txt files: %d total -> %d after exclusion", len(allEntries), len(excludeEntries))
}

// ========================================
// Create Operation Tests
// ========================================

func TestCreate_BasicTar(t *testing.T) {
	// Create temp directory for test output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.tar")

	// Create test files to archive
	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive
	info, err := fulpack.Create(
		[]string{testFile},
		outputPath,
		fulpack.ArchiveFormatTAR,
		nil,
	)

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatTAR {
		t.Errorf("Expected format TAR, got %s", info.Format)
	}

	if info.EntryCount != 1 {
		t.Errorf("Expected 1 entry, got %d", info.EntryCount)
	}

	if info.TotalSize != 11 { // "hello world" is 11 bytes
		t.Errorf("Expected total size 11, got %d", info.TotalSize)
	}

	// Verify archive was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Archive file was not created")
	}

	t.Logf("Created TAR archive: %d entries, %d bytes", info.EntryCount, info.TotalSize)
}

func TestCreate_BasicTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.tar.gz")

	// Create test files
	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world compressed"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create compressed archive
	info, err := fulpack.Create(
		[]string{testFile},
		outputPath,
		fulpack.ArchiveFormatTARGZ,
		&fulpack.CreateOptions{
			CompressionLevel: 9,
		},
	)

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatTARGZ {
		t.Errorf("Expected format TAR.GZ, got %s", info.Format)
	}

	if info.Compression != "gzip" {
		t.Errorf("Expected compression 'gzip', got %s", info.Compression)
	}

	// Verify compression ratio exists (may be < 1.0 for very small files due to overhead)
	if info.CompressionRatio <= 0 {
		t.Errorf("Expected positive compression ratio, got %.2f", info.CompressionRatio)
	}

	t.Logf("Created TAR.GZ archive: %d entries, compressed %.2fx", info.EntryCount, info.CompressionRatio)
}

func TestCreate_BasicZip(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.zip")

	// Create test files
	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	for i := 0; i < 3; i++ {
		testFile := filepath.Join(testDir, "file"+string(rune('A'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("content "+string(rune('A'+i))), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create ZIP archive
	info, err := fulpack.Create(
		[]string{testDir},
		outputPath,
		fulpack.ArchiveFormatZIP,
		nil,
	)

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if info.Format != fulpack.ArchiveFormatZIP {
		t.Errorf("Expected format ZIP, got %s", info.Format)
	}

	// Should have directory + 3 files = 4 entries
	if info.EntryCount < 3 {
		t.Errorf("Expected at least 3 entries, got %d", info.EntryCount)
	}

	t.Logf("Created ZIP archive: %d entries, %d bytes", info.EntryCount, info.TotalSize)
}

func TestCreate_WithPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "filtered.tar")

	// Create test files with different extensions
	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	files := []string{"file1.txt", "file2.md", "file3.txt", "file4.log"}
	for _, name := range files {
		path := filepath.Join(testDir, name)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create archive with only .txt files
	info, err := fulpack.Create(
		[]string{testDir},
		outputPath,
		fulpack.ArchiveFormatTAR,
		&fulpack.CreateOptions{
			IncludePatterns: []string{"**/*.txt"},
		},
	)

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Should have directory + 2 .txt files
	if info.EntryCount < 2 {
		t.Errorf("Expected at least 2 .txt files, got %d entries", info.EntryCount)
	}

	t.Logf("Created filtered archive: %d entries (only .txt files)", info.EntryCount)
}

func TestCreate_WithChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "checksummed.tar.gz")

	// Create test file
	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "data.bin")
	if err := os.WriteFile(testFile, []byte("important data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archive with checksum
	info, err := fulpack.Create(
		[]string{testFile},
		outputPath,
		fulpack.ArchiveFormatTARGZ,
		&fulpack.CreateOptions{
			ChecksumAlgorithm: "sha256",
		},
	)

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !info.HasChecksums {
		t.Errorf("Expected archive to have checksums")
	}

	if info.ChecksumAlgorithm != "sha256" {
		t.Errorf("Expected algorithm 'sha256', got %s", info.ChecksumAlgorithm)
	}

	if _, exists := info.Checksums["sha256"]; !exists {
		t.Errorf("Expected SHA256 checksum in result")
	}

	t.Logf("Created archive with checksums: algorithm=%s, digest=%s",
		info.ChecksumAlgorithm, info.Checksums["sha256"])
}

// ========================================
// Extract Operation Tests
// ========================================

func TestExtract_BasicTar(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")
	destDir := t.TempDir()

	result, err := fulpack.Extract(archive, destDir, nil)
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	if result.ExtractedCount == 0 {
		t.Errorf("Expected extracted files, got 0")
	}

	if result.ErrorCount > 0 {
		t.Errorf("Expected no errors, got %d: %v", result.ErrorCount, result.Errors)
	}

	t.Logf("Extracted %d files, %d bytes", result.ExtractedCount, result.BytesWritten)
}

func TestExtract_BasicTarGz(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar.gz")
	destDir := t.TempDir()

	result, err := fulpack.Extract(archive, destDir, nil)
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	if result.ExtractedCount == 0 {
		t.Errorf("Expected extracted files, got 0")
	}

	if result.BytesWritten == 0 {
		t.Errorf("Expected bytes written, got 0")
	}

	t.Logf("Extracted %d files, %d bytes written", result.ExtractedCount, result.BytesWritten)
}

func TestExtract_WithPatterns(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")
	destDir := t.TempDir()

	// Extract only .txt files
	result, err := fulpack.Extract(archive, destDir, &fulpack.ExtractOptions{
		IncludePatterns: []string{"**/*.txt"},
	})
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// Should have extracted some files and skipped others
	totalProcessed := result.ExtractedCount + result.SkippedCount
	if totalProcessed == 0 {
		t.Errorf("Expected to process some entries")
	}

	t.Logf("Extracted %d .txt files, skipped %d non-.txt files",
		result.ExtractedCount, result.SkippedCount)
}

func TestExtract_PathTraversal(t *testing.T) {
	archive := filepath.Join(fixturesDir, "pathological.tar.gz")
	destDir := t.TempDir()

	result, err := fulpack.Extract(archive, destDir, nil)

	// Should complete (not fatal error), but report extraction errors
	if err != nil {
		t.Logf("Extract returned error (acceptable for pathological archive): %v", err)
	}

	// Should have detected security issues
	if result.ErrorCount == 0 {
		t.Errorf("Expected security errors for pathological archive, got 0")
	}

	// Verify error codes are security-related
	hasSecurityError := false
	for _, extractErr := range result.Errors {
		if extractErr.Code == "PATH_TRAVERSAL" || extractErr.Code == "SYMLINK_ESCAPE" {
			hasSecurityError = true
			break
		}
	}

	if !hasSecurityError {
		t.Errorf("Expected PATH_TRAVERSAL or SYMLINK_ESCAPE errors, got: %v", result.Errors)
	}

	t.Logf("Pathological archive: %d errors detected (security working)", result.ErrorCount)
}

func TestExtract_OverwritePolicy(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar.gz")
	destDir := t.TempDir()

	// First extraction
	result1, err := fulpack.Extract(archive, destDir, nil)
	if err != nil {
		t.Fatalf("First extract failed: %v", err)
	}

	originalCount := result1.ExtractedCount

	// Second extraction with skip policy
	result2, err := fulpack.Extract(archive, destDir, &fulpack.ExtractOptions{
		Overwrite: "skip",
	})
	if err != nil {
		t.Fatalf("Second extract failed: %v", err)
	}

	// Should have skipped files that already exist
	if result2.SkippedCount == 0 {
		t.Errorf("Expected skipped files on re-extraction, got 0")
	}

	t.Logf("First extraction: %d files, Second (skip): %d skipped",
		originalCount, result2.SkippedCount)
}

// ========================================
// Verify Operation Tests
// ========================================

func TestVerify_ValidArchive(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar.gz")

	result, err := fulpack.Verify(archive, nil)
	if err != nil {
		t.Fatalf("Verify() failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected archive to be valid, got invalid with errors: %v", result.Errors)
	}

	if result.EntryCount == 0 {
		t.Errorf("Expected non-zero entry count")
	}

	if len(result.ChecksPerformed) == 0 {
		t.Errorf("Expected checks to be performed")
	}

	t.Logf("Verified archive: %d entries, %d checks performed, valid=%v",
		result.EntryCount, len(result.ChecksPerformed), result.Valid)
}

func TestVerify_PathologicalArchive(t *testing.T) {
	archive := filepath.Join(fixturesDir, "pathological.tar.gz")

	result, err := fulpack.Verify(archive, nil)
	if err != nil {
		t.Fatalf("Verify() failed: %v", err)
	}

	// Pathological archive should fail validation
	if result.Valid {
		t.Errorf("Expected pathological archive to be invalid")
	}

	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors for pathological archive")
	}

	// Check for path traversal errors
	hasPathTraversalError := false
	for _, valErr := range result.Errors {
		if valErr.Code == "PATH_TRAVERSAL" {
			hasPathTraversalError = true
			break
		}
	}

	if !hasPathTraversalError {
		t.Errorf("Expected PATH_TRAVERSAL error in validation, got: %v", result.Errors)
	}

	t.Logf("Pathological archive validation: %d errors detected", len(result.Errors))
}

func TestVerify_CorruptArchive(t *testing.T) {
	// Create a corrupt archive (truncated file)
	tmpDir := t.TempDir()
	corruptPath := filepath.Join(tmpDir, "corrupt.tar.gz")

	// Write some garbage data
	if err := os.WriteFile(corruptPath, []byte("not a valid tar.gz file"), 0644); err != nil {
		t.Fatalf("Failed to create corrupt file: %v", err)
	}

	result, err := fulpack.Verify(corruptPath, nil)
	if err != nil {
		t.Logf("Verify returned error (acceptable for corrupt archive): %v", err)
	}

	// Should report validation failure
	if result != nil && result.Valid {
		t.Errorf("Expected corrupt archive to fail validation")
	}

	if result != nil {
		t.Logf("Corrupt archive validation: valid=%v, errors=%d", result.Valid, len(result.Errors))
	}
}

func TestExtract_WithExcludePatterns(t *testing.T) {
	archive := filepath.Join(fixturesDir, "basic.tar")
	destDir := t.TempDir()

	// Extract excluding .txt files
	result, err := fulpack.Extract(archive, destDir, &fulpack.ExtractOptions{
		ExcludePatterns: []string{"**/*.txt"},
	})
	if err != nil {
		t.Fatalf("Extract() failed: %v", err)
	}

	// Verify no .txt files were extracted
	entries, _ := os.ReadDir(destDir)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".txt" {
			t.Errorf("Found .txt file that should have been excluded: %s", entry.Name())
		}
	}

	// Should have extracted some files and excluded others
	if result.ExtractedCount == 0 {
		t.Errorf("Expected some files to be extracted")
	}
	if result.SkippedCount == 0 {
		t.Errorf("Expected some .txt files to be skipped via exclude pattern")
	}

	t.Logf("Extracted %d files, excluded %d .txt files", result.ExtractedCount, result.SkippedCount)
}

func TestCreate_UnsupportedChecksumAlgorithm(t *testing.T) {
	tmpDir := t.TempDir()

	testDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "data.bin")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	for _, algorithm := range []string{"sha512", "sha1", "md5", "not-a-checksum"} {
		t.Run(algorithm, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, algorithm+".tar.gz")
			info, err := fulpack.Create(
				[]string{testFile},
				outputPath,
				fulpack.ArchiveFormatTARGZ,
				&fulpack.CreateOptions{ChecksumAlgorithm: algorithm},
			)

			if info != nil {
				t.Fatalf("Create() returned archive info for unsupported checksum %q", algorithm)
			}

			var fulpackErr *fulpack.FulpackError
			if !errors.As(err, &fulpackErr) || fulpackErr.Code != fulpack.ErrCodeInvalidOptions {
				t.Fatalf("Create() error = %v, want INVALID_OPTIONS", err)
			}

			if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
				t.Fatalf("Create() left an output artifact: stat error = %v", statErr)
			}
		})
	}
}

func TestCreate_DefaultChecksumAlgorithm(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "data.bin")
	outputPath := filepath.Join(tmpDir, "default.tar.gz")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := fulpack.Create([]string{testFile}, outputPath, fulpack.ArchiveFormatTARGZ, nil)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if info.ChecksumAlgorithm != "sha256" || info.Checksums["sha256"] == "" {
		t.Fatalf("Create() default checksum = %#v, want sha256 digest", info.Checksums)
	}
}

func TestCreate_XXH3ChecksumAlgorithm(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "data.bin")
	outputPath := filepath.Join(tmpDir, "xxh3.tar.gz")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := fulpack.Create(
		[]string{testFile}, outputPath, fulpack.ArchiveFormatTARGZ,
		&fulpack.CreateOptions{ChecksumAlgorithm: "xxh3-128"},
	)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if info.ChecksumAlgorithm != "xxh3-128" || info.Checksums["xxh3-128"] == "" {
		t.Fatalf("Create() checksum = %#v, want xxh3-128 digest", info.Checksums)
	}
}

func TestCreate_GzipUsesDefaultCompressionLevel(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "data.txt")
	data := bytes.Repeat([]byte("compression level must be stable for plain gzip\n"), 32768)
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lowPath := filepath.Join(tmpDir, "low.gz")
	highPath := filepath.Join(tmpDir, "high.gz")
	for _, tc := range []struct {
		path  string
		level int
	}{{lowPath, 1}, {highPath, 9}} {
		if _, err := fulpack.Create(
			[]string{testFile}, tc.path, fulpack.ArchiveFormatGZIP,
			&fulpack.CreateOptions{CompressionLevel: tc.level},
		); err != nil {
			t.Fatalf("Create() failed for gzip level %d: %v", tc.level, err)
		}
	}

	low, err := os.ReadFile(lowPath)
	if err != nil {
		t.Fatalf("Failed to read low-level gzip: %v", err)
	}
	high, err := os.ReadFile(highPath)
	if err != nil {
		t.Fatalf("Failed to read high-level gzip: %v", err)
	}
	if !bytes.Equal(low, high) {
		t.Fatal("plain gzip output differs across requested compression levels")
	}
}

func TestCreate_TarGzAndZipHonorCompressionLevel(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "data.txt")
	data := []byte(strings.Repeat("compressible content with enough variation for flate levels ", 32768))
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	for _, format := range []fulpack.ArchiveFormat{fulpack.ArchiveFormatTARGZ, fulpack.ArchiveFormatZIP} {
		t.Run(string(format), func(t *testing.T) {
			lowPath := filepath.Join(tmpDir, string(format)+"-low")
			highPath := filepath.Join(tmpDir, string(format)+"-high")
			low, err := fulpack.Create(
				[]string{testFile}, lowPath, format, &fulpack.CreateOptions{CompressionLevel: 1},
			)
			if err != nil {
				t.Fatalf("Create() failed for level 1: %v", err)
			}
			high, err := fulpack.Create(
				[]string{testFile}, highPath, format, &fulpack.CreateOptions{CompressionLevel: 9},
			)
			if err != nil {
				t.Fatalf("Create() failed for level 9: %v", err)
			}
			if low.CompressedSize <= high.CompressedSize {
				t.Fatalf("compression level was not honored: level 1 size %d, level 9 size %d", low.CompressedSize, high.CompressedSize)
			}
		})
	}
}
