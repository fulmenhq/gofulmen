package fulhash

import (
	"bytes"
	"strings"
	"testing"
)

// TestMultiHash_BasicFunctionality tests MultiHash with multiple algorithms
func TestMultiHash_BasicFunctionality(t *testing.T) {
	data := []byte("Test data for multihash")
	algorithms := []Algorithm{XXH3_128, SHA256, CRC32, CRC32C}

	results, err := MultiHash(data, algorithms)
	if err != nil {
		t.Fatalf("MultiHash() failed: %v", err)
	}

	// Verify all algorithms produced results
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// Verify each algorithm's result matches individual Hash() calls
	for _, alg := range algorithms {
		digest, ok := results[alg]
		if !ok {
			t.Errorf("Algorithm %s missing from results", alg)
			continue
		}

		// Compute individual hash
		expected, err := Hash(data, WithAlgorithm(alg))
		if err != nil {
			t.Fatalf("Hash() failed for %s: %v", alg, err)
		}

		if !bytes.Equal(digest.Bytes(), expected.Bytes()) {
			t.Errorf("MultiHash result for %s doesn't match individual Hash()", alg)
		}
	}
}

// TestMultiHash_Deduplication verifies duplicate algorithms are handled
func TestMultiHash_Deduplication(t *testing.T) {
	data := []byte("Test data")
	// Provide duplicates
	algorithms := []Algorithm{XXH3_128, SHA256, XXH3_128, SHA256, CRC32}

	results, err := MultiHash(data, algorithms)
	if err != nil {
		t.Fatalf("MultiHash() failed: %v", err)
	}

	// Should only have 3 unique results
	if len(results) != 3 {
		t.Errorf("Expected 3 deduplicated results, got %d", len(results))
	}

	expectedAlgs := []Algorithm{XXH3_128, SHA256, CRC32}
	for _, alg := range expectedAlgs {
		if _, ok := results[alg]; !ok {
			t.Errorf("Expected algorithm %s in results", alg)
		}
	}
}

// TestMultiHash_UnsupportedAlgorithm verifies error handling
func TestMultiHash_UnsupportedAlgorithm(t *testing.T) {
	data := []byte("Test data")
	algorithms := []Algorithm{XXH3_128, Algorithm("unsupported"), SHA256}

	_, err := MultiHash(data, algorithms)
	if err == nil {
		t.Error("Expected error for unsupported algorithm")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("Error should mention unsupported algorithm: %v", err)
	}
}

// TestMultiHashString tests string variant
func TestMultiHashString(t *testing.T) {
	data := "Test string data"
	algorithms := []Algorithm{XXH3_128, SHA256}

	results, err := MultiHashString(data, algorithms)
	if err != nil {
		t.Fatalf("MultiHashString() failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify results match Hash() on same data
	for _, alg := range algorithms {
		expected, err := HashString(data, WithAlgorithm(alg))
		if err != nil {
			t.Fatalf("HashString() failed: %v", err)
		}

		digest := results[alg]
		if !bytes.Equal(digest.Bytes(), expected.Bytes()) {
			t.Errorf("MultiHashString result doesn't match HashString for %s", alg)
		}
	}
}

// TestMultiHashReader tests reader variant
func TestMultiHashReader(t *testing.T) {
	data := "Test data for multihash reader"
	algorithms := []Algorithm{XXH3_128, SHA256, CRC32, CRC32C}

	reader := strings.NewReader(data)
	results, err := MultiHashReader(reader, algorithms)
	if err != nil {
		t.Fatalf("MultiHashReader() failed: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// Verify results match individual HashReader() calls
	for _, alg := range algorithms {
		r := strings.NewReader(data)
		expected, err := HashReader(r, WithAlgorithm(alg))
		if err != nil {
			t.Fatalf("HashReader() failed for %s: %v", alg, err)
		}

		digest := results[alg]
		if !bytes.Equal(digest.Bytes(), expected.Bytes()) {
			t.Errorf("MultiHashReader result doesn't match HashReader for %s", alg)
		}
	}
}

// TestMultiHashReader_SinglePass verifies reader is only read once
func TestMultiHashReader_SinglePass(t *testing.T) {
	data := "Test data"
	algorithms := []Algorithm{XXH3_128, SHA256, CRC32}

	// Use a reader that can only be read once
	reader := strings.NewReader(data)
	results, err := MultiHashReader(reader, algorithms)
	if err != nil {
		t.Fatalf("MultiHashReader() failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify reader was consumed
	n, _ := reader.Read(make([]byte, 1))
	if n != 0 {
		t.Error("Reader should be fully consumed")
	}
}

// TestMultiHashReader_EmptyAlgorithms tests edge case
func TestMultiHashReader_EmptyAlgorithms(t *testing.T) {
	reader := strings.NewReader("test")
	results, err := MultiHashReader(reader, []Algorithm{})
	if err != nil {
		t.Fatalf("MultiHashReader() failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty algorithms, got %d", len(results))
	}
}
