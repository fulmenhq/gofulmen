package fulhash

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"sync"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"github.com/zeebo/xxh3"
)

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrInvalidDigestFormat  = errors.New("invalid digest format")
)

//nolint:unused // Deprecated - kept for backward compatibility, will be removed in Phase 5
var (
	globalTelemetrySystem *telemetry.System
	telemetryMu           sync.RWMutex
)

// SetTelemetrySystem configures the global telemetry system for FulHash operations.
// Deprecated: Use telemetry.SetGlobalSystem() instead. Will be removed in Phase 5.
//
//nolint:unused // Kept for backward compatibility during transition
func SetTelemetrySystem(sys *telemetry.System) {
	telemetryMu.Lock()
	defer telemetryMu.Unlock()
	globalTelemetrySystem = sys
}

//nolint:unused // Kept for backward compatibility during transition
func getTelemetrySystem() *telemetry.System {
	telemetryMu.RLock()
	defer telemetryMu.RUnlock()
	return globalTelemetrySystem
}

// Hash computes the hash of the given data.
//
// Telemetry: Emits algorithm-specific operation counters, bytes_hashed_total, and operation latency.
func Hash(data []byte, opts ...Option) (Digest, error) {
	start := time.Now()
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	tags := map[string]string{
		metrics.TagAlgorithm: string(o.algorithm),
	}

	var bytes []byte
	switch o.algorithm {
	case XXH3_128:
		sum := xxh3.Hash128(data)
		b := sum.Bytes()
		bytes = b[:]
		// Emit XXH3-128 specific counter
		telemetry.EmitCounter(metrics.FulHashOperationsTotalXXH3128, 1, tags)
	case SHA256:
		h := sha256.New()
		h.Write(data)
		bytes = h.Sum(nil)
		// Emit SHA256 specific counter
		telemetry.EmitCounter(metrics.FulHashOperationsTotalSHA256, 1, tags)
	default:
		return Digest{}, fmt.Errorf("%w %q, supported algorithms: %s, %s", ErrUnsupportedAlgorithm, o.algorithm, XXH3_128, SHA256)
	}

	// Emit bytes hashed counter
	telemetry.EmitCounter(metrics.FulHashBytesHashedTotal, float64(len(data)), tags)

	// Emit operation latency
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), tags)

	return Digest{algorithm: o.algorithm, bytes: bytes}, nil
}

// HashString computes the hash of the given string.
//
// Telemetry: Emits hash_string_total counter plus algorithm-specific counters.
func HashString(s string, opts ...Option) (Digest, error) {
	// Emit string-specific counter
	telemetry.EmitCounter(metrics.FulHashHashStringTotal, 1, nil)
	return Hash([]byte(s), opts...)
}

// HashReader computes the hash of data from an io.Reader.
//
// Telemetry: Emits algorithm-specific counters and operation latency.
func HashReader(r io.Reader, opts ...Option) (Digest, error) {
	start := time.Now()
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	tags := map[string]string{
		metrics.TagAlgorithm: string(o.algorithm),
	}

	hasher, err := newHasher(o.algorithm)
	if err != nil {
		return Digest{}, err
	}

	buf := make([]byte, o.bufferSize)
	bytesRead, err := io.CopyBuffer(hasher, r, buf)
	if err != nil {
		return Digest{}, err
	}

	// Emit algorithm-specific counter
	switch o.algorithm {
	case XXH3_128:
		telemetry.EmitCounter(metrics.FulHashOperationsTotalXXH3128, 1, tags)
	case SHA256:
		telemetry.EmitCounter(metrics.FulHashOperationsTotalSHA256, 1, tags)
	}

	// Emit bytes hashed counter
	telemetry.EmitCounter(metrics.FulHashBytesHashedTotal, float64(bytesRead), tags)

	// Emit operation latency
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), tags)

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
