package fulhash

import (
	"testing"
)

// Benchmark small payload (1KB)
func BenchmarkHash_Small_XXH3(b *testing.B) {
	data := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(XXH3_128))
	}
}

func BenchmarkHash_Small_SHA256(b *testing.B) {
	data := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(SHA256))
	}
}

// Benchmark medium payload (1MB)
func BenchmarkHash_Medium_XXH3(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(XXH3_128))
	}
}

func BenchmarkHash_Medium_SHA256(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(SHA256))
	}
}

// Benchmark large payload (10MB)
func BenchmarkHash_Large_XXH3(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(XXH3_128))
	}
}

func BenchmarkHash_Large_SHA256(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(SHA256))
	}
}

// Benchmark streaming hasher
func BenchmarkHasher_Write_XXH3(b *testing.B) {
	data := make([]byte, 1024)
	hasher, _ := NewHasher(WithAlgorithm(XXH3_128))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Write(data)
		hasher.Reset()
	}
}

func BenchmarkHasher_Write_SHA256(b *testing.B) {
	data := make([]byte, 1024)
	hasher, _ := NewHasher(WithAlgorithm(SHA256))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Write(data)
		hasher.Reset()
	}
}
