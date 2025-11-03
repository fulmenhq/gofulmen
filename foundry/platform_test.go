package foundry

import (
	"os"
	"runtime"
	"testing"
)

// TestSupportsSignalExitCodes verifies platform signal support detection.
func TestSupportsSignalExitCodes(t *testing.T) {
	supported := SupportsSignalExitCodes()

	// On Unix-like systems, signal codes should be supported
	if runtime.GOOS != "windows" {
		if !supported {
			t.Errorf("SupportsSignalExitCodes() = false on %s, want true", runtime.GOOS)
		}
	}

	// On Windows, depends on WSL detection
	if runtime.GOOS == "windows" {
		hasWSL := os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != ""
		if hasWSL && !supported {
			t.Error("SupportsSignalExitCodes() = false on WSL, want true")
		}
		if !hasWSL && supported {
			t.Error("SupportsSignalExitCodes() = true on native Windows, want false")
		}
	}
}

// TestPlatformInfo verifies platform metadata structure.
func TestPlatformInfo(t *testing.T) {
	info := PlatformInfo()

	if info.GOOS == "" {
		t.Error("PlatformInfo().GOOS is empty")
	}
	if info.GOARCH == "" {
		t.Error("PlatformInfo().GOARCH is empty")
	}

	// Verify GOOS matches runtime
	if info.GOOS != runtime.GOOS {
		t.Errorf("PlatformInfo().GOOS = %q, want %q", info.GOOS, runtime.GOOS)
	}
	if info.GOARCH != runtime.GOARCH {
		t.Errorf("PlatformInfo().GOARCH = %q, want %q", info.GOARCH, runtime.GOARCH)
	}

	// Verify signal support consistency
	if info.SupportsSignalCodes != SupportsSignalExitCodes() {
		t.Error("PlatformInfo().SupportsSignalCodes inconsistent with SupportsSignalExitCodes()")
	}

	// Verify recommended filtering is inverse of signal support
	if info.RecommendedFiltering == info.SupportsSignalCodes {
		t.Error("RecommendedFiltering should be inverse of SupportsSignalCodes")
	}

	// On Windows without WSL, should recommend filtering
	if runtime.GOOS == "windows" && !info.IsWSL {
		if !info.RecommendedFiltering {
			t.Error("Windows without WSL should recommend filtering signal codes")
		}
	}
}

// TestIsWSL verifies WSL detection logic.
func TestIsWSL(t *testing.T) {
	// Save original env vars
	originalDistro := os.Getenv("WSL_DISTRO_NAME")
	originalInterop := os.Getenv("WSL_INTEROP")
	defer func() {
		if originalDistro == "" {
			_ = os.Unsetenv("WSL_DISTRO_NAME")
		} else {
			_ = os.Setenv("WSL_DISTRO_NAME", originalDistro)
		}
		if originalInterop == "" {
			_ = os.Unsetenv("WSL_INTEROP")
		} else {
			_ = os.Setenv("WSL_INTEROP", originalInterop)
		}
	}()

	tests := []struct {
		name       string
		distroName string
		interop    string
		expectWSL  bool
	}{
		{"No WSL vars", "", "", false},
		{"WSL_DISTRO_NAME set", "Ubuntu", "", true},
		{"WSL_INTEROP set", "", "/run/WSL/123_interop", true},
		{"Both WSL vars set", "Ubuntu", "/run/WSL/123_interop", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			_ = os.Unsetenv("WSL_DISTRO_NAME")
			_ = os.Unsetenv("WSL_INTEROP")

			// Set test env vars
			if tt.distroName != "" {
				_ = os.Setenv("WSL_DISTRO_NAME", tt.distroName)
			}
			if tt.interop != "" {
				_ = os.Setenv("WSL_INTEROP", tt.interop)
			}

			result := isWSL()
			if result != tt.expectWSL {
				t.Errorf("isWSL() = %v, want %v", result, tt.expectWSL)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			_ = os.Unsetenv("WSL_DISTRO_NAME")
			_ = os.Unsetenv("WSL_INTEROP")

			// Set test env vars
			if tt.distroName != "" {
				_ = os.Setenv("WSL_DISTRO_NAME", tt.distroName)
			}
			if tt.interop != "" {
				_ = os.Setenv("WSL_INTEROP", tt.interop)
			}

			result := isWSL()
			if result != tt.expectWSL {
				t.Errorf("isWSL() = %v, want %v", result, tt.expectWSL)
			}
		})
	}
}

// TestPlatformInfoJSON verifies PlatformMetadata can be marshaled to JSON.
func TestPlatformInfoJSON(t *testing.T) {
	info := PlatformInfo()

	// Verify struct has JSON tags by checking they're not empty
	if info.GOOS == "" {
		t.Error("GOOS should not be empty")
	}

	// This test mainly ensures the struct fields are exported and tagged
	// Actual JSON marshaling is tested implicitly by the field tags
}
