package fulhash

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"github.com/zeebo/xxh3"
)

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrInvalidDigestFormat  = errors.New("invalid digest format")
)

var (
	globalTelemetrySystem *telemetry.System
	telemetryMu           sync.RWMutex
)

// SetTelemetrySystem configures the global telemetry system for FulHash operations.
// Call with a configured *telemetry.System to enable metrics emission.
// Call with nil to disable telemetry (default behavior).
func SetTelemetrySystem(sys *telemetry.System) {
	telemetryMu.Lock()
	defer telemetryMu.Unlock()
	globalTelemetrySystem = sys
}

func getTelemetrySystem() *telemetry.System {
	telemetryMu.RLock()
	defer telemetryMu.RUnlock()
	return globalTelemetrySystem
}

// Hash computes the hash of the given data.
//
// Telemetry: Emits fulhash_hash_count counter on success.
// Emits fulhash_errors_count counter on algorithm errors.
func Hash(data []byte, opts ...Option) (Digest, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	telSys := getTelemetrySystem()
	tags := map[string]string{
		metrics.TagAlgorithm: string(o.algorithm),
		metrics.TagStatus:    metrics.StatusSuccess,
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
		tags[metrics.TagStatus] = metrics.StatusError
		tags[metrics.TagErrorType] = "unsupported_algorithm"
		if telSys != nil {
			_ = telSys.Counter(metrics.FulHashErrorsCount, 1, tags)
		}
		return Digest{}, fmt.Errorf("%w %q, supported algorithms: %s, %s", ErrUnsupportedAlgorithm, o.algorithm, XXH3_128, SHA256)
	}

	if telSys != nil {
		_ = telSys.Counter(metrics.FulHashHashCount, 1, tags)
	}

	return Digest{algorithm: o.algorithm, bytes: bytes}, nil
}

// HashString computes the hash of the given string.
//
// Telemetry: Emits fulhash_hash_count counter on success.
// Emits fulhash_errors_count counter on algorithm errors.
func HashString(s string, opts ...Option) (Digest, error) {
	return Hash([]byte(s), opts...)
}

// HashReader computes the hash of data from an io.Reader.
//
// Telemetry: Emits fulhash_hash_count counter on success.
// Emits fulhash_errors_count counter on I/O or algorithm errors.
func HashReader(r io.Reader, opts ...Option) (Digest, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	telSys := getTelemetrySystem()
	tags := map[string]string{
		metrics.TagAlgorithm: string(o.algorithm),
		metrics.TagStatus:    metrics.StatusSuccess,
	}

	hasher, err := newHasher(o.algorithm)
	if err != nil {
		tags[metrics.TagStatus] = metrics.StatusError
		tags[metrics.TagErrorType] = "unsupported_algorithm"
		if telSys != nil {
			_ = telSys.Counter(metrics.FulHashErrorsCount, 1, tags)
		}
		return Digest{}, err
	}

	buf := make([]byte, o.bufferSize)
	_, err = io.CopyBuffer(hasher, r, buf)
	if err != nil {
		tags[metrics.TagStatus] = metrics.StatusError
		tags[metrics.TagErrorType] = "io_error"
		if telSys != nil {
			_ = telSys.Counter(metrics.FulHashErrorsCount, 1, tags)
		}
		return Digest{}, err
	}

	if telSys != nil {
		_ = telSys.Counter(metrics.FulHashHashCount, 1, tags)
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
		return nil, fmt.Errorf("%w %q, supported algorithms: %s, %s", ErrUnsupportedAlgorithm, alg, XXH3_128, SHA256)
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
