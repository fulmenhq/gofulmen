// Package buildinfo resolves identity for the running host binary.
//
// It does not describe the gofulmen library checkout. Resolve reads caller
// linker stamps and Go's compile-time build metadata only; it never reads
// FULMEN_HOST_* process variables or executes Git. Use crucible.GetVersion for
// gofulmen and Crucible SDK pins, and pass those pins to FormatExtended when a
// host command wants to display both identities.
//
// Callers normally keep linker variables in their main package and pass them
// to Resolve:
//
//	dirty="$(if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
//	  status="$(git status --porcelain --untracked-files=all)" || exit
//	  test -n "$status" && printf true || printf false
//	fi)"
//	go build -ldflags "-X main.version=$VERSION -X main.commit=$(git rev-parse HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.dirty=$dirty"
//
// Call ResolveWithDirty when main.dirty is injected. Full Git HEAD values are
// retained in Info and JSON; FormatExtended shortens only full hexadecimal
// object IDs for human output.
package buildinfo

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/fulmenhq/gofulmen/crucible"
)

const (
	defaultVersion    = "dev"
	defaultUnknown    = "unknown"
	shortCommitLength = 7
)

// Info identifies the running host binary. Dirty is nil when its state was
// not compiled into the binary metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	Dirty     *bool  `json:"dirty,omitempty"`
	GoVersion string `json:"runtime"`
	Platform  string `json:"platform"`
}

// Resolve returns host-binary identity from caller linker stamps and
// compile-time Go build metadata. It never consults process environment or
// executes Git.
func Resolve(ldVersion, ldCommit, ldBuildDate string) Info {
	return ResolveWithDirty(ldVersion, ldCommit, ldBuildDate, "")
}

// ResolveWithDirty is Resolve with an optional linker-injected dirty stamp.
// A strict true or false dirty stamp takes precedence over build metadata.
func ResolveWithDirty(ldVersion, ldCommit, ldBuildDate, ldDirty string) Info {
	info, _ := debug.ReadBuildInfo()
	return resolveWithBuildInfo(ldVersion, ldCommit, ldBuildDate, ldDirty, "", info)
}

// ResolveWithEmbed is Resolve with an optional application version embedded at
// build time. A non-placeholder embedded version fills only after build-info
// metadata and before the dev fallback.
func ResolveWithEmbed(ldVersion, ldCommit, ldBuildDate, embeddedVersion string) Info {
	info, _ := debug.ReadBuildInfo()
	return resolveWithBuildInfo(ldVersion, ldCommit, ldBuildDate, "", embeddedVersion, info)
}

func resolveWithBuildInfo(ldVersion, ldCommit, ldBuildDate, ldDirty, embeddedVersion string, buildInfo *debug.BuildInfo) Info {
	result := Info{
		Version:   defaultVersion,
		Commit:    defaultUnknown,
		BuildDate: defaultUnknown,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if buildInfo != nil {
		if version := normalizeVersion(buildInfo.Main.Version); version != "" {
			result.Version = version
		}
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if !isPlaceholder(setting.Value) {
					result.Commit = strings.TrimSpace(setting.Value)
				}
			case "vcs.time":
				if !isPlaceholder(setting.Value) {
					result.BuildDate = strings.TrimSpace(setting.Value)
				}
			case "vcs.modified":
				if dirty, ok := parseDirty(setting.Value); ok {
					result.Dirty = &dirty
				}
			}
		}
	}

	if result.Version == defaultVersion {
		if version := normalizeVersion(embeddedVersion); version != "" {
			result.Version = version
		}
	}
	if !isPlaceholder(ldVersion) {
		result.Version = strings.TrimSpace(ldVersion)
	}
	if !isPlaceholder(ldCommit) {
		result.Commit = strings.TrimSpace(ldCommit)
	}
	if !isPlaceholder(ldBuildDate) {
		result.BuildDate = strings.TrimSpace(ldBuildDate)
	}
	if dirty, ok := parseDirty(ldDirty); ok {
		result.Dirty = &dirty
	}
	return result
}

// FormatBasic renders the binary name and version on one line.
func (i Info) FormatBasic(binaryName string) string {
	if strings.TrimSpace(binaryName) == "" {
		return i.Version
	}
	return strings.TrimSpace(binaryName) + " " + i.Version
}

// FormatExtended renders host identity and, when supplied, a separately
// labeled gofulmen/Crucible SDK pin block.
func (i Info) FormatExtended(binaryName string, pins *crucible.Version) string {
	lines := []string{i.FormatBasic(binaryName)}
	if !isPlaceholder(i.Commit) {
		lines = append(lines, "Commit: "+shortCommit(i.Commit))
	}
	if !isPlaceholder(i.BuildDate) {
		lines = append(lines, "BuildDate: "+i.BuildDate)
	}
	if i.Dirty != nil {
		lines = append(lines, fmt.Sprintf("Dirty: %t", *i.Dirty))
	}
	lines = append(lines,
		"Runtime: "+i.GoVersion,
		"Platform: "+i.Platform,
	)
	if pins != nil {
		lines = append(lines,
			"Pins:",
			"  gofulmen: "+pins.Gofulmen,
			"  Crucible: "+pins.Crucible,
		)
	}
	return strings.Join(lines, "\n")
}

// JSON returns Info with stable field names. Commit is preserved in full;
// shortening is only applied to human-readable extended output.
func (i Info) JSON() ([]byte, error) {
	return json.Marshal(i)
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(value)
	if isPlaceholder(value) || value == "(devel)" {
		return ""
	}
	return strings.TrimPrefix(value, "v")
}

func isPlaceholder(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "dev", "unknown":
		return true
	default:
		return false
	}
}

func parseDirty(value string) (bool, bool) {
	switch strings.TrimSpace(value) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

func shortCommit(value string) string {
	if !isFullHexObjectID(value) {
		return value
	}
	return value[:shortCommitLength]
}

func isFullHexObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !isHexadecimal(character) {
			return false
		}
	}
	return true
}

func isHexadecimal(character rune) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'f') ||
		(character >= 'A' && character <= 'F')
}
