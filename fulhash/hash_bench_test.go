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

// CRC32 Benchmarks
func BenchmarkHash_Small_CRC32(b *testing.B) {
	data := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32))
	}
}

func BenchmarkHash_Medium_CRC32(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32))
	}
}

func BenchmarkHash_Large_CRC32(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32))
	}
}

func BenchmarkHasher_Write_CRC32(b *testing.B) {
	data := make([]byte, 1024)
	hasher, _ := NewHasher(WithAlgorithm(CRC32))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Write(data)
		hasher.Reset()
	}
}

// CRC32C Benchmarks
func BenchmarkHash_Small_CRC32C(b *testing.B) {
	data := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32C))
	}
}

func BenchmarkHash_Medium_CRC32C(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32C))
	}
}

func BenchmarkHash_Large_CRC32C(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data, WithAlgorithm(CRC32C))
	}
}

func BenchmarkHasher_Write_CRC32C(b *testing.B) {
	data := make([]byte, 1024)
	hasher, _ := NewHasher(WithAlgorithm(CRC32C))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Write(data)
		hasher.Reset()
	}
}

// MultiHash Benchmarks
func BenchmarkMultiHash_AllAlgorithms(b *testing.B) {
	data := make([]byte, 1024)
	algorithms := []Algorithm{XXH3_128, SHA256, CRC32, CRC32C}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MultiHash(data, algorithms)
	}
}

// Verify Benchmarks
func BenchmarkVerify_Match(b *testing.B) {
	data := make([]byte, 1024)
	expected, _ := Hash(data, WithAlgorithm(SHA256))
	expectedStr := expected.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Verify(data, expectedStr)
	}
}
