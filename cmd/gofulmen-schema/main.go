package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fulmenhq/gofulmen/schema"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "schema":
		runSchemaCommand(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(1)
	}
}

func runSchemaCommand(args []string) {
	if len(args) == 0 {
		schemaUsage()
		os.Exit(1)
	}
	sub := args[0]
	subArgs := args[1:]

	switch sub {
	case "validate":
		if err := schemaValidate(subArgs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "validate-schema":
		if err := schemaValidateSchema(subArgs); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown schema subcommand %q\n", sub)
		schemaUsage()
		os.Exit(1)
	}
}

func schemaValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	schemaID := fs.String("schema-id", "", "Catalog schema identifier (e.g., pathfinder/v1.0.0/path-result)")
	format := fs.String("format", "text", "Output format (text|json)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *schemaID == "" {
		return errors.New("--schema-id is required")
	}
	if fs.NArg() != 1 {
		return errors.New("provide exactly one data file")
	}

	dataPath := fs.Arg(0)
	diags, err := schema.ValidateFileByID(*schemaID, dataPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	switch strings.ToLower(*format) {
	case "json":
		payload := map[string]any{
			"file":        dataPath,
			"schema_id":   *schemaID,
			"valid":       len(diags) == 0,
			"diagnostics": diags,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	default:
		if len(diags) == 0 {
			fmt.Printf("✅ %s valid against %s\n", dataPath, *schemaID)
		} else {
			fmt.Printf("❌ %s invalid against %s\n", dataPath, *schemaID)
			for _, d := range diags {
				fmt.Printf("  - %s (%s): %s\n", d.Pointer, d.Keyword, d.Message)
			}
		}
		return nil
	}
}

func schemaValidateSchema(args []string) error {
	fs := flag.NewFlagSet("validate-schema", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "text", "Output format (text|json)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("provide exactly one schema file")
	}

	path := fs.Arg(0)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	diags, err := schema.ValidateSchemaBytes(content)
	if err != nil {
		return fmt.Errorf("schema compilation failed: %w", err)
	}

	switch strings.ToLower(*format) {
	case "json":
		payload := map[string]any{
			"file":        filepath.Clean(path),
			"valid":       len(diags) == 0,
			"diagnostics": diags,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	default:
		if len(diags) == 0 {
			fmt.Printf("✅ %s schema is valid\n", path)
		} else {
			fmt.Printf("❌ %s schema has issues\n", path)
			for _, d := range diags {
				fmt.Printf("  - %s (%s): %s\n", d.Pointer, d.Keyword, d.Message)
			}
		}
		return nil
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `gofulmen-schema commands:
  schema validate --schema-id <id> <data-file>
  schema validate-schema <schema-file>
`)
}

func schemaUsage() {
	fmt.Fprintf(os.Stderr, `schema commands:
  validate        Validate data against a catalog schema (JSON/YAML).
  validate-schema Validate a schema definition using embedded metaschemas.
`)
}
