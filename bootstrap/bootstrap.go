package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	ManifestPath string
	Force        bool
	Verbose      bool
}

func InstallTools(opts Options) error {
	if opts.ManifestPath == "" {
		opts.ManifestPath = ".goneat/tools.yaml"
	}

	manifestPath := resolveManifestPath(opts.ManifestPath)

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	platform := GetPlatform()

	if supported, msg := IsPlatformSupported(platform); !supported {
		return fmt.Errorf("unsupported platform: %s - %s", platform, msg)
	} else if msg != "" && opts.Verbose {
		fmt.Fprintf(os.Stderr, "⚠️  %s\n", msg)
	}

	if opts.Verbose {
		fmt.Printf("Installing tools for %s...\n", platform)
		fmt.Printf("Manifest: %s\n", manifestPath)
		fmt.Printf("Tools: %d\n\n", len(manifest.Tools))
	}

	var errors []error
	successCount := 0

	for _, tool := range manifest.Tools {
		if opts.Verbose {
			fmt.Printf("📦 %s (%s)...", tool.ID, tool.Install.Type)
		}

		err := installTool(&tool, platform, opts)
		if err != nil {
			if opts.Verbose {
				fmt.Printf(" ❌\n")
			}
			errors = append(errors, fmt.Errorf("%s: %w", tool.ID, err))

			if tool.Required {
				if opts.Verbose {
					fmt.Printf("\n❌ Required tool %s failed to install\n", tool.ID)
				}
				break
			}
		} else {
			if opts.Verbose {
				fmt.Printf(" ✅\n")
			}
			successCount++
		}
	}

	if len(errors) > 0 {
		if opts.Verbose {
			fmt.Printf("\n")
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			}
		}
		return fmt.Errorf("failed to install %d tool(s)", len(errors))
	}

	if opts.Verbose {
		fmt.Printf("\n✅ Successfully installed %d tool(s)\n", successCount)
	}

	return nil
}

func VerifyTools(opts Options) error {
	if opts.ManifestPath == "" {
		opts.ManifestPath = ".goneat/tools.yaml"
	}

	manifestPath := resolveManifestPath(opts.ManifestPath)

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	if opts.Verbose {
		fmt.Printf("Verifying tools...\n")
		fmt.Printf("Manifest: %s\n\n", manifestPath)
	}

	var errors []error

	for _, tool := range manifest.Tools {
		if opts.Verbose {
			fmt.Printf("🔍 %s...", tool.ID)
		}

		err := verifyTool(&tool)
		if err != nil {
			if opts.Verbose {
				fmt.Printf(" ❌\n")
			}
			errors = append(errors, fmt.Errorf("%s: %w", tool.ID, err))
		} else {
			if opts.Verbose {
				fmt.Printf(" ✅\n")
			}
		}
	}

	if len(errors) > 0 {
		if opts.Verbose {
			fmt.Printf("\n")
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "Missing: %v\n", err)
			}
		}
		return fmt.Errorf("%d tool(s) not available", len(errors))
	}

	if opts.Verbose {
		fmt.Printf("\n✅ All tools verified\n")
	}

	return nil
}

func installTool(tool *Tool, platform Platform, opts Options) error {
	switch tool.Install.Type {
	case "verify":
		return installVerify(tool)

	case "go":
		return installGo(tool)

	case "download":
		return installDownload(tool, platform)

	case "link":
		return installLink(tool)

	default:
		return fmt.Errorf("unsupported install type: %s", tool.Install.Type)
	}
}

func verifyTool(tool *Tool) error {
	return installVerify(tool)
}

func resolveManifestPath(defaultPath string) string {
	dir := filepath.Dir(defaultPath)
	base := filepath.Base(defaultPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	localPath := filepath.Join(dir, nameWithoutExt+".local"+ext)

	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	return defaultPath
}
