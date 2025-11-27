package fulhash

import (
	"strings"
	"testing"
)

// TestVerify_Match tests successful verification
func TestVerify_Match(t *testing.T) {
	data := []byte("Test data for verification")

	tests := []struct {
		name string
		alg  Algorithm
	}{
		{"XXH3_128", XXH3_128},
		{"SHA256", SHA256},
		{"CRC32", CRC32},
		{"CRC32C", CRC32C},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute expected digest
			expected, err := Hash(data, WithAlgorithm(tt.alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}

			// Verify should return true
			match, err := Verify(data, expected.String())
			if err != nil {
				t.Fatalf("Verify() failed: %v", err)
			}
			if !match {
				t.Error("Verify() should return true for matching digest")
			}
		})
	}
}

// TestVerify_Mismatch tests verification failure
func TestVerify_Mismatch(t *testing.T) {
	data := []byte("Test data")
	wrongData := []byte("Wrong data")

	// Compute digest for correct data
	expected, err := Hash(data, WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	// Verify with wrong data should return false
	match, err := Verify(wrongData, expected.String())
	if err != nil {
		t.Fatalf("Verify() should not error on mismatch: %v", err)
	}
	if match {
		t.Error("Verify() should return false for mismatched digest")
	}
}

// TestVerify_InvalidFormat tests error handling for invalid digest format
func TestVerify_InvalidFormat(t *testing.T) {
	data := []byte("Test data")

	tests := []struct {
		name     string
		digest   string
		wantErr  bool
		errMatch string
	}{
		{"no separator", "invaliddigest", true, "expected format"},
		{"invalid hex", "sha256:notvalidhex", true, "invalid"},
		{"unsupported algorithm", "md5:abc123", true, "unsupported"},
		{"empty algorithm", ":abc123", true, "unsupported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := Verify(data, tt.digest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("Error should contain %q: %v", tt.errMatch, err)
			}
			if match {
				t.Error("Verify() should not return true when error occurs")
			}
		})
	}
}

// TestVerifyString tests string variant
func TestVerifyString(t *testing.T) {
	data := "Test string data"

	// Compute expected digest
	expected, err := HashString(data, WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("HashString() failed: %v", err)
	}

	// Verify should match
	match, err := VerifyString(data, expected.String())
	if err != nil {
		t.Fatalf("VerifyString() failed: %v", err)
	}
	if !match {
		t.Error("VerifyString() should return true for matching digest")
	}

	// Verify with wrong data should not match
	match, err = VerifyString("Wrong data", expected.String())
	if err != nil {
		t.Fatalf("VerifyString() should not error: %v", err)
	}
	if match {
		t.Error("VerifyString() should return false for mismatched data")
	}
}

// TestVerifyReader tests reader variant
func TestVerifyReader(t *testing.T) {
	data := "Test data for reader verification"

	// Compute expected digest
	reader := strings.NewReader(data)
	expected, err := HashReader(reader, WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("HashReader() failed: %v", err)
	}

	// Verify should match
	reader = strings.NewReader(data)
	match, err := VerifyReader(reader, expected.String())
	if err != nil {
		t.Fatalf("VerifyReader() failed: %v", err)
	}
	if !match {
		t.Error("VerifyReader() should return true for matching digest")
	}

	// Verify with wrong data should not match
	reader = strings.NewReader("Wrong data")
	match, err = VerifyReader(reader, expected.String())
	if err != nil {
		t.Fatalf("VerifyReader() should not error: %v", err)
	}
	if match {
		t.Error("VerifyReader() should return false for mismatched data")
	}
}

// TestVerify_AllAlgorithms tests verification with all supported algorithms
func TestVerify_AllAlgorithms(t *testing.T) {
	data := []byte("Cross-algorithm verification test")

	algorithms := []Algorithm{XXH3_128, SHA256, CRC32, CRC32C}

	for _, alg := range algorithms {
		t.Run(string(alg), func(t *testing.T) {
			// Compute digest
			digest, err := Hash(data, WithAlgorithm(alg))
			if err != nil {
				t.Fatalf("Hash() failed: %v", err)
			}

			// Verify should match
			match, err := Verify(data, digest.String())
			if err != nil {
				t.Fatalf("Verify() failed for %s: %v", alg, err)
			}
			if !match {
				t.Errorf("Verify() failed to match for %s", alg)
			}
		})
	}
}

// TestVerify_CrossAlgorithmMismatch ensures different algorithms don't cross-verify
func TestVerify_CrossAlgorithmMismatch(t *testing.T) {
	data := []byte("Test data")

	// Compute digest with SHA256
	sha256Digest, err := Hash(data, WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	// Manually construct wrong algorithm digest string
	wrongAlgDigest := "xxh3-128:" + sha256Digest.Hex()

	// Should not match (or error if xxh3-128 produces different length)
	match, _ := Verify(data, wrongAlgDigest)
	if match {
		t.Error("Verify() should not match digest from different algorithm")
	}
}
