// Package main provides a CLI tool to export Crucible schemas with provenance metadata
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/fulmenhq/gofulmen/foundry"
	"github.com/fulmenhq/gofulmen/schema/export"
)

const (
	usageText = `gofulmen-export-schema - Export Crucible schemas with provenance metadata

Usage:
  gofulmen-export-schema --schema-id=<id> --out=<path> [options]

Required Flags:
  --schema-id string
        Crucible schema identifier (e.g., "logging/v1.0.0/config")
  --out string
        Output file path

Optional Flags:
  --format string
        Output format: json|yaml (default: auto-detect from extension)
  --provenance-style string
        Provenance style: object|comment|none (default: object)
  --no-provenance
        Disable provenance metadata inclusion
  --no-validate
        Skip schema validation before export
  --force
        Overwrite existing files without prompting
  --help
        Show this help message

Exit Codes:
  0  - Success
  40 - Invalid arguments (ExitInvalidArgument)
  54 - File write error (ExitFileWriteError)
  60 - Schema validation error (ExitDataInvalid)

Examples:
  # Export logging config schema as JSON
  gofulmen-export-schema \
    --schema-id=logging/v1.0.0/config \
    --out=vendor/crucible/schemas/logging-config.json

  # Export as YAML with comment-style provenance
  gofulmen-export-schema \
    --schema-id=logging/v1.0.0/config \
    --out=schema.yaml \
    --provenance-style=comment

  # Export without provenance
  gofulmen-export-schema \
    --schema-id=logging/v1.0.0/config \
    --out=schema.json \
    --no-provenance
`
)

type cliOptions struct {
	schemaID        string
	outPath         string
	format          string
	provenanceStyle string
	noProvenance    bool
	noValidate      bool
	force           bool
	help            bool
}

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Parse command-line flags
	opts := parseFlags()

	// Show help if requested
	if opts.help {
		fmt.Fprint(os.Stderr, usageText)
		return 0
	}

	// Validate required flags
	if opts.schemaID == "" {
		fmt.Fprintf(os.Stderr, "Error: --schema-id is required\n\n")
		fmt.Fprint(os.Stderr, usageText)
		return foundry.ExitInvalidArgument
	}

	if opts.outPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --out is required\n\n")
		fmt.Fprint(os.Stderr, usageText)
		return foundry.ExitInvalidArgument
	}

	// Build export options
	exportOpts := export.NewExportOptions(opts.schemaID, opts.outPath)

	// Apply CLI flags
	if opts.format != "" {
		switch opts.format {
		case "json":
			exportOpts.Format = export.FormatJSON
		case "yaml", "yml":
			exportOpts.Format = export.FormatYAML
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid format %q (must be json or yaml)\n", opts.format)
			return foundry.ExitInvalidArgument
		}
	}

	if opts.provenanceStyle != "" {
		switch opts.provenanceStyle {
		case "object":
			exportOpts.ProvenanceStyle = export.ProvenanceObject
		case "comment":
			exportOpts.ProvenanceStyle = export.ProvenanceComment
		case "none":
			exportOpts.ProvenanceStyle = export.ProvenanceNone
			exportOpts.IncludeProvenance = false
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid provenance-style %q (must be object, comment, or none)\n", opts.provenanceStyle)
			return foundry.ExitInvalidArgument
		}
	}

	if opts.noProvenance {
		exportOpts.IncludeProvenance = false
	}

	if opts.noValidate {
		exportOpts.ValidateSchema = false
	}

	exportOpts.Overwrite = opts.force

	// Perform the export
	ctx := context.Background()
	if err := export.Export(ctx, exportOpts); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// Map error to appropriate exit code using errors.Is
		switch {
		case errors.Is(err, export.ErrFileExists):
			_, _ = fmt.Fprintf(os.Stderr, "\nHint: Use --force to overwrite existing files\n")
			return foundry.ExitFileWriteError

		case errors.Is(err, export.ErrSchemaNotFound):
			return foundry.ExitConfigInvalid

		case errors.Is(err, export.ErrSchemaValidation):
			return foundry.ExitDataInvalid

		case errors.Is(err, export.ErrPathValidation):
			return foundry.ExitFileWriteError

		case errors.Is(err, export.ErrFileWrite):
			return foundry.ExitFileWriteError

		default:
			// For option validation errors or unknown errors
			return foundry.ExitInvalidArgument
		}
	}

	// Success
	_, _ = fmt.Fprintf(os.Stdout, "Successfully exported schema to: %s\n", opts.outPath)
	return 0
}

func parseFlags() cliOptions {
	opts := cliOptions{}

	flag.StringVar(&opts.schemaID, "schema-id", "", "Crucible schema identifier")
	flag.StringVar(&opts.outPath, "out", "", "Output file path")
	flag.StringVar(&opts.format, "format", "", "Output format (json|yaml)")
	flag.StringVar(&opts.provenanceStyle, "provenance-style", "", "Provenance style (object|comment|none)")
	flag.BoolVar(&opts.noProvenance, "no-provenance", false, "Disable provenance metadata")
	flag.BoolVar(&opts.noValidate, "no-validate", false, "Skip schema validation")
	flag.BoolVar(&opts.force, "force", false, "Overwrite existing files")
	flag.BoolVar(&opts.help, "help", false, "Show help message")

	flag.Parse()

	return opts
}
