package buildinfo

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/crucible"
)

func TestResolvePriority(t *testing.T) {
	buildInfo := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
			{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	got := resolveWithBuildInfo("dev", "unknown", "unknown", "", "9.9.9", buildInfo)
	if got.Version != "1.2.3" || got.Commit != "abcdef1234567890" || got.BuildDate != "2026-01-02T03:04:05Z" {
		t.Fatalf("build-info fallback = %#v", got)
	}
	if got.Dirty == nil || !*got.Dirty {
		t.Fatalf("dirty = %#v, want true", got.Dirty)
	}
	if got.GoVersion != runtime.Version() || got.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Fatalf("runtime identity = %#v", got)
	}

	got = resolveWithBuildInfo("2.0.0", "feedface", "2026-02-03T04:05:06Z", "false", "9.9.9", buildInfo)
	if got.Version != "2.0.0" || got.Commit != "feedface" || got.BuildDate != "2026-02-03T04:05:06Z" {
		t.Fatalf("ldflags must win: %#v", got)
	}
}

func TestResolveBuildInfoFallbacks(t *testing.T) {
	tests := []struct {
		name    string
		build   *debug.BuildInfo
		embed   string
		version string
		dirty   *bool
	}{
		{
			name:    "devel uses embed",
			build:   &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			embed:   "v1.2.3",
			version: "1.2.3",
		},
		{
			name:    "no metadata uses dev",
			version: "dev",
		},
		{
			name: "strict false dirty",
			build: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.modified", Value: "false"},
			}},
			version: "dev",
			dirty:   boolPointer(false),
		},
		{
			name: "invalid dirty is unknown",
			build: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.modified", Value: "yes"},
			}},
			version: "dev",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := resolveWithBuildInfo("dev", "unknown", "unknown", "", test.embed, test.build)
			if got.Version != test.version {
				t.Fatalf("version = %q, want %q", got.Version, test.version)
			}
			if !sameBoolPointer(got.Dirty, test.dirty) {
				t.Fatalf("dirty = %#v, want %#v", got.Dirty, test.dirty)
			}
		})
	}
}

func TestResolveDoesNotUseProcessEnvironment(t *testing.T) {
	t.Setenv("FULMEN_HOST_VERSION", "forged-version")
	t.Setenv("FULMEN_HOST_COMMIT", "forged-commit")
	t.Setenv("FULMEN_HOST_BUILD_DATE", "forged-date")
	t.Setenv("FULMEN_HOST_DIRTY", "true")
	t.Setenv("FULMEN_HOST_RUNTIME", "forged-runtime")
	t.Setenv("FULMEN_HOST_PLATFORM", "forged/platform")

	got := Resolve("stamped-version", "stamped-commit", "2026-01-02T03:04:05Z")
	if got.Version != "stamped-version" || got.Commit != "stamped-commit" || got.BuildDate != "2026-01-02T03:04:05Z" {
		t.Fatalf("Resolve accepted process environment: %#v", got)
	}
	fromEnvironment := ResolveFromProcessEnv("stamped-version", "stamped-commit", "2026-01-02T03:04:05Z")
	if fromEnvironment.Version != "forged-version" || fromEnvironment.Commit != "forged-commit" || fromEnvironment.Dirty == nil || !*fromEnvironment.Dirty || fromEnvironment.GoVersion != "forged-runtime" || fromEnvironment.Platform != "forged/platform" {
		t.Fatalf("ResolveFromProcessEnv did not apply explicit environment: %#v", fromEnvironment)
	}
}

func TestResolveWithoutStampsDoesNotUseProcessEnvironment(t *testing.T) {
	t.Setenv("FULMEN_HOST_VERSION", "forged-version")
	t.Setenv("FULMEN_HOST_COMMIT", "forged-commit")
	t.Setenv("FULMEN_HOST_BUILD_DATE", "forged-date")
	t.Setenv("FULMEN_HOST_RUNTIME", "forged-runtime")
	t.Setenv("FULMEN_HOST_PLATFORM", "forged/platform")

	got := Resolve("dev", "unknown", "unknown")
	for field, value := range map[string]string{
		"version":    got.Version,
		"commit":     got.Commit,
		"build date": got.BuildDate,
		"runtime":    got.GoVersion,
		"platform":   got.Platform,
	} {
		if strings.Contains(value, "forged") {
			t.Fatalf("unstamped Resolve accepted process %s: %#v", field, got)
		}
	}
}

func TestResolveWithDirtyOverridesBuildInfo(t *testing.T) {
	buildInfo := &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "false"}}}
	got := resolveWithBuildInfo("dev", "unknown", "unknown", "true", "", buildInfo)
	if got.Dirty == nil || !*got.Dirty {
		t.Fatalf("dirty override = %#v, want true", got.Dirty)
	}
	got = resolveWithBuildInfo("dev", "unknown", "unknown", "invalid", "", nil)
	if got.Dirty != nil {
		t.Fatalf("invalid dirty stamp = %#v", got.Dirty)
	}
	for _, value := range []string{"TRUE", "False", "TrUe"} {
		got = resolveWithBuildInfo("dev", "unknown", "unknown", value, "", nil)
		if got.Dirty != nil {
			t.Fatalf("non-strict dirty stamp %q = %#v", value, got.Dirty)
		}
	}
}

func TestShortCommitPreservesNonHexValues(t *testing.T) {
	if got := shortCommit("diagnostic-value-not-a-sha"); got != "diagnostic-value-not-a-sha" {
		t.Fatalf("shortCommit changed non-SHA value: %q", got)
	}
	if got := shortCommit("abcdef1234567890abcdef1234567890abcdef12"); got != "abcdef1" {
		t.Fatalf("shortCommit did not shorten SHA: %q", got)
	}
}

func TestInfoFormattingAndJSON(t *testing.T) {
	dirty := true
	info := Info{
		Version:   "1.2.3",
		Commit:    "abcdef1234567890",
		BuildDate: "2026-01-02T03:04:05Z",
		Dirty:     &dirty,
		GoVersion: "go1.test",
		Platform:  "test/test",
	}
	if got := info.FormatBasic("tool"); got != "tool 1.2.3" {
		t.Fatalf("FormatBasic = %q", got)
	}
	extended := info.FormatExtended("tool", &crucible.Version{Gofulmen: "0.3.6", Crucible: "0.2.0"})
	for _, expected := range []string{"Commit: abcdef1", "BuildDate:", "Dirty: true", "Runtime: go1.test", "Platform: test/test", "Pins:", "gofulmen: 0.3.6", "Crucible: 0.2.0"} {
		if !strings.Contains(extended, expected) {
			t.Fatalf("FormatExtended missing %q: %s", expected, extended)
		}
	}
	encoded, err := info.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !strings.Contains(string(encoded), "abcdef1234567890") {
		t.Fatalf("JSON shortened commit: %s", encoded)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if decoded["buildDate"] != info.BuildDate || decoded["platform"] != info.Platform || decoded["runtime"] != info.GoVersion {
		t.Fatalf("unexpected JSON fields: %v", decoded)
	}
}

func boolPointer(value bool) *bool { return &value }

func sameBoolPointer(left, right *bool) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
