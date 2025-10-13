package bootstrap

import (
	"fmt"
	"runtime"
	"strings"
)

// Platform represents the operating system and architecture
type Platform struct {
	OS   string // darwin, linux, windows
	Arch string // amd64, arm64
}

// GetPlatform returns the current platform
func GetPlatform() Platform {
	return Platform{
		OS:   normalizeOS(runtime.GOOS),
		Arch: normalizeArch(runtime.GOARCH),
	}
}

// String returns the platform as "os-arch" (e.g., "darwin-arm64")
func (p Platform) String() string {
	return fmt.Sprintf("%s-%s", p.OS, p.Arch)
}

// normalizeOS converts runtime.GOOS to our standard values
func normalizeOS(goos string) string {
	switch goos {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return goos
	}
}

// normalizeArch converts runtime.GOARCH to our standard values
func normalizeArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return goarch
	}
}

// InterpolateURL replaces {{os}} and {{arch}} placeholders in a URL
// with the current platform values
//
// Example:
//
//	url := "https://example.com/tool_{{os}}_{{arch}}.tar.gz"
//	result := InterpolateURL(url, platform)
//	// result: "https://example.com/tool_darwin_arm64.tar.gz" (on macOS ARM)
func InterpolateURL(urlTemplate string, platform Platform) string {
	result := urlTemplate
	result = strings.ReplaceAll(result, "{{os}}", platform.OS)
	result = strings.ReplaceAll(result, "{{arch}}", platform.Arch)
	return result
}

// IsPlatformSupported checks if the current platform is supported for downloads
// (macOS and Linux are fully supported, Windows has limitations)
func IsPlatformSupported(platform Platform) (bool, string) {
	switch platform.OS {
	case "darwin", "linux":
		return true, ""
	case "windows":
		return true, "Windows support requires tar and gzip to be installed manually (e.g., via scoop)"
	default:
		return false, fmt.Sprintf("Unsupported platform: %s", platform.OS)
	}
}
