package fulpack_test

import (
	"path/filepath"
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
