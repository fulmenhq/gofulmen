package fulpack

import (
	"fmt"

	"github.com/fulmenhq/gofulmen/foundry"
)

// Error codes for fulpack operations.
const (
	// ErrCodeInvalidFormat indicates an unsupported or invalid archive format.
	ErrCodeInvalidFormat = "INVALID_FORMAT"

	// ErrCodePathTraversal indicates a path traversal attempt.
	ErrCodePathTraversal = "PATH_TRAVERSAL"

	// ErrCodeAbsolutePath indicates an absolute path in archive.
	ErrCodeAbsolutePath = "ABSOLUTE_PATH"

	// ErrCodeSymlinkEscape indicates a symlink targeting outside bounds.
	ErrCodeSymlinkEscape = "SYMLINK_ESCAPE"

	// ErrCodeDecompressionBomb indicates a potential decompression bomb.
	ErrCodeDecompressionBomb = "DECOMPRESSION_BOMB"

	// ErrCodeChecksumMismatch indicates checksum verification failure.
	ErrCodeChecksumMismatch = "CHECKSUM_MISMATCH"

	// ErrCodeFileExists indicates target file already exists.
	ErrCodeFileExists = "FILE_EXISTS"

	// ErrCodeCorruptArchive indicates archive structure corruption.
	ErrCodeCorruptArchive = "CORRUPT_ARCHIVE"

	// ErrCodeMaxSizeExceeded indicates max size limit exceeded.
	ErrCodeMaxSizeExceeded = "MAX_SIZE_EXCEEDED"

	// ErrCodeMaxEntriesExceeded indicates max entries limit exceeded.
	ErrCodeMaxEntriesExceeded = "MAX_ENTRIES_EXCEEDED"

	// ErrCodeUnsupportedCompression indicates unsupported compression algorithm.
	ErrCodeUnsupportedCompression = "UNSUPPORTED_COMPRESSION"
)

// Foundry exit code mappings for fulpack errors.
var exitCodeMap = map[string]foundry.ExitCode{
	ErrCodeInvalidFormat:          foundry.ExitInvalidArgument,
	ErrCodePathTraversal:          foundry.ExitSecurityViolation,
	ErrCodeAbsolutePath:           foundry.ExitSecurityViolation,
	ErrCodeSymlinkEscape:          foundry.ExitSecurityViolation,
	ErrCodeDecompressionBomb:      foundry.ExitResourceExhausted,
	ErrCodeChecksumMismatch:       foundry.ExitDataCorrupt,
	ErrCodeFileExists:             foundry.ExitFileWriteError,
	ErrCodeCorruptArchive:         foundry.ExitDataCorrupt,
	ErrCodeMaxSizeExceeded:        foundry.ExitResourceExhausted,
	ErrCodeMaxEntriesExceeded:     foundry.ExitResourceExhausted,
	ErrCodeUnsupportedCompression: foundry.ExitInvalidArgument,
}

// FulpackError represents a fulpack operation error with context.
type FulpackError struct {
	// Code is the error code (e.g., "PATH_TRAVERSAL").
	Code string

	// Message is the human-readable error message.
	Message string

	// Operation is the operation that failed.
	Operation Operation

	// Path is the affected file/entry path (if applicable).
	Path string

	// Cause is the underlying error (if any).
	Cause error
}

// Error implements the error interface.
func (e *FulpackError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("fulpack %s failed: %s [%s] (path: %s)", e.Operation, e.Message, e.Code, e.Path)
	}
	return fmt.Sprintf("fulpack %s failed: %s [%s]", e.Operation, e.Message, e.Code)
}

// Unwrap returns the underlying cause error.
func (e *FulpackError) Unwrap() error {
	return e.Cause
}

// ExitCode returns the appropriate foundry exit code for this error.
func (e *FulpackError) ExitCode() foundry.ExitCode {
	if code, ok := exitCodeMap[e.Code]; ok {
		return code
	}
	return foundry.ExitFailure
}

// newError creates a new FulpackError.
func newError(code string, message string, op Operation, path string, cause error) *FulpackError {
	return &FulpackError{
		Code:      code,
		Message:   message,
		Operation: op,
		Path:      path,
		Cause:     cause,
	}
}

// newErrorf creates a new FulpackError with formatted message.
func newErrorf(code string, op Operation, path string, cause error, format string, args ...interface{}) *FulpackError {
	return &FulpackError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Operation: op,
		Path:      path,
		Cause:     cause,
	}
}
