package buildinfo

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// Stamps are build-time values suitable for caller linker variables. Unknown
// fields mean the producer could not safely read the main module repository.
type Stamps struct {
	Commit    string
	BuildDate string
	Dirty     *bool
}

// CollectStamps reads Git metadata for an explicit main-module directory. It
// removes Git redirect variables before every subprocess so an invoking
// environment cannot redirect the probe to another repository.
func CollectStamps(moduleDir string) Stamps {
	commit, ok := runGit(moduleDir, "rev-parse", "HEAD")
	if !ok {
		return Stamps{Commit: defaultUnknown, BuildDate: defaultUnknown}
	}
	status, statusOK := runGit(moduleDir, "status", "--porcelain", "--untracked-files=all")
	stamps := Stamps{Commit: commit, BuildDate: time.Now().UTC().Format(time.RFC3339)}
	if statusOK {
		dirty := strings.TrimSpace(status) != ""
		stamps.Dirty = &dirty
	}
	return stamps
}

func runGit(moduleDir string, args ...string) (string, bool) {
	command := exec.Command("git", args...) // #nosec G204 -- arguments are fixed by CollectStamps
	command.Dir = moduleDir
	command.Env = sanitizedGitEnvironment(os.Environ())
	output, err := command.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(output)), true
}

func sanitizedGitEnvironment(environment []string) []string {
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		key, _, _ := strings.Cut(entry, "=")
		if isGitRedirectVariable(key) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func isGitRedirectVariable(key string) bool {
	switch key {
	case "GIT_DIR", "GIT_OBJECT_DIRECTORY", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_WORK_TREE", "GIT_ALTERNATE_OBJECT_DIRECTORIES":
		return true
	default:
		return strings.HasPrefix(key, "GIT_CONFIG")
	}
}
