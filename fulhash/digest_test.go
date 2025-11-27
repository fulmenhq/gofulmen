package fulhash

import (
	"bytes"
	"testing"

	cruciblefulhash "github.com/fulmenhq/crucible/fulhash"
)

// TestToCrucible verifies conversion to Crucible format
func TestToCrucible(t *testing.T) {
	data := []byte("test data")
	digest, err := Hash(data, WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	cd := digest.ToCrucible()

	// Verify all fields populated
	if cd.Algorithm != string(SHA256) {
		t.Errorf("Algorithm mismatch: got %s, want %s", cd.Algorithm, SHA256)
	}
	if cd.Hex == "" {
		t.Error("Hex field should be populated")
	}
	if cd.Formatted == "" {
		t.Error("Formatted field should be populated")
	}
	if !bytes.Equal(cd.Bytes, digest.Bytes()) {
		t.Error("Bytes field should match digest bytes")
	}

	// Verify formatted matches our String()
	if cd.Formatted != digest.String() {
		t.Errorf("Formatted mismatch: got %s, want %s", cd.Formatted, digest.String())
	}
}

// TestFromCrucible_WithBytes verifies conversion from Crucible format with Bytes
func TestFromCrucible_WithBytes(t *testing.T) {
	testBytes := []byte{0x01, 0x02, 0x03, 0x04}
	cd := cruciblefulhash.Digest{
		Algorithm: "sha256",
		Hex:       "01020304",
		Formatted: "sha256:01020304",
		Bytes:     testBytes,
	}

	digest, err := FromCrucible(cd)
	if err != nil {
		t.Fatalf("FromCrucible() failed: %v", err)
	}

	if digest.Algorithm() != SHA256 {
		t.Errorf("Algorithm mismatch: got %s, want %s", digest.Algorithm(), SHA256)
	}
	if !bytes.Equal(digest.Bytes(), testBytes) {
		t.Error("Bytes should be preserved from Crucible digest")
	}
}

// TestFromCrucible_WithHexOnly verifies conversion when only Hex is present
func TestFromCrucible_WithHexOnly(t *testing.T) {
	cd := cruciblefulhash.Digest{
		Algorithm: "xxh3-128",
		Hex:       "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
		Formatted: "xxh3-128:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
		// Bytes is nil/empty - should decode from Hex
	}

	digest, err := FromCrucible(cd)
	if err != nil {
		t.Fatalf("FromCrucible() failed: %v", err)
	}

	if digest.Algorithm() != XXH3_128 {
		t.Errorf("Algorithm mismatch: got %s, want %s", digest.Algorithm(), XXH3_128)
	}

	// Verify bytes were decoded from hex
	expectedHex := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"
	actualHex := digest.Hex()
	if actualHex != expectedHex {
		t.Errorf("Hex mismatch: got %s, want %s", actualHex, expectedHex)
	}

	if len(digest.Bytes()) != 16 {
		t.Errorf("Expected 16 bytes from hex decode, got %d", len(digest.Bytes()))
	}
}

// TestFromCrucible_Empty verifies error when both Bytes and Hex are empty
func TestFromCrucible_Empty(t *testing.T) {
	cd := cruciblefulhash.Digest{
		Algorithm: "sha256",
		Formatted: "sha256:",
		// Both Bytes and Hex are empty
	}

	_, err := FromCrucible(cd)
	if err == nil {
		t.Error("FromCrucible() should fail when both Bytes and Hex are empty")
	}
}

// TestFromCrucible_InvalidHex verifies error on invalid hex
func TestFromCrucible_InvalidHex(t *testing.T) {
	cd := cruciblefulhash.Digest{
		Algorithm: "sha256",
		Hex:       "notvalidhex",
		Formatted: "sha256:notvalidhex",
	}

	_, err := FromCrucible(cd)
	if err == nil {
		t.Error("FromCrucible() should fail on invalid hex")
	}
}

// TestFromCrucible_UnsupportedAlgorithm verifies error on unsupported algorithm
func TestFromCrucible_UnsupportedAlgorithm(t *testing.T) {
	cd := cruciblefulhash.Digest{
		Algorithm: "md5",
		Hex:       "abc123",
		Formatted: "md5:abc123",
	}

	_, err := FromCrucible(cd)
	if err == nil {
		t.Error("FromCrucible() should fail on unsupported algorithm")
	}
}

// TestRoundTrip verifies ToCrucible -> FromCrucible preserves data
func TestRoundTrip(t *testing.T) {
	data := []byte("roundtrip test data")
	algorithms := []Algorithm{XXH3_128, SHA256, CRC32, CRC32C}

	for _, alg := range algorithms {
		t.Run(string(alg), func(t *testing.T) {
			original, err := Hash(data, WithAlgorithm(alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}

			// Convert to Crucible and back
			cd := original.ToCrucible()
			roundtripped, err := FromCrucible(cd)
			if err != nil {
				t.Fatalf("FromCrucible() failed: %v", err)
			}

			// Verify identical
			if roundtripped.Algorithm() != original.Algorithm() {
				t.Error("Algorithm changed after roundtrip")
			}
			if !bytes.Equal(roundtripped.Bytes(), original.Bytes()) {
				t.Error("Bytes changed after roundtrip")
			}
			if roundtripped.String() != original.String() {
				t.Errorf("String changed after roundtrip: got %s, want %s",
					roundtripped.String(), original.String())
			}
		})
	}
}
