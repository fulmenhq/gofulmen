package fulhash

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// MultiHash computes hashes for multiple algorithms in a single pass over the data.
// Returns a map of algorithm to digest. Algorithms are deduplicated and validated before processing.
//
// Telemetry: Emits per-algorithm operation counters and aggregate metrics.
// Bytes are counted once even though multiple hashes are computed.
func MultiHash(data []byte, algorithms []Algorithm, opts ...Option) (map[Algorithm]Digest, error) {
	start := time.Now()

	// Short-circuit empty algorithm list
	if len(algorithms) == 0 {
		return make(map[Algorithm]Digest), nil
	}

	// Deduplicate and validate algorithms
	algSet := make(map[Algorithm]bool)
	for _, alg := range algorithms {
		if !alg.IsValid() {
			errorTags := map[string]string{
				metrics.TagErrorType: "unsupported_algorithm",
				metrics.TagStatus:    metrics.StatusError,
				metrics.TagAlgorithm: string(alg),
			}
			telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, errorTags)
			return nil, fmt.Errorf("%w %q", ErrUnsupportedAlgorithm, alg)
		}
		algSet[alg] = true
	}

	// Sort algorithms for deterministic iteration order
	uniqueAlgs := make([]Algorithm, 0, len(algSet))
	for alg := range algSet {
		uniqueAlgs = append(uniqueAlgs, alg)
	}
	sort.Slice(uniqueAlgs, func(i, j int) bool {
		return string(uniqueAlgs[i]) < string(uniqueAlgs[j])
	})

	// Create hashers for all algorithms
	hashers := make([]Hasher, len(uniqueAlgs))
	for i, alg := range uniqueAlgs {
		h, err := newHasher(alg)
		if err != nil {
			return nil, fmt.Errorf("failed to create hasher for %s: %w", alg, err)
		}
		hashers[i] = h
	}

	// Create multi-writer for true single-pass (write data once to all hashers)
	writers := make([]io.Writer, len(hashers))
	for i, h := range hashers {
		writers[i] = h
	}
	multiWriter := io.MultiWriter(writers...)

	// Single pass: write data once, fanned out to all hashers
	_, err := multiWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to hashers: %w", err)
	}

	// Collect results and emit telemetry
	results := make(map[Algorithm]Digest, len(uniqueAlgs))
	for i, alg := range uniqueAlgs {
		digest := hashers[i].Sum()
		results[alg] = digest

		// Emit per-algorithm counter
		tags := map[string]string{
			metrics.TagAlgorithm: string(alg),
		}
		switch alg {
		case XXH3_128:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalXXH3128, 1, tags)
		case SHA256:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalSHA256, 1, tags)
		case CRC32:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalCRC32, 1, tags)
		case CRC32C:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalCRC32C, 1, tags)
		}
	}

	// Emit bytes hashed once (single pass over data)
	telemetry.EmitCounter(metrics.FulHashBytesHashedTotal, float64(len(data)), map[string]string{
		metrics.TagOperation: "multihash",
	})

	// Emit aggregate latency
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), map[string]string{
		metrics.TagOperation: "multihash",
	})

	return results, nil
}

// MultiHashString computes hashes for multiple algorithms on a string in a single pass.
func MultiHashString(s string, algorithms []Algorithm, opts ...Option) (map[Algorithm]Digest, error) {
	return MultiHash([]byte(s), algorithms, opts...)
}

// MultiHashReader computes hashes for multiple algorithms from an io.Reader in a single pass.
// This implementation reads the data once and computes all hashes, avoiding multiple reads.
//
// Telemetry: Emits per-algorithm counters and aggregate metrics.
func MultiHashReader(r io.Reader, algorithms []Algorithm, opts ...Option) (map[Algorithm]Digest, error) {
	start := time.Now()

	// Short-circuit empty algorithm list without consuming reader
	if len(algorithms) == 0 {
		return make(map[Algorithm]Digest), nil
	}

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Deduplicate and validate algorithms
	algSet := make(map[Algorithm]bool)
	for _, alg := range algorithms {
		if !alg.IsValid() {
			errorTags := map[string]string{
				metrics.TagErrorType: "unsupported_algorithm",
				metrics.TagStatus:    metrics.StatusError,
				metrics.TagAlgorithm: string(alg),
			}
			telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, errorTags)
			return nil, fmt.Errorf("%w %q", ErrUnsupportedAlgorithm, alg)
		}
		algSet[alg] = true
	}

	// Sort algorithms for deterministic iteration
	uniqueAlgs := make([]Algorithm, 0, len(algSet))
	for alg := range algSet {
		uniqueAlgs = append(uniqueAlgs, alg)
	}
	sort.Slice(uniqueAlgs, func(i, j int) bool {
		return string(uniqueAlgs[i]) < string(uniqueAlgs[j])
	})

	// Create hashers for all algorithms
	hashers := make([]Hasher, len(uniqueAlgs))
	for i, alg := range uniqueAlgs {
		h, err := newHasher(alg)
		if err != nil {
			return nil, fmt.Errorf("failed to create hasher for %s: %w", alg, err)
		}
		hashers[i] = h
	}

	// Create a multi-writer that writes to all hashers
	writers := make([]io.Writer, len(hashers))
	for i, h := range hashers {
		writers[i] = h
	}
	multiWriter := io.MultiWriter(writers...)

	// Read once and write to all hashers
	buf := make([]byte, o.bufferSize)
	bytesRead, err := io.CopyBuffer(multiWriter, r, buf)
	if err != nil {
		errorTags := map[string]string{
			metrics.TagErrorType: "io_error",
			metrics.TagStatus:    metrics.StatusError,
		}
		telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, errorTags)
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Collect results and emit telemetry
	results := make(map[Algorithm]Digest, len(uniqueAlgs))
	for i, alg := range uniqueAlgs {
		digest := hashers[i].Sum()
		results[alg] = digest

		// Emit per-algorithm counter
		tags := map[string]string{
			metrics.TagAlgorithm: string(alg),
		}
		switch alg {
		case XXH3_128:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalXXH3128, 1, tags)
		case SHA256:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalSHA256, 1, tags)
		case CRC32:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalCRC32, 1, tags)
		case CRC32C:
			telemetry.EmitCounter(metrics.FulHashOperationsTotalCRC32C, 1, tags)
		}
	}

	// Emit bytes hashed (counted once even though used for multiple algorithms)
	telemetry.EmitCounter(metrics.FulHashBytesHashedTotal, float64(bytesRead), map[string]string{
		metrics.TagOperation: "multihash",
	})

	// Emit aggregate latency
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), map[string]string{
		metrics.TagOperation: "multihash",
	})

	return results, nil
}
