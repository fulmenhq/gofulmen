package fulhash

import (
	"bytes"
	"strings"
	"testing"
)

// TestCRC32_BasicFunctionality tests CRC32 algorithm support
func TestCRC32_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		alg   Algorithm
	}{
		{"CRC32 empty", []byte(""), CRC32},
		{"CRC32 hello world", []byte("Hello, World!"), CRC32},
		{"CRC32C empty", []byte(""), CRC32C},
		{"CRC32C hello world", []byte("Hello, World!"), CRC32C},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digest, err := Hash(tt.input, WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}
			if digest.Algorithm() != tt.alg {
				t.Errorf("Algorithm mismatch: got %s, want %s", digest.Algorithm(), tt.alg)
			}
			if len(digest.Bytes()) != 4 {
				t.Errorf("CRC32 should be 4 bytes, got %d", len(digest.Bytes()))
			}
		})
	}
}

// TestCRC32_Consistency verifies CRC32 produces consistent results
func TestCRC32_Consistency(t *testing.T) {
	input := []byte("Test data for CRC32 consistency check")

	// Hash the same data multiple times
	digest1, err := Hash(input, WithAlgorithm(CRC32))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	digest2, err := Hash(input, WithAlgorithm(CRC32))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	if !bytes.Equal(digest1.Bytes(), digest2.Bytes()) {
		t.Errorf("CRC32 produced inconsistent results: %s vs %s", digest1.Hex(), digest2.Hex())
	}
}

// TestCRC32C_Consistency verifies CRC32C produces consistent results
func TestCRC32C_Consistency(t *testing.T) {
	input := []byte("Test data for CRC32C consistency check")

	digest1, err := Hash(input, WithAlgorithm(CRC32C))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	digest2, err := Hash(input, WithAlgorithm(CRC32C))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	if !bytes.Equal(digest1.Bytes(), digest2.Bytes()) {
		t.Errorf("CRC32C produced inconsistent results: %s vs %s", digest1.Hex(), digest2.Hex())
	}
}

// TestCRC32_StreamingVsBlock verifies streaming and block hashing produce same results
func TestCRC32_StreamingVsBlock(t *testing.T) {
	tests := []struct {
		name string
		alg  Algorithm
		data string
	}{
		{"CRC32 short", CRC32, "Hello, World!"},
		{"CRC32 long", CRC32, strings.Repeat("Lorem ipsum dolor sit amet. ", 100)},
		{"CRC32C short", CRC32C, "Hello, World!"},
		{"CRC32C long", CRC32C, strings.Repeat("Lorem ipsum dolor sit amet. ", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockDigest, err := Hash([]byte(tt.data), WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}

			reader := strings.NewReader(tt.data)
			streamDigest, err := HashReader(reader, WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("HashReader() failed: %v", err)
			}

			if !bytes.Equal(blockDigest.Bytes(), streamDigest.Bytes()) {
				t.Errorf("Block vs streaming mismatch: block=%s, stream=%s",
					blockDigest.Hex(), streamDigest.Hex())
			}
		})
	}
}

// TestCRC32_Hasher tests streaming hasher for CRC32
func TestCRC32_Hasher(t *testing.T) {
	tests := []struct {
		name   string
		alg    Algorithm
		chunks []string
	}{
		{"CRC32 multiple writes", CRC32, []string{"Hello, ", "World", "!"}},
		{"CRC32C multiple writes", CRC32C, []string{"Test ", "data ", "for ", "CRC32C"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Stream with multiple writes
			hasher, err := NewHasher(WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("NewHasher() failed: %v", err)
			}

			for _, chunk := range tt.chunks {
				_, err := hasher.Write([]byte(chunk))
				if err != nil {
					t.Fatalf("Write() failed: %v", err)
				}
			}
			streamDigest := hasher.Sum()

			// Compare with block hash
			fullData := strings.Join(tt.chunks, "")
			blockDigest, err := Hash([]byte(fullData), WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}

			if !bytes.Equal(streamDigest.Bytes(), blockDigest.Bytes()) {
				t.Errorf("Hasher mismatch: hasher=%s, block=%s",
					streamDigest.Hex(), blockDigest.Hex())
			}
		})
	}
}

// TestCRC32_ParseDigest verifies CRC32 digests can be parsed
func TestCRC32_ParseDigest(t *testing.T) {
	data := []byte("Test data")

	// Test CRC32
	digest, err := Hash(data, WithAlgorithm(CRC32))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	formatted := digest.String()
	parsed, err := ParseDigest(formatted)
	if err != nil {
		t.Fatalf("ParseDigest() failed: %v", err)
	}

	if parsed.Algorithm() != CRC32 {
		t.Errorf("Algorithm mismatch: got %s, want %s", parsed.Algorithm(), CRC32)
	}
	if !bytes.Equal(parsed.Bytes(), digest.Bytes()) {
		t.Errorf("Bytes mismatch after parse")
	}

	// Test CRC32C
	digest2, err := Hash(data, WithAlgorithm(CRC32C))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	formatted2 := digest2.String()
	parsed2, err := ParseDigest(formatted2)
	if err != nil {
		t.Fatalf("ParseDigest() failed: %v", err)
	}

	if parsed2.Algorithm() != CRC32C {
		t.Errorf("Algorithm mismatch: got %s, want %s", parsed2.Algorithm(), CRC32C)
	}
	if !bytes.Equal(parsed2.Bytes(), digest2.Bytes()) {
		t.Errorf("Bytes mismatch after parse")
	}
}
