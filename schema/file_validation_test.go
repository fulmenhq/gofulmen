package schema

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestValidateInstanceWithSchemaFile_OfflineCatalog(t *testing.T) {
	dir := t.TempDir()
	rootDir := filepath.Join(dir, "root")
	refDir := filepath.Join(dir, "catalog")
	mustMkdirAll(t, rootDir)
	mustMkdirAll(t, filepath.Join(refDir, "schemas"))

	root := filepath.Join(rootDir, "root.schema.json")
	mustWriteFile(t, root, `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://schemas.example.test/root.schema.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "widget"],
  "properties": {
    "name": {"type": "string"},
    "widget": {"$ref": "https://schemas.example.test/widgets/v1/widget.schema.json"}
  }
}`)
	mustWriteFile(t, filepath.Join(refDir, "schemas", "widget.schema.json"), `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://schemas.example.test/widgets/v1/widget.schema.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["kind"],
  "properties": {"kind": {"const": "ok"}}
}`)

	opts := &FileSchemaOptions{RefDirs: []string{refDir}}
	valid := map[string]any{"name": "name", "widget": map[string]any{"kind": "ok"}}
	diags, err := ValidateInstanceWithSchemaFile(root, valid, opts)
	if err != nil {
		t.Fatalf("ValidateInstanceWithSchemaFile returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}

	invalid := map[string]any{"widget": map[string]any{"kind": "wrong"}, "extra": true}
	diags, err = ValidateInstanceWithSchemaFile(root, invalid, opts)
	if err != nil {
		t.Fatalf("invalid instance returned compile error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected validation diagnostics")
	}
	for _, diagnostic := range diags {
		if diagnostic.Severity != SeverityError || diagnostic.Source != sourceGoFulmen {
			t.Fatalf("unexpected diagnostic taxonomy: %#v", diagnostic)
		}
	}
	assertDiagnosticKeyword(t, diags, "required")
	assertDiagnosticKeyword(t, diags, "additionalProperties")
	assertDiagnosticKeyword(t, diags, "const")
}

func TestValidateInstanceWithSchemaFile_PathOnly(t *testing.T) {
	dir := t.TempDir()
	rootDir := filepath.Join(dir, "root")
	refDir := filepath.Join(dir, "catalog")
	mustMkdirAll(t, rootDir)
	mustMkdirAll(t, filepath.Join(refDir, "schemas"))

	root := filepath.Join(rootDir, "root.schema.json")
	mustWriteFile(t, root, `{
  "$ref": "https://schemas.example.test/schemas/a/../widget.schema.json"
}`)
	mustWriteFile(t, filepath.Join(refDir, "schemas", "widget.schema.json"), `{
  "$id": "https://different.example.test/not-used.json",
  "type": "string"
}`)

	diags, err := ValidateInstanceWithSchemaFile(root, "ok", &FileSchemaOptions{
		RefDirs:    []string{refDir},
		Resolution: FileSchemaResolutionPathOnly,
	})
	if err != nil {
		t.Fatalf("PathOnly returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestValidateInstanceWithSchemaFile_RelativeReferenceInRefDir(t *testing.T) {
	dir := t.TempDir()
	rootDir := filepath.Join(dir, "root")
	refDir := filepath.Join(dir, "catalog")
	mustMkdirAll(t, rootDir)
	mustMkdirAll(t, filepath.Join(refDir, "shared"))

	root := filepath.Join(rootDir, "root.schema.json")
	mustWriteFile(t, root, `{"$ref":"shared/widget.schema.json"}`)
	mustWriteFile(t, filepath.Join(refDir, "shared", "widget.schema.json"), `{"type":"string"}`)

	diags, err := ValidateInstanceWithSchemaFile(root, "ok", &FileSchemaOptions{RefDirs: []string{refDir}})
	if err != nil {
		t.Fatalf("relative reference in RefDirs returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestValidateInstanceAndInstanceFile(t *testing.T) {
	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "root.schema.json")
	jsonInstance := filepath.Join(dir, "instance.json")
	yamlInstance := filepath.Join(dir, "instance.yaml")
	mustWriteFile(t, schemaFile, `{"type":"object","required":["name"]}`)
	mustWriteFile(t, jsonInstance, `{"name":"ok"}`)
	mustWriteFile(t, yamlInstance, "name: ok\n")

	for _, instanceFile := range []string{jsonInstance, yamlInstance} {
		diags, err := ValidateInstanceFile(schemaFile, instanceFile, nil)
		if err != nil || len(diags) != 0 {
			t.Fatalf("ValidateInstanceFile(%s) = %v, %v", instanceFile, diags, err)
		}
	}

	diags, err := ValidateInstance([]byte(`{"type":"string"}`), "ok", nil)
	if err != nil || len(diags) != 0 {
		t.Fatalf("ValidateInstance = %v, %v", diags, err)
	}
	if _, err := ValidateInstance([]byte(`{"$ref":"sibling.schema.json"}`), "ok", nil); err == nil {
		t.Fatal("expected relative in-memory reference to fail")
	}

	refDir := filepath.Join(dir, "catalog")
	mustMkdirAll(t, refDir)
	sibling := filepath.Join(refDir, "sibling.schema.json")
	mustWriteFile(t, sibling, `{"type":"string"}`)
	opts := &FileSchemaOptions{RefDirs: []string{refDir}}
	for _, schemaData := range [][]byte{
		[]byte(`{"$ref":"sibling.schema.json"}`),
		[]byte(`{"$ref":` + quoteJSON(fileURL(sibling)) + `}`),
	} {
		diags, err := ValidateInstance(schemaData, "ok", opts)
		if err != nil || len(diags) != 0 {
			t.Fatalf("in-memory catalog reference = %v, %v", diags, err)
		}
	}
}

func TestValidateInstanceWithSchemaFile_Drafts(t *testing.T) {
	for _, schema := range []string{
		`{"type":"object","properties":{"kind":{"const":"ok"}},"required":["kind"]}`,
		`{"$schema":"https://json-schema.org/draft-07/schema","type":"object","properties":{"kind":{"const":"ok"}},"required":["kind"]}`,
	} {
		dir := t.TempDir()
		root := filepath.Join(dir, "root.schema.json")
		mustWriteFile(t, root, schema)
		diags, err := ValidateInstanceWithSchemaFile(root, map[string]any{"kind": "ok"}, nil)
		if err != nil || len(diags) != 0 {
			t.Fatalf("draft schema = %v, %v", diags, err)
		}
	}
}

func TestValidateInstanceWithSchemaFile_RejectsUnsafeReferences(t *testing.T) {
	dir := t.TempDir()
	rootDir := filepath.Join(dir, "catalog")
	outside := filepath.Join(dir, "outside.schema.json")
	mustMkdirAll(t, rootDir)
	mustWriteFile(t, outside, `{"type":"string"}`)

	for name, ref := range map[string]string{
		"traversal":     "../outside.schema.json",
		"scheme":        "ftp://schemas.example.test/widget.schema.json",
		"remote_file":   "file://files.example.test/widget.schema.json",
		"offline_https": "https://unreachable.example.test/widget.schema.json",
	} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(rootDir, name+".schema.json")
			mustWriteFile(t, root, `{"$ref":`+quoteJSON(ref)+`}`)
			if _, err := ValidateInstanceWithSchemaFile(root, "value", nil); err == nil {
				t.Fatalf("expected %s reference to fail", name)
			}
		})
	}

	symlink := filepath.Join(rootDir, "escape.schema.json")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	root := filepath.Join(rootDir, "symlink-root.schema.json")
	mustWriteFile(t, root, `{"$ref":"escape.schema.json"}`)
	if _, err := ValidateInstanceWithSchemaFile(root, "value", nil); err == nil {
		t.Fatal("expected outside-root symlink reference to fail")
	}
}

func TestFileSchemaResolverRejectsDuplicateIDsAndAmbiguousSuffixes(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "one.schema.json"), `{"$id":"https://schemas.example.test/duplicate","type":"string"}`)
	mustWriteFile(t, filepath.Join(dir, "two.schema.json"), `{"$id":"https://schemas.example.test/duplicate","type":"string"}`)
	if _, err := newFileSchemaResolver("", &FileSchemaOptions{RefDirs: []string{dir}}); err == nil {
		t.Fatal("expected duplicate ID failure")
	}

	dir = t.TempDir()
	mustMkdirAll(t, filepath.Join(dir, "one"))
	mustMkdirAll(t, filepath.Join(dir, "two"))
	mustWriteFile(t, filepath.Join(dir, "one", "shared.schema.json"), `{"type":"string"}`)
	mustWriteFile(t, filepath.Join(dir, "two", "shared.schema.json"), `{"type":"string"}`)
	root := filepath.Join(dir, "root.schema.json")
	mustWriteFile(t, root, `{"$ref":"https://schemas.example.test/shared.schema.json"}`)
	if _, err := ValidateInstanceWithSchemaFile(root, "value", nil); err == nil {
		t.Fatal("expected ambiguous suffix failure")
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func quoteJSON(value string) string {
	return strconv.Quote(value)
}

func assertDiagnosticKeyword(t *testing.T, diags []Diagnostic, keyword string) {
	t.Helper()
	for _, diagnostic := range diags {
		if strings.Contains(diagnostic.Keyword, keyword) {
			return
		}
	}
	t.Fatalf("missing %q diagnostic: %v", keyword, diags)
}
