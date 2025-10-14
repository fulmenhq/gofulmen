package foundry

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// DetectMimeTypeFromReader detects MIME type from streaming data.
//
// Reads up to maxBytes from the reader to detect the file type without
// buffering the entire stream. Returns a MimeType and a new io.Reader
// that includes the already-read bytes for continued processing.
//
// If maxBytes <= 0, defaults to 512 bytes (sufficient for most magic numbers).
//
// This function is ideal for:
//   - HTTP request body inspection
//   - Large file processing without full buffering
//   - Streaming data validation
//   - Protocol detection in network streams
//
// Example:
//
//	func handleUpload(w http.ResponseWriter, r *http.Request) {
//	    mimeType, reader, err := foundry.DetectMimeTypeFromReader(r.Body, 512)
//	    if err != nil {
//	        http.Error(w, "Detection failed", 500)
//	        return
//	    }
//
//	    if mimeType != nil && mimeType.ID == "json" {
//	        // Continue processing with reader (includes already-read bytes)
//	        processJSON(reader)
//	    } else {
//	        http.Error(w, "JSON required", 415)
//	    }
//	}
func DetectMimeTypeFromReader(r io.Reader, maxBytes int) (*MimeType, io.Reader, error) {
	if maxBytes <= 0 {
		maxBytes = 512 // Default: sufficient for most magic numbers
	}

	// Read up to maxBytes for detection
	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	// Truncate buffer to actual bytes read
	buf = buf[:n]

	// Detect MIME type from buffer
	mimeType, detectErr := DetectMimeType(buf)
	if detectErr != nil {
		// Return error but preserve the reader with buffered bytes
		newReader := io.MultiReader(bytes.NewReader(buf), r)
		return nil, newReader, fmt.Errorf("MIME detection failed: %w", detectErr)
	}

	// Combine buffered bytes with remaining stream
	newReader := io.MultiReader(bytes.NewReader(buf), r)
	return mimeType, newReader, nil
}

// DetectMimeTypeFromFile detects MIME type from a file path.
//
// This is a convenience function that opens the file, reads the beginning
// for MIME detection, and closes it. Returns the detected MIME type or nil
// if the type cannot be determined.
//
// This function does not return a reader; use os.Open if you need to
// continue processing the file after detection.
//
// Example:
//
//	mimeType, err := foundry.DetectMimeTypeFromFile("upload.dat")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if mimeType != nil {
//	    fmt.Println("Detected:", mimeType.Mime)
//	    // application/json
//	} else {
//	    fmt.Println("Unknown file type")
//	}
func DetectMimeTypeFromFile(path string) (*MimeType, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes for detection (standard magic number size)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Truncate to actual bytes read
	buf = buf[:n]

	// Detect MIME type
	mimeType, err := DetectMimeType(buf)
	if err != nil {
		return nil, fmt.Errorf("MIME detection failed: %w", err)
	}

	return mimeType, nil
}
