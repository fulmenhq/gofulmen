package fulhash

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Digest represents a computed hash with metadata.
type Digest struct {
	algorithm Algorithm
	bytes     []byte
}

// Algorithm returns the hashing algorithm used.
func (d Digest) Algorithm() Algorithm {
	return d.algorithm
}

// Bytes returns the raw hash bytes.
func (d Digest) Bytes() []byte {
	return d.bytes
}

// Hex returns the lowercase hexadecimal representation.
func (d Digest) Hex() string {
	return hex.EncodeToString(d.bytes)
}

// String returns the formatted digest as "algorithm:hex".
func (d Digest) String() string {
	return fmt.Sprintf("%s:%s", d.algorithm, d.Hex())
}

// FormatDigest returns the formatted digest string.
func FormatDigest(d Digest) string {
	return d.String()
}

// ParseDigest parses a formatted digest string.
func ParseDigest(s string) (Digest, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Digest{}, fmt.Errorf("%w: expected format 'algorithm:hex', got %q", ErrInvalidDigestFormat, s)
	}

	alg := Algorithm(parts[0])
	hexStr := parts[1]

	if alg != XXH3_128 && alg != SHA256 {
		return Digest{}, fmt.Errorf("%w %q, supported algorithms: %s, %s", ErrUnsupportedAlgorithm, alg, XXH3_128, SHA256)
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return Digest{}, fmt.Errorf("invalid hex in digest %q: %w", s, err)
	}

	return Digest{algorithm: alg, bytes: bytes}, nil
}
