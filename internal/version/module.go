package version

import (
	"runtime/debug"
	"strings"
)

// ModuleVersion returns the semantic module version for modulePath when available
// via runtime build info.
//
// It returns an empty string when the version is unknown (for example, when
// running from source with an unversioned main module).
func ModuleVersion(modulePath string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return ""
	}

	if info.Main.Path == modulePath {
		if v := normalizeBuildInfoVersion(info.Main.Version); v != "" {
			return v
		}
	}

	for _, dep := range info.Deps {
		if dep == nil || dep.Path != modulePath {
			continue
		}
		if v := normalizeBuildInfoVersion(dep.Version); v != "" {
			return v
		}
	}

	return ""
}

func normalizeBuildInfoVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "(devel)" {
		return ""
	}
	return v
}
