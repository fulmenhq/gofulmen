package pathfinder

import (
	"context"
	"testing"
)

// BenchmarkFindFiles_SmallTree benchmarks finding files in a small directory tree
func BenchmarkFindFiles_SmallTree(b *testing.B) {
	ctx := context.Background()
	finder := NewFinder()
	query := FindQuery{
		Root:    "testdata/basic",
		Include: []string{"*"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = finder.FindFiles(ctx, query)
	}
}

// BenchmarkFindFiles_NestedTree benchmarks finding files in a nested directory tree
func BenchmarkFindFiles_NestedTree(b *testing.B) {
	ctx := context.Background()
	finder := NewFinder()
	query := FindQuery{
		Root:    "testdata/nested",
		Include: []string{"**/*.go"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = finder.FindFiles(ctx, query)
	}
}

// BenchmarkFindFiles_WithExclude benchmarks finding files with exclude patterns
func BenchmarkFindFiles_WithExclude(b *testing.B) {
	ctx := context.Background()
	finder := NewFinder()
	query := FindQuery{
		Root:    "testdata/mixed",
		Include: []string{"**/*.go"},
		Exclude: []string{"**/*_test.go"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = finder.FindFiles(ctx, query)
	}
}

// BenchmarkFindGoFiles benchmarks the FindGoFiles convenience method
func BenchmarkFindGoFiles(b *testing.B) {
	ctx := context.Background()
	finder := NewFinder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = finder.FindGoFiles(ctx, "testdata/nested")
	}
}

// BenchmarkValidatePath benchmarks path validation
func BenchmarkValidatePath(b *testing.B) {
	paths := []string{
		"valid/path/to/file.go",
		"another/valid/path.txt",
		"some/deep/nested/path/file.md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = ValidatePath(path)
		}
	}
}
