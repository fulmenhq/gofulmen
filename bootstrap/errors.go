package bootstrap

import "fmt"

type ChecksumMismatchError struct {
	FilePath string
	Expected string
	Actual   string
}

func (e *ChecksumMismatchError) Error() string {
	return fmt.Sprintf(`checksum verification failed:
   File: %s
   Expected: %s
   Actual:   %s
  
   This could indicate:
   - File was corrupted during download
   - File has been tampered with
   - Wrong checksum in manifest`, e.FilePath, e.Expected, e.Actual)
}

type CommandNotFoundError struct {
	Command    string
	Suggestion string
}

func (e *CommandNotFoundError) Error() string {
	msg := fmt.Sprintf("command not found: %s", e.Command)
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\n\n   %s", e.Suggestion)
	}
	return msg
}

type DownloadError struct {
	URL      string
	Platform Platform
	Err      error
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf(`failed to download:
   Platform: %s
   URL: %s
   Error: %v
  
   Possible solutions:
   - Check if the release exists for %s
   - Verify the URL pattern in the manifest`, e.Platform, e.URL, e.Err, e.Platform)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

type ExtractionError struct {
	Archive string
	Err     error
}

func (e *ExtractionError) Error() string {
	return fmt.Sprintf("failed to extract archive %s: %v", e.Archive, e.Err)
}

func (e *ExtractionError) Unwrap() error {
	return e.Err
}

type UnsafePath struct {
	Path string
}

func (e *UnsafePath) Error() string {
	return fmt.Sprintf("unsafe path in archive: %s (contains '..' or is absolute)", e.Path)
}

type ManifestError struct {
	Path string
	Err  error
}

func (e *ManifestError) Error() string {
	return fmt.Sprintf("invalid manifest at %s: %v", e.Path, e.Err)
}

func (e *ManifestError) Unwrap() error {
	return e.Err
}
