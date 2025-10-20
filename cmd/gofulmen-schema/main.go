package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fulmenhq/gofulmen/schema"
)

const goneatEnv = "GOFULMEN_GONEAT_PATH"

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
	useGoneat := fs.Bool("use-goneat", false, "Use goneat CLI if available (falls back to local validation)")
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

	if *useGoneat {
		if output, err := runGoneatValidate(*schemaID, dataPath, *format); err == nil {
			fmt.Print(output)
			return nil
		} else {
			fmt.Fprintf(os.Stderr, "warning: %v; falling back to local validation\n", err)
		}
	}

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
	useGoneat := fs.Bool("use-goneat", false, "Use goneat CLI if available (falls back to local validation)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("provide exactly one schema file")
	}

	path := fs.Arg(0)

	if *useGoneat {
		if output, err := runGoneatValidateSchema(path, *format); err == nil {
			fmt.Print(output)
			return nil
		} else {
			fmt.Fprintf(os.Stderr, "warning: %v; falling back to local validation\n", err)
		}
	}

	content, err := os.ReadFile(path) // #nosec G304 -- User-provided path is intentional for CLI tool
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

func runGoneatValidate(schemaID, dataPath, format string) (string, error) {
	goneatFormat := mapGoneatFormat(format)
	args := []string{"validate", "data", "--format", goneatFormat, "--data", dataPath}
	if schemaID != "" {
		args = append(args, "--schema", schemaID)
	}
	return runGoneat(args...)
}

func runGoneatValidateSchema(schemaPath, format string) (string, error) {
	goneatFormat := mapGoneatFormat(format)
	args := []string{"schema", "validate-schema", "--format", goneatFormat, schemaPath}
	return runGoneat(args...)
}

func mapGoneatFormat(format string) string {
	switch strings.ToLower(format) {
	case "json":
		return "json"
	default:
		return "markdown"
	}
}

func runGoneat(args ...string) (string, error) {
	binary := os.Getenv(goneatEnv)
	if binary == "" {
		binary = "goneat"
	}

	cmd := exec.Command(binary, args...) // #nosec G204 -- Binary path from env var is expected for goneat integration
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("goneat binary not found (set %s or install goneat)", goneatEnv)
		}
		return "", fmt.Errorf("goneat command failed: %v (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
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
