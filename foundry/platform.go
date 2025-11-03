package foundry

import (
	"os"
	"runtime"
)

// PlatformMetadata provides information about the current runtime platform
// and its exit code capabilities.
type PlatformMetadata struct {
	GOOS                 string `json:"goos"`
	GOARCH               string `json:"goarch"`
	IsWSL                bool   `json:"is_wsl"`
	SupportsSignalCodes  bool   `json:"supports_signal_codes"`
	RecommendedFiltering bool   `json:"recommended_filtering"`
}

// SupportsSignalExitCodes returns true if the current platform supports
// POSIX signal-based exit codes (128-165).
//
// Returns false on Windows unless running under WSL (Windows Subsystem for Linux).
// Returns true on all Unix-like systems (Linux, macOS, BSD, etc.).
//
// Use this function to conditionally handle signal exit codes:
//
//	if foundry.SupportsSignalExitCodes() {
//	    // Use EXIT_SIGNAL_TERM, EXIT_SIGNAL_INT, etc.
//	    os.Exit(foundry.ExitSignalTerm)
//	} else {
//	    // Use generic failure on Windows
//	    os.Exit(foundry.ExitFailure)
//	}
func SupportsSignalExitCodes() bool {
	if runtime.GOOS == "windows" {
		// Check if running under WSL
		return isWSL()
	}
	// All Unix-like systems support signal exit codes
	return true
}

// PlatformInfo returns comprehensive metadata about the current platform
// and its exit code capabilities.
//
// This is useful for logging, diagnostics, and telemetry:
//
//	info := foundry.PlatformInfo()
//	logger.Info("platform capabilities",
//	    zap.String("goos", info.GOOS),
//	    zap.Bool("supports_signal_codes", info.SupportsSignalCodes))
func PlatformInfo() PlatformMetadata {
	isWSLEnv := isWSL()
	supportsSignals := runtime.GOOS != "windows" || isWSLEnv

	return PlatformMetadata{
		GOOS:                 runtime.GOOS,
		GOARCH:               runtime.GOARCH,
		IsWSL:                isWSLEnv,
		SupportsSignalCodes:  supportsSignals,
		RecommendedFiltering: !supportsSignals,
	}
}

// isWSL detects if the current process is running under Windows Subsystem for Linux.
// Returns true if WSL_DISTRO_NAME or WSL_INTEROP environment variables are set.
func isWSL() bool {
	// WSL sets WSL_DISTRO_NAME in all WSL 1 and WSL 2 sessions
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	// WSL 2 also sets WSL_INTEROP
	if os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	return false
}
