// Package fulhash provides canonical hashing utilities for the Fulmen ecosystem.
//
// This package implements the FulHash module standard for consistent,
// performant hashing across gofulmen, pyfulmen, and tsfulmen.
//
// Reference: docs/crucible-go/standards/library/modules/fulhash.md
// Fixtures: config/crucible-go/library/fulhash/fixtures.yaml
package fulhash

// Algorithm represents a supported hashing algorithm.
type Algorithm string

const (
	// XXH3_128 is the default fast non-cryptographic algorithm.
	XXH3_128 Algorithm = "xxh3-128"
	// SHA256 is the cryptographic algorithm for security-sensitive use cases.
	SHA256 Algorithm = "sha256"
)

// Option configures hashing behavior.
type Option func(*options)

type options struct {
	algorithm  Algorithm
	bufferSize int
}

// WithAlgorithm sets the hashing algorithm.
func WithAlgorithm(alg Algorithm) Option {
	return func(o *options) {
		o.algorithm = alg
	}
}

// WithBufferSize sets the buffer size for streaming operations (default 32KiB).
func WithBufferSize(size int) Option {
	return func(o *options) {
		if size <= 0 {
			o.bufferSize = 32 * 1024 // default
		} else {
			o.bufferSize = size
		}
	}
}

// defaultOptions returns the default options.
func defaultOptions() *options {
	return &options{
		algorithm:  XXH3_128,
		bufferSize: 32 * 1024, // 32KiB
	}
}
