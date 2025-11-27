package fulhash

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// Verify validates that data matches the expected digest in "algorithm:hex" format.
// Returns (true, nil) if the digest matches, (false, nil) if it doesn't match,
// and (false, error) for parse errors, unsupported algorithms, or I/O failures.
//
// The expected format is "algorithm:hex", e.g., "sha256:abc123..." or "xxh3-128:def456...".
//
// Telemetry: Emits verification counters with match/mismatch status.
func Verify(data []byte, expected string, opts ...Option) (bool, error) {
	start := time.Now()

	// Parse the expected digest
	expectedDigest, err := ParseDigest(expected)
	if err != nil {
		errorTags := map[string]string{
			metrics.TagErrorType: "parse_error",
			metrics.TagStatus:    metrics.StatusError,
		}
		telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, errorTags)
		return false, fmt.Errorf("failed to parse expected digest: %w", err)
	}

	// Compute actual digest using the algorithm from the expected digest
	algOpts := append([]Option{WithAlgorithm(expectedDigest.Algorithm())}, opts...)
	actualDigest, err := Hash(data, algOpts...)
	if err != nil {
		return false, fmt.Errorf("failed to compute digest: %w", err)
	}

	// Compare digests
	matches := bytes.Equal(actualDigest.Bytes(), expectedDigest.Bytes())

	// Emit telemetry
	tags := map[string]string{
		metrics.TagAlgorithm: string(expectedDigest.Algorithm()),
		metrics.TagOperation: "verify",
	}
	if matches {
		tags[metrics.TagResult] = "match"
		tags[metrics.TagStatus] = metrics.StatusSuccess
	} else {
		tags[metrics.TagResult] = "mismatch"
		tags[metrics.TagStatus] = metrics.StatusError
		// Emit explicit mismatch counter for monitoring
		mismatchTags := map[string]string{
			metrics.TagAlgorithm: string(expectedDigest.Algorithm()),
			metrics.TagErrorType: "digest_mismatch",
			metrics.TagStatus:    metrics.StatusError,
			metrics.TagResult:    "mismatch",
		}
		telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, mismatchTags)
	}
	telemetry.EmitCounter(metrics.FulHashHashCount, 1, tags)
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), tags)

	return matches, nil
}

// VerifyString validates that a string matches the expected digest.
func VerifyString(s string, expected string, opts ...Option) (bool, error) {
	return Verify([]byte(s), expected, opts...)
}

// VerifyReader validates that data from an io.Reader matches the expected digest.
// The reader is consumed once during verification.
//
// Telemetry: Emits verification counters with match/mismatch status.
func VerifyReader(r io.Reader, expected string, opts ...Option) (bool, error) {
	start := time.Now()

	// Parse the expected digest
	expectedDigest, err := ParseDigest(expected)
	if err != nil {
		errorTags := map[string]string{
			metrics.TagErrorType: "parse_error",
			metrics.TagStatus:    metrics.StatusError,
		}
		telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, errorTags)
		return false, fmt.Errorf("failed to parse expected digest: %w", err)
	}

	// Compute actual digest using the algorithm from the expected digest
	algOpts := append([]Option{WithAlgorithm(expectedDigest.Algorithm())}, opts...)
	actualDigest, err := HashReader(r, algOpts...)
	if err != nil {
		return false, fmt.Errorf("failed to compute digest: %w", err)
	}

	// Compare digests
	matches := bytes.Equal(actualDigest.Bytes(), expectedDigest.Bytes())

	// Emit telemetry
	tags := map[string]string{
		metrics.TagAlgorithm: string(expectedDigest.Algorithm()),
		metrics.TagOperation: "verify",
	}
	if matches {
		tags[metrics.TagResult] = "match"
		tags[metrics.TagStatus] = metrics.StatusSuccess
	} else {
		tags[metrics.TagResult] = "mismatch"
		tags[metrics.TagStatus] = metrics.StatusError
		// Emit explicit mismatch counter for monitoring
		mismatchTags := map[string]string{
			metrics.TagAlgorithm: string(expectedDigest.Algorithm()),
			metrics.TagErrorType: "digest_mismatch",
			metrics.TagStatus:    metrics.StatusError,
			metrics.TagResult:    "mismatch",
		}
		telemetry.EmitCounter(metrics.FulHashErrorsCount, 1, mismatchTags)
	}
	telemetry.EmitCounter(metrics.FulHashHashCount, 1, tags)
	telemetry.EmitHistogram(metrics.FulHashOperationMs, time.Since(start), tags)

	return matches, nil
}
