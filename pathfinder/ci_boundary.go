package pathfinder

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CIBoundaryHint represents a safe boundary derived from a CI environment.
//
// The boundary is intended to be used with FindRepositoryRoot via WithBoundary().
//
// Important security property:
// - A CI boundary hint must never broaden traversal.
// - The returned Boundary always contains startPath.
// - The returned Boundary is never the filesystem root.
//
// Consumers should still treat env-derived hints as CI-only inputs.
// For production code paths, prefer explicit configuration rather than env overrides.
type CIBoundaryHint struct {
	// Boundary is an absolute path used as an upward traversal ceiling.
	Boundary string

	// Provider is a best-effort identifier like "github", "gitlab", "jenkins", or "generic".
	Provider string

	// Source is the environment variable used (e.g. "GITHUB_WORKSPACE").
	Source string

	// Reason is a short, non-sensitive explanation safe for logs.
	Reason string
}

// DetectCIBoundaryHint attempts to derive a safe traversal boundary from common
// CI environment variables.
//
// It returns (hint, true) when a valid boundary is found; otherwise (zero, false).
//
// The returned hint is safe by construction:
// - candidate boundary must exist and be a directory
// - candidate boundary must be an ancestor of startPath
// - candidate boundary must not be the filesystem root
func DetectCIBoundaryHint(startPath string) (CIBoundaryHint, bool) {
	absStart, ok := normalizeStartPath(startPath)
	if !ok {
		return CIBoundaryHint{}, false
	}

	fsRoot := filepath.Clean(getFilesystemRoot(absStart))

	provider, candidates := ciCandidates()
	for _, c := range candidates {
		boundary, ok := validateBoundaryCandidate(absStart, fsRoot, c.Value)
		if !ok {
			continue
		}

		return CIBoundaryHint{
			Boundary: boundary,
			Provider: provider,
			Source:   c.Source,
			Reason:   c.Reason,
		}, true
	}

	return CIBoundaryHint{}, false
}

type ciCandidate struct {
	Source string
	Value  string
	Reason string
}

func ciCandidates() (string, []ciCandidate) {
	// Prefer a Fulmen-standard variable if a forge chooses to set it.
	// This is only treated as a boundary hint, never an unconditional root.
	fulmenRoot := strings.TrimSpace(os.Getenv("FULMEN_WORKSPACE_ROOT"))

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "github", []ciCandidate{
			{Source: "FULMEN_WORKSPACE_ROOT", Value: fulmenRoot, Reason: "ci:github+fulmen"},
			{Source: "GITHUB_WORKSPACE", Value: strings.TrimSpace(os.Getenv("GITHUB_WORKSPACE")), Reason: "ci:github"},
		}
	}

	if os.Getenv("GITLAB_CI") == "true" {
		return "gitlab", []ciCandidate{
			{Source: "FULMEN_WORKSPACE_ROOT", Value: fulmenRoot, Reason: "ci:gitlab+fulmen"},
			{Source: "CI_PROJECT_DIR", Value: strings.TrimSpace(os.Getenv("CI_PROJECT_DIR")), Reason: "ci:gitlab"},
		}
	}

	if strings.TrimSpace(os.Getenv("JENKINS_URL")) != "" {
		return "jenkins", []ciCandidate{
			{Source: "FULMEN_WORKSPACE_ROOT", Value: fulmenRoot, Reason: "ci:jenkins+fulmen"},
			{Source: "WORKSPACE", Value: strings.TrimSpace(os.Getenv("WORKSPACE")), Reason: "ci:jenkins"},
		}
	}

	if os.Getenv("CI") == "true" {
		return "generic", []ciCandidate{
			{Source: "FULMEN_WORKSPACE_ROOT", Value: fulmenRoot, Reason: "ci:generic+fulmen"},
			{Source: "GITHUB_WORKSPACE", Value: strings.TrimSpace(os.Getenv("GITHUB_WORKSPACE")), Reason: "ci:generic"},
			{Source: "CI_PROJECT_DIR", Value: strings.TrimSpace(os.Getenv("CI_PROJECT_DIR")), Reason: "ci:generic"},
			{Source: "WORKSPACE", Value: strings.TrimSpace(os.Getenv("WORKSPACE")), Reason: "ci:generic"},
		}
	}

	// Not in CI (no strong signal).
	return "", nil
}

func normalizeStartPath(startPath string) (string, bool) {
	if strings.TrimSpace(startPath) == "" {
		return "", false
	}

	absStart, err := filepath.Abs(startPath)
	if err != nil {
		return "", false
	}

	st, err := os.Stat(absStart)
	if err != nil {
		return "", false
	}

	if !st.IsDir() {
		absStart = filepath.Dir(absStart)
	}

	return filepath.Clean(absStart), true
}

func validateBoundaryCandidate(absStart string, fsRoot string, candidate string) (string, bool) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return "", false
	}

	candidate = filepath.Clean(candidate)
	if !filepath.IsAbs(candidate) {
		return "", false
	}

	st, err := os.Stat(candidate)
	if err != nil || !st.IsDir() {
		return "", false
	}

	absBoundary, err := filepath.Abs(candidate)
	if err != nil {
		return "", false
	}

	absBoundary = filepath.Clean(absBoundary)
	if absBoundary == fsRoot {
		return "", false
	}

	if runtime.GOOS != "windows" {
		if absBoundary == "/" {
			return "", false
		}
	}

	if !isWithinBoundary(absStart, absBoundary) {
		return "", false
	}

	return absBoundary, true
}
