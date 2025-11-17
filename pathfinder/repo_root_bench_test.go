package pathfinder

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkFindRepositoryRoot_ImmediateMatch benchmarks finding a marker in the current directory
func BenchmarkFindRepositoryRoot_ImmediateMatch(b *testing.B) {
	// Create temp directory with marker
	tempDir := b.TempDir()
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		b.Fatal(err)
	}

	// Set boundary to parent of tempDir to allow finding within tempDir
	boundary := filepath.Dir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(tempDir, GoModMarkers, WithBoundary(boundary))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_ThreeLevelsUp benchmarks typical case (3 directories up)
func BenchmarkFindRepositoryRoot_ThreeLevelsUp(b *testing.B) {
	// Create nested structure: tempDir/a/b/c (marker at tempDir)
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(filepath.Dir(tempDir)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_FiveLevelsUp benchmarks deeper nesting (5 directories up)
func BenchmarkFindRepositoryRoot_FiveLevelsUp(b *testing.B) {
	// Create nested structure: tempDir/a/b/c/d/e (marker at tempDir)
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c", "d", "e")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(filepath.Dir(tempDir)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_MaxDepth benchmarks hitting max depth limit
func BenchmarkFindRepositoryRoot_MaxDepth(b *testing.B) {
	// Create deeply nested structure without marker
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will hit max depth and fail
		_, _ = FindRepositoryRoot(nestedPath, GoModMarkers, WithMaxDepth(10), WithBoundary(filepath.Dir(tempDir)))
	}
}

// BenchmarkFindRepositoryRoot_MultipleMarkers benchmarks with multiple marker types
func BenchmarkFindRepositoryRoot_MultipleMarkers(b *testing.B) {
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}

	// Create .git directory as marker
	gitPath := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitPath, 0755); err != nil {
		b.Fatal(err)
	}

	markers := []string{".git", "go.mod", "package.json"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(nestedPath, markers, WithBoundary(filepath.Dir(tempDir)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_RealRepository benchmarks on actual repository
func BenchmarkFindRepositoryRoot_RealRepository(b *testing.B) {
	// This benchmark uses the actual gofulmen repository
	// Run from pathfinder directory, find .git at repo root
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(".", GitMarkers)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_WithCustomBoundary benchmarks with explicit boundary
func BenchmarkFindRepositoryRoot_WithCustomBoundary(b *testing.B) {
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		b.Fatal(err)
	}

	boundary := filepath.Dir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(boundary))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindRepositoryRoot_FileVsDirectory benchmarks directory marker vs file marker
func BenchmarkFindRepositoryRoot_FileVsDirectory(b *testing.B) {
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}

	b.Run("FileMarker", func(b *testing.B) {
		markerPath := filepath.Join(tempDir, "go.mod")
		if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
			b.Fatal(err)
		}
		defer func() {
			_ = os.Remove(markerPath) // Cleanup, error not critical in benchmark
		}()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(filepath.Dir(tempDir)))
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("DirectoryMarker", func(b *testing.B) {
		gitPath := filepath.Join(tempDir, ".git")
		if err := os.Mkdir(gitPath, 0755); err != nil {
			b.Fatal(err)
		}
		defer func() {
			_ = os.RemoveAll(gitPath) // Cleanup, error not critical in benchmark
		}()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := FindRepositoryRoot(nestedPath, GitMarkers, WithBoundary(filepath.Dir(tempDir)))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkFindRepositoryRoot_Parallel benchmarks concurrent calls
func BenchmarkFindRepositoryRoot_Parallel(b *testing.B) {
	tempDir := b.TempDir()
	nestedPath := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		b.Fatal(err)
	}
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(filepath.Dir(tempDir)))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
