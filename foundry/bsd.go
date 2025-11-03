package foundry

// MapToBSD maps a Fulmen exit code to its BSD equivalent.
// Returns (bsdCode, true) if a BSD mapping exists, (0, false) if not.
//
// BSD exit codes from <sysexits.h> provide standardized error codes
// for command-line programs. This function enables compatibility with
// BSD-style exit codes for scripts and utilities that expect them.
//
// Example usage:
//
//	if bsdCode, ok := foundry.MapToBSD(foundry.ExitConfigInvalid); ok {
//	    os.Exit(bsdCode)
//	}
func MapToBSD(code ExitCode) (int, bool) {
	info, found := GetExitCodeInfo(code)
	if !found {
		return 0, false
	}

	if info.BSDEquivalent == "" {
		return 0, false
	}

	// The BSDEquivalent field contains the BSD constant name (e.g., "EX_CONFIG")
	// We need to map these to their numeric values from sysexits.h
	bsdCode, ok := bsdNameToCode[info.BSDEquivalent]
	if !ok {
		return 0, false
	}

	return bsdCode, true
}

// MapFromBSD maps a BSD exit code to its Fulmen equivalent.
// Returns (fulmenCode, true) if a mapping exists, (0, false) if not.
//
// This is useful for converting exit codes from BSD-style programs
// back to Fulmen's standardized exit codes for consistent handling.
//
// Example usage:
//
//	if fulmenCode, ok := foundry.MapFromBSD(78); ok {
//	    // Handle as fulmen exit code
//	}
func MapFromBSD(bsdCode int) (ExitCode, bool) {
	// First, find the BSD constant name
	var bsdName string
	for name, code := range bsdNameToCode {
		if code == bsdCode {
			bsdName = name
			break
		}
	}

	if bsdName == "" {
		return 0, false
	}

	// Now search through all exit codes for one with this BSD equivalent
	allCodes := ListExitCodes()
	for _, info := range allCodes {
		if info.BSDEquivalent == bsdName {
			return info.Code, true
		}
	}

	return 0, false
}

// GetBSDCodeInfo returns metadata about a BSD exit code.
// Returns nil if the code is not recognized.
func GetBSDCodeInfo(bsdCode int) *BSDCodeInfo {
	for name, code := range bsdNameToCode {
		if code == bsdCode {
			if desc, ok := bsdDescriptions[name]; ok {
				return &BSDCodeInfo{
					Code:        bsdCode,
					Name:        name,
					Description: desc,
				}
			}
			return &BSDCodeInfo{
				Code: bsdCode,
				Name: name,
			}
		}
	}
	return nil
}

// BSDCodeInfo provides metadata about a BSD exit code.
type BSDCodeInfo struct {
	Code        int    `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// bsdNameToCode maps BSD sysexits.h constant names to their numeric values.
// Source: <sysexits.h> standard constants
var bsdNameToCode = map[string]int{
	"EX_OK":          0,  // successful termination
	"EX_USAGE":       64, // command line usage error
	"EX_DATAERR":     65, // data format error
	"EX_NOINPUT":     66, // cannot open input
	"EX_NOUSER":      67, // addressee unknown
	"EX_NOHOST":      68, // host name unknown
	"EX_UNAVAILABLE": 69, // service unavailable
	"EX_SOFTWARE":    70, // internal software error
	"EX_OSERR":       71, // system error (e.g., can't fork)
	"EX_OSFILE":      72, // critical OS file missing
	"EX_CANTCREAT":   73, // can't create (user) output file
	"EX_IOERR":       74, // input/output error
	"EX_TEMPFAIL":    75, // temp failure; user is invited to retry
	"EX_PROTOCOL":    76, // remote error in protocol
	"EX_NOPERM":      77, // permission denied
	"EX_CONFIG":      78, // configuration error
}

// bsdDescriptions provides human-readable descriptions for BSD exit codes.
var bsdDescriptions = map[string]string{
	"EX_OK":          "Successful termination",
	"EX_USAGE":       "Command line usage error",
	"EX_DATAERR":     "Data format error",
	"EX_NOINPUT":     "Cannot open input",
	"EX_NOUSER":      "Addressee unknown",
	"EX_NOHOST":      "Host name unknown",
	"EX_UNAVAILABLE": "Service unavailable",
	"EX_SOFTWARE":    "Internal software error",
	"EX_OSERR":       "System error (e.g., can't fork)",
	"EX_OSFILE":      "Critical OS file missing",
	"EX_CANTCREAT":   "Can't create (user) output file",
	"EX_IOERR":       "Input/output error",
	"EX_TEMPFAIL":    "Temporary failure; user is invited to retry",
	"EX_PROTOCOL":    "Remote error in protocol",
	"EX_NOPERM":      "Permission denied",
	"EX_CONFIG":      "Configuration error",
}
