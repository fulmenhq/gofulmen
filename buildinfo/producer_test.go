package buildinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectStampsUsesExplicitRepository(t *testing.T) {
	mainRepo := createGitRepository(t, "main")
	foreignRepo := createGitRepository(t, "foreign")
	mainCommit := gitOutput(t, mainRepo, "rev-parse", "HEAD")

	for _, variable := range []string{
		"GIT_DIR=" + filepath.Join(foreignRepo, ".git"),
		"GIT_OBJECT_DIRECTORY=" + filepath.Join(foreignRepo, ".git", "objects"),
		"GIT_INDEX_FILE=" + filepath.Join(foreignRepo, ".git", "index"),
		"GIT_COMMON_DIR=" + filepath.Join(foreignRepo, ".git"),
		"GIT_WORK_TREE=" + foreignRepo,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES=" + filepath.Join(foreignRepo, ".git", "objects"),
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_GLOBAL=" + filepath.Join(foreignRepo, "config"),
	} {
		key, value, _ := strings.Cut(variable, "=")
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, value)
			stamps := CollectStamps(mainRepo)
			if stamps.Commit != mainCommit {
				t.Fatalf("CollectStamps with %s = %#v, want own commit %q", key, stamps, mainCommit)
			}
			if _, err := time.Parse(time.RFC3339, stamps.BuildDate); err != nil {
				t.Fatalf("build date %q is not an RFC3339 build timestamp: %v", stamps.BuildDate, err)
			}
		})
	}
}

func TestCollectStampsDetectsTrackedAndUntrackedChanges(t *testing.T) {
	repo := createGitRepository(t, "dirty")
	tracked := filepath.Join(repo, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	foreignRepo := createGitRepository(t, "foreign-index")
	t.Setenv("GIT_INDEX_FILE", filepath.Join(foreignRepo, ".git", "index"))
	stamps := CollectStamps(repo)
	if stamps.Dirty == nil || !*stamps.Dirty {
		t.Fatalf("dirty stamps = %#v, want true", stamps)
	}
}

func TestCollectStampsSupportsDetachedAndPackedReferences(t *testing.T) {
	repo := createGitRepository(t, "geometry")
	commit := gitOutput(t, repo, "rev-parse", "HEAD")
	gitRun(t, repo, "checkout", "--detach", "-q")
	gitRun(t, repo, "pack-refs", "--all", "--prune")
	stamps := CollectStamps(repo)
	if stamps.Commit != commit || stamps.Dirty == nil || *stamps.Dirty {
		t.Fatalf("detached packed stamps = %#v, want clean commit %q", stamps, commit)
	}
}

func TestCollectStampsSupportsLinkedWorktree(t *testing.T) {
	repo := createGitRepository(t, "worktree-main")
	worktree := filepath.Join(t.TempDir(), "linked")
	gitRun(t, repo, "branch", "linked-test")
	gitRun(t, repo, "worktree", "add", "-q", worktree, "linked-test")
	t.Cleanup(func() { gitRun(t, repo, "worktree", "remove", "--force", worktree) })
	commit := gitOutput(t, repo, "rev-parse", "HEAD")
	stamps := CollectStamps(worktree)
	if stamps.Commit != commit || stamps.Dirty == nil || *stamps.Dirty {
		t.Fatalf("linked worktree stamps = %#v, want clean commit %q", stamps, commit)
	}
}

func TestCollectStampsProbeFailureIsUnknown(t *testing.T) {
	stamps := CollectStamps(t.TempDir())
	if stamps.Commit != defaultUnknown || stamps.BuildDate != defaultUnknown || stamps.Dirty != nil {
		t.Fatalf("failed probe = %#v", stamps)
	}
}

func TestSanitizedGitEnvironment(t *testing.T) {
	environment := []string{
		"PATH=/bin",
		"GIT_DIR=/foreign",
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=user.name",
		"GIT_CONFIG_VALUE_0=foreign",
		"GIT_CONFIG_GLOBAL=/foreign/config",
		"GIT_WORK_TREE=/foreign",
	}
	filtered := strings.Join(sanitizedGitEnvironment(environment), "\n")
	if filtered != "PATH=/bin" {
		t.Fatalf("sanitized environment = %q", filtered)
	}
}

func createGitRepository(t *testing.T, name string) string {
	t.Helper()
	directory := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatal(err)
	}
	gitRun(t, directory, "init", "-q")
	gitRun(t, directory, "config", "user.email", "buildinfo@example.test")
	gitRun(t, directory, "config", "user.name", "Build Info Test")
	if err := os.WriteFile(filepath.Join(directory, "tracked.txt"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, directory, "add", "tracked.txt")
	gitRun(t, directory, "commit", "-qm", "initial")
	return directory
}

func gitOutput(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output))
}

func gitRun(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}
