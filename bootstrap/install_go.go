package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func installGo(tool *Tool) error {
	if _, err := exec.LookPath("go"); err != nil {
		return &CommandNotFoundError{
			Command:    "go",
			Suggestion: "Install Go from https://go.dev/dl/",
		}
	}

	moduleVersion := fmt.Sprintf("%s@%s", tool.Install.Module, tool.Install.Version)

	cmd := exec.Command("go", "install", moduleVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", moduleVersion, err)
	}

	binName := filepath.Base(tool.Install.Module)
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			gopath = filepath.Join(home, "go")
		}
	}

	if gopath != "" {
		binPath := filepath.Join(gopath, "bin", binName)
		if _, err := os.Stat(binPath); err == nil {
			return nil
		}
	}

	if _, err := exec.LookPath(binName); err != nil {
		return fmt.Errorf("installed %s but cannot find in PATH - ensure GOPATH/bin or GOBIN is in your PATH", binName)
	}

	return nil
}
