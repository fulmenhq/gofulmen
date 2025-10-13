package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func installVerify(tool *Tool) error {
	switch tool.Install.Type {
	case "link", "download":
		// For linked and downloaded tools, check the actual binary path
		binPath := filepath.Join(tool.Install.Destination, tool.Install.BinName)
		if _, err := os.Stat(binPath); err != nil {
			return fmt.Errorf("binary not found at %s: %w", binPath, err)
		}
		// Also verify it's executable
		if _, err := exec.LookPath(binPath); err != nil {
			return fmt.Errorf("binary at %s is not executable: %w", binPath, err)
		}

	case "verify":
		// For verify type, check command in PATH
		cmd := tool.Install.Command
		_, err := exec.LookPath(cmd)
		if err != nil {
			suggestion := fmt.Sprintf("Install %s and ensure it's in your PATH", cmd)

			switch cmd {
			case "git":
				suggestion = "Install Git: https://git-scm.com/downloads"
			case "curl":
				suggestion = "Install curl via your package manager (apt, brew, etc.)"
			case "wget":
				suggestion = "Install wget via your package manager (apt, brew, etc.)"
			}

			return &CommandNotFoundError{
				Command:    cmd,
				Suggestion: suggestion,
			}
		}

	default:
		return fmt.Errorf("unsupported install type for verification: %s", tool.Install.Type)
	}

	return nil
}
