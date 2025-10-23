package fulhash

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/zeebo/xxh3"
)

// Common errors
var (
	// ErrUnsupportedAlgorithm is returned when an unsupported algorithm is requested.
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	// ErrInvalidDigestFormat is returned when a digest string has invalid format.
	ErrInvalidDigestFormat = errors.New("invalid digest format")
)

// Hash computes the hash of the given data.
func Hash(data []byte, opts ...Option) (Digest, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var bytes []byte
	switch o.algorithm {
	case XXH3_128:
		sum := xxh3.Hash128(data)
		b := sum.Bytes()
		bytes = b[:]
	case SHA256:
		h := sha256.New()
		h.Write(data)
		bytes = h.Sum(nil)
	default:
		return Digest{}, fmt.Errorf("%w: %q", ErrUnsupportedAlgorithm, o.algorithm)
	}

	return Digest{algorithm: o.algorithm, bytes: bytes}, nil
}

// HashString computes the hash of the given string.
func HashString(s string, opts ...Option) (Digest, error) {
	return Hash([]byte(s), opts...)
}

// HashReader computes the hash of data from an io.Reader.
func HashReader(r io.Reader, opts ...Option) (Digest, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	hasher, err := newHasher(o.algorithm)
	if err != nil {
		return Digest{}, err
	}

	buf := make([]byte, o.bufferSize)
	_, err = io.CopyBuffer(hasher, r, buf)
	if err != nil {
		return Digest{}, err
	}

	return hasher.Sum(), nil
}

// Hasher is the streaming hasher interface.
type Hasher interface {
	io.Writer
	Sum() Digest
	Reset()
}

// NewHasher creates a new streaming hasher.
func NewHasher(opts ...Option) (Hasher, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return newHasher(o.algorithm)
}

// newHasher creates a hasher for the given algorithm.
func newHasher(alg Algorithm) (Hasher, error) {
	switch alg {
	case XXH3_128:
		return &xxh3Hasher{hasher: xxh3.New()}, nil
	case SHA256:
		return &sha256Hasher{hasher: sha256.New()}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedAlgorithm, alg)
	}
}

// xxh3Hasher implements Hasher for xxh3-128.
type xxh3Hasher struct {
	hasher *xxh3.Hasher
}

func (h *xxh3Hasher) Write(p []byte) (n int, err error) {
	return h.hasher.Write(p)
}

func (h *xxh3Hasher) Sum() Digest {
	sum := h.hasher.Sum128()
	b := sum.Bytes()
	return Digest{algorithm: XXH3_128, bytes: b[:]}
}

func (h *xxh3Hasher) Reset() {
	h.hasher.Reset()
}

// sha256Hasher implements Hasher for sha256.
type sha256Hasher struct {
	hasher hash.Hash
}

func (h *sha256Hasher) Write(p []byte) (n int, err error) {
	return h.hasher.Write(p)
}

func (h *sha256Hasher) Sum() Digest {
	sum := h.hasher.Sum(nil)
	return Digest{algorithm: SHA256, bytes: sum}
}

func (h *sha256Hasher) Reset() {
	h.hasher.Reset()
}
