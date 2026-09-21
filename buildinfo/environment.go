package buildinfo

import "os"

// ResolveFromProcessEnv applies explicitly requested FULMEN_HOST_* process
// overrides after Resolve. Use Resolve for untrusted launcher environments.
func ResolveFromProcessEnv(ldVersion, ldCommit, ldBuildDate string) Info {
	result := Resolve(ldVersion, ldCommit, ldBuildDate)
	if value := os.Getenv("FULMEN_HOST_VERSION"); !isPlaceholder(value) {
		result.Version = value
	}
	if value := os.Getenv("FULMEN_HOST_COMMIT"); !isPlaceholder(value) {
		result.Commit = value
	}
	if value := os.Getenv("FULMEN_HOST_BUILD_DATE"); !isPlaceholder(value) {
		result.BuildDate = value
	}
	if dirty, ok := parseDirty(os.Getenv("FULMEN_HOST_DIRTY")); ok {
		result.Dirty = &dirty
	}
	if value := os.Getenv("FULMEN_HOST_RUNTIME"); !isPlaceholder(value) {
		result.GoVersion = value
	}
	if value := os.Getenv("FULMEN_HOST_PLATFORM"); !isPlaceholder(value) {
		result.Platform = value
	}
	return result
}
