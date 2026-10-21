package ciworkflow

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	requiredGoDirective     = "1.26.10"
	requiredGoneat          = "v0.6.5"
	requiredGolangciLint    = "v2.14.0"
	requiredGolangciLintPkg = "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0"
	requiredGosecPkg        = "github.com/securego/gosec/v2/cmd/gosec@v2.29.0"
	requiredRunnerImage     = "ghcr.io/fulmenhq/goneat-tools-runner-glibc:v0.5.12@sha256:616a4b5361d2cf7d32ac32496d58e9c288660d5606f0a27efdbf56e3a462ab2d"
	requiredGateScript      = "./scripts/verify-runner-toolchain.sh"
	requiredGoToken         = "go1.26.10 "
)

type workflowFile struct {
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Container *workflowContainer `yaml:"container"`
	Steps     []workflowStep     `yaml:"steps"`
}

type workflowContainer struct {
	Image string `yaml:"image"`
}

type workflowStep struct {
	Name string            `yaml:"name"`
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	With map[string]string `yaml:"with"`
}

type toolsFile struct {
	Tools map[string]toolsEntry `yaml:"tools"`
}

type toolsEntry struct {
	InstallPackage string `yaml:"install_package"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root not found from test file")
		}
		dir = parent
	}
}

func TestCIWorkflowPins(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	var wf workflowFile
	if err := yaml.Unmarshal(raw, &wf); err != nil {
		t.Fatalf("parse ci.yml: %v", err)
	}

	testJob, ok := wf.Jobs["test"]
	if !ok {
		t.Fatal("ci.yml missing jobs.test")
	}
	if testJob.Container == nil {
		t.Fatal("jobs.test missing container")
	}
	if testJob.Container.Image != requiredRunnerImage {
		t.Fatalf("jobs.test container image=%q, want %q", testJob.Container.Image, requiredRunnerImage)
	}

	var sawGate bool
	for _, step := range testJob.Steps {
		if step.Name != "Verify runner toolchain" {
			continue
		}
		sawGate = true
		if strings.TrimSpace(step.Run) != requiredGateScript {
			t.Fatalf("Verify runner toolchain run=%q, want %q", step.Run, requiredGateScript)
		}
	}
	if !sawGate {
		t.Fatal("jobs.test missing Verify runner toolchain step")
	}

	installJob, ok := wf.Jobs["install-test"]
	if !ok {
		t.Fatal("ci.yml missing jobs.install-test")
	}
	var sawSetupGo bool
	for _, step := range installJob.Steps {
		if !strings.Contains(step.Uses, "actions/setup-go@") {
			continue
		}
		sawSetupGo = true
		if step.With["go-version"] != requiredGoDirective {
			t.Fatalf("install-test setup-go go-version=%q, want %q", step.With["go-version"], requiredGoDirective)
		}
	}
	if !sawSetupGo {
		t.Fatal("install-test missing setup-go")
	}
}

func TestGoModFloorHasNoToolchainLine(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "\ngo "+requiredGoDirective+"\n") {
		t.Fatalf("go.mod missing go %s directive", requiredGoDirective)
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "toolchain ") {
			t.Fatalf("go.mod must not set toolchain; found %q", line)
		}
	}
}

func TestMakefileGoneatPinAndNoGOTOOLCHAIN(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "GONEAT_VERSION ?= "+requiredGoneat) {
		t.Fatalf("Makefile missing GONEAT_VERSION ?= %s", requiredGoneat)
	}
	if !strings.Contains(body, "GOLANGCI_LINT_VERSION ?= "+requiredGolangciLint) {
		t.Fatalf("Makefile missing GOLANGCI_LINT_VERSION ?= %s", requiredGolangciLint)
	}
	if strings.Contains(body, "GOTOOLCHAIN") {
		t.Fatal("Makefile must not wire GOTOOLCHAIN")
	}
}

func TestGosecPinnedExact(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".goneat", "tools.yaml"))
	if err != nil {
		t.Fatalf("read tools.yaml: %v", err)
	}
	var cfg toolsFile
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse tools.yaml: %v", err)
	}
	gosec, ok := cfg.Tools["gosec"]
	if !ok {
		t.Fatal("tools.yaml missing gosec")
	}
	if gosec.InstallPackage != requiredGosecPkg {
		t.Fatalf("gosec install_package=%q, want %q", gosec.InstallPackage, requiredGosecPkg)
	}
	if strings.Contains(gosec.InstallPackage, "@latest") {
		t.Fatal("gosec must not use @latest")
	}

	golangci, ok := cfg.Tools["golangci-lint"]
	if !ok {
		t.Fatal("tools.yaml missing golangci-lint")
	}
	if golangci.InstallPackage != requiredGolangciLintPkg {
		t.Fatalf("golangci-lint install_package=%q, want %q", golangci.InstallPackage, requiredGolangciLintPkg)
	}
}

func TestRunnerToolchainGateExactGo(t *testing.T) {
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "verify-runner-toolchain.sh")
	raw, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("read gate script: %v", err)
	}
	if !bytes.Contains(raw, []byte(requiredGoToken)) {
		t.Fatalf("gate script must require exact token %q", requiredGoToken)
	}

	cases := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{name: "exact 1.26.10", line: "go version go1.26.10 linux/amd64", wantErr: false},
		{name: "1.26.9", line: "go version go1.26.9 linux/amd64", wantErr: true},
		{name: "1.27.3", line: "go version go1.27.3 linux/amd64", wantErr: true},
		{name: "1.26.11", line: "go version go1.26.11 linux/amd64", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fake := filepath.Join(dir, "go")
			content := "#!/bin/sh\necho '" + tc.line + "'\n"
			if err := os.WriteFile(fake, []byte(content), 0o755); err != nil {
				t.Fatalf("write fake go: %v", err)
			}
			cmd := exec.Command(script)
			cmd.Env = append(os.Environ(),
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"VERIFY_RUNNER_TOOLCHAIN_QUIET=1",
			)
			out, err := cmd.CombinedOutput()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected failure for %q, output:\n%s", tc.line, out)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected success for %q: %v\n%s", tc.line, err, out)
			}
		})
	}
}
