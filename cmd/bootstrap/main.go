package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fulmenhq/gofulmen/bootstrap"
)

func main() {
	var (
		install      = flag.Bool("install", false, "Install tools from manifest")
		verify       = flag.Bool("verify", false, "Verify tools are available")
		manifestPath = flag.String("manifest", ".goneat/tools.yaml", "Path to tools manifest")
		force        = flag.Bool("force", false, "Force reinstall even if exists")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		help         = flag.Bool("help", false, "Show usage information")
	)

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	if !*install && !*verify {
		fmt.Fprintf(os.Stderr, "Error: must specify --install or --verify\n\n")
		printUsage()
		os.Exit(1)
	}

	opts := bootstrap.Options{
		ManifestPath: *manifestPath,
		Force:        *force,
		Verbose:      *verbose,
	}

	var err error

	if *install {
		err = bootstrap.InstallTools(opts)
	} else if *verify {
		err = bootstrap.VerifyTools(opts)
	}

	if err != nil {
		if !*verbose {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Bootstrap - Simple tool installation for Go repositories

Usage:
  go run github.com/fulmenhq/gofulmen/cmd/bootstrap [options]

Options:
  --install            Install tools from manifest
  --verify             Verify tools are available
  --manifest <path>    Path to tools manifest (default: .goneat/tools.yaml)
  --force              Force reinstall even if exists
  --verbose            Verbose output
  --help               Show this help message

Examples:
  # Install all tools from manifest
  go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install

  # Verify all tools are available
  go run github.com/fulmenhq/gofulmen/cmd/bootstrap --verify

  # Custom manifest path
  go run github.com/fulmenhq/gofulmen/cmd/bootstrap --manifest /path/to/tools.yaml --install

  # Verbose output
  go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install --verbose

Platform Support:
  ✅ macOS (arm64, amd64)
  ✅ Linux (arm64, amd64)
  ⚠️  Windows (requires tar and gzip installed via scoop)

For more information, see: https://github.com/fulmenhq/gofulmen/tree/main/bootstrap`)
}
