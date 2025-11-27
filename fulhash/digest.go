package fulhash

import (
	"encoding/hex"
	"fmt"
	"strings"

	cruciblefulhash "github.com/fulmenhq/crucible/fulhash"
)

// Digest represents a computed hash with metadata.
// This wraps the algorithm and raw bytes, providing convenient accessor methods
// while maintaining source compatibility with existing gofulmen code.
//
// The internal representation uses simple fields for efficient computation,
// but can be converted to/from Crucible's standard Digest format for
// interoperability across the FulmenHQ ecosystem.
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
// This format matches the Crucible standard digest representation.
func (d Digest) String() string {
	return fmt.Sprintf("%s:%s", d.algorithm, d.Hex())
}

// FormatDigest returns the formatted digest string.
func FormatDigest(d Digest) string {
	return d.String()
}

// ToCrucible converts this digest to Crucible's standard Digest format.
// This enables interoperability with other FulmenHQ libraries (pyfulmen, tsfulmen).
func (d Digest) ToCrucible() cruciblefulhash.Digest {
	hexStr := d.Hex()
	formatted := d.String()
	return cruciblefulhash.Digest{
		Algorithm: string(d.algorithm),
		Hex:       hexStr,
		Formatted: formatted,
		Bytes:     d.bytes,
	}
}

// FromCrucible creates a Digest from Crucible's standard Digest format.
// This enables consuming digests from other FulmenHQ libraries.
// If Bytes is empty, decodes from Hex field (per SSOT schema, Bytes is optional).
func FromCrucible(cd cruciblefulhash.Digest) (Digest, error) {
	alg := Algorithm(cd.Algorithm)
	if !alg.IsValid() {
		return Digest{}, fmt.Errorf("%w %q", ErrUnsupportedAlgorithm, alg)
	}

	// Use Bytes if present, otherwise decode from Hex
	var bytes []byte
	if len(cd.Bytes) > 0 {
		bytes = cd.Bytes
	} else if cd.Hex != "" {
		var err error
		bytes, err = hex.DecodeString(cd.Hex)
		if err != nil {
			return Digest{}, fmt.Errorf("invalid hex in Crucible digest: %w", err)
		}
	} else {
		return Digest{}, fmt.Errorf("crucible digest missing both bytes and hex fields")
	}

	return Digest{algorithm: alg, bytes: bytes}, nil
}

// ParseDigest parses a formatted digest string in "algorithm:hex" format.
// Supports all algorithms: xxh3-128, sha256, crc32, crc32c.
func ParseDigest(s string) (Digest, error) {

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Digest{}, fmt.Errorf("%w: expected format 'algorithm:hex', got %q", ErrInvalidDigestFormat, s)
	}

	alg := Algorithm(parts[0])
	hexStr := parts[1]

	// Validate algorithm is supported
	if !alg.IsValid() {
		return Digest{}, fmt.Errorf("%w %q, supported algorithms: %s, %s, %s, %s",
			ErrUnsupportedAlgorithm, alg, XXH3_128, SHA256, CRC32, CRC32C)
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return Digest{}, fmt.Errorf("invalid hex in digest %q: %w", s, err)
	}

	return Digest{algorithm: alg, bytes: bytes}, nil
}
