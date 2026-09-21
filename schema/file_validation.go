package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// FileSchemaResolution controls how absolute schema IDs are resolved from an
// on-disk catalog.
type FileSchemaResolution string

const (
	// FileSchemaResolutionPreferID resolves exact $id values before trying a
	// path suffix. It is the default.
	FileSchemaResolutionPreferID FileSchemaResolution = "prefer-id"
	// FileSchemaResolutionPathOnly resolves absolute references by an
	// unambiguous path suffix and ignores $id values.
	FileSchemaResolutionPathOnly FileSchemaResolution = "path-only"
)

// FileSchemaOptions configures offline resolution for file-backed schemas.
type FileSchemaOptions struct {
	// RefDirs are additional schema catalog roots. The root schema directory is
	// always included.
	RefDirs []string
	// Resolution selects $id or path-suffix lookup. The zero value prefers IDs.
	Resolution FileSchemaResolution
}

// ValidateInstance validates a decoded instance against an in-memory JSON
// schema. References resolve only through the explicitly supplied RefDirs.
func ValidateInstance(schemaData []byte, instance any, options *FileSchemaOptions) ([]Diagnostic, error) {
	resolver, err := newFileSchemaResolver("", options)
	if err != nil {
		return nil, err
	}

	if err := resolver.validateInMemoryReferences(schemaData); err != nil {
		return nil, err
	}
	schemaURL := "memory://schema.json"
	if len(resolver.allowedRoots) > 0 {
		schemaURL = fileURL(filepath.Join(resolver.allowedRoots[0], "__gofulmen_memory_schema__.json"))
	}
	compiler := resolver.compiler()
	if err := compiler.AddResource(schemaURL, bytes.NewReader(schemaData)); err != nil {
		return nil, fmt.Errorf("add in-memory schema: %w", err)
	}
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return (&Validator{schema: compiled}).ValidateData(instance)
}

// ValidateInstanceWithSchemaFile validates a decoded instance against an
// on-disk JSON or YAML schema and resolves references from its offline catalog.
func ValidateInstanceWithSchemaFile(schemaFile string, instance any, options *FileSchemaOptions) ([]Diagnostic, error) {
	resolver, err := newFileSchemaResolver(schemaFile, options)
	if err != nil {
		return nil, err
	}

	compiled, err := resolver.compiler().Compile(fileURL(resolver.schemaFile))
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return (&Validator{schema: compiled}).ValidateData(instance)
}

// ValidateInstanceFile validates a JSON or YAML instance file against an
// on-disk JSON or YAML schema.
func ValidateInstanceFile(schemaFile, instanceFile string, options *FileSchemaOptions) ([]Diagnostic, error) {
	data, err := os.ReadFile(instanceFile) // #nosec G304 -- caller chooses the instance file
	if err != nil {
		return nil, err
	}
	if isJSON(data) {
		var instance any
		if err := json.Unmarshal(data, &instance); err != nil {
			return nil, fmt.Errorf("invalid JSON instance: %w", err)
		}
		return ValidateInstanceWithSchemaFile(schemaFile, instance, options)
	}

	var instance any
	if err := yaml.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("invalid YAML instance: %w", err)
	}
	return ValidateInstanceWithSchemaFile(schemaFile, instance, options)
}

type fileSchemaResolver struct {
	allowedRoots []string
	files        []string
	idIndex      map[string]string
	resolution   FileSchemaResolution
	schemaFile   string
}

func newFileSchemaResolver(schemaFile string, options *FileSchemaOptions) (*fileSchemaResolver, error) {
	resolution := FileSchemaResolutionPreferID
	if options != nil && options.Resolution != "" {
		resolution = options.Resolution
	}
	if resolution != FileSchemaResolutionPreferID && resolution != FileSchemaResolutionPathOnly {
		return nil, fmt.Errorf("unsupported file schema resolution %q", resolution)
	}

	r := &fileSchemaResolver{
		idIndex:    make(map[string]string),
		resolution: resolution,
	}
	if schemaFile != "" {
		canonical, err := canonicalExistingPath(schemaFile)
		if err != nil {
			return nil, fmt.Errorf("resolve schema file: %w", err)
		}
		info, err := os.Stat(canonical)
		if err != nil {
			return nil, fmt.Errorf("stat schema file: %w", err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("schema file is a directory: %s", schemaFile)
		}
		r.schemaFile = canonical
		if err := r.addRoot(filepath.Dir(canonical)); err != nil {
			return nil, err
		}
	}
	if options != nil {
		for _, dir := range options.RefDirs {
			if err := r.addRoot(dir); err != nil {
				return nil, err
			}
		}
	}
	if err := r.indexCatalog(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *fileSchemaResolver) addRoot(dir string) error {
	canonical, err := canonicalExistingPath(dir)
	if err != nil {
		return fmt.Errorf("resolve schema catalog root %q: %w", dir, err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return fmt.Errorf("stat schema catalog root %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("schema catalog root is not a directory: %s", dir)
	}
	for _, root := range r.allowedRoots {
		if root == canonical {
			return nil
		}
	}
	r.allowedRoots = append(r.allowedRoots, canonical)
	return nil
}

func (r *fileSchemaResolver) indexCatalog() error {
	files := make(map[string]struct{})
	for _, root := range r.allowedRoots {
		err := filepath.WalkDir(root, func(candidate string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !isSchemaPath(candidate) {
				return nil
			}
			canonical, err := canonicalExistingPath(candidate)
			if err != nil {
				return err
			}
			if !r.isContained(canonical) {
				return fmt.Errorf("schema path not contained in catalog roots: %s", canonical)
			}
			files[canonical] = struct{}{}
			return nil
		})
		if err != nil {
			return fmt.Errorf("scan schema catalog %s: %w", root, err)
		}
	}
	if r.schemaFile != "" {
		files[r.schemaFile] = struct{}{}
	}
	for file := range files {
		r.files = append(r.files, file)
	}
	sort.Strings(r.files)

	for _, file := range r.files {
		data, err := LoadSchemaFile(file)
		if err != nil {
			return fmt.Errorf("load schema %s: %w", file, err)
		}
		var document any
		if err := json.Unmarshal(data, &document); err != nil {
			return fmt.Errorf("parse schema %s: %w", file, err)
		}
		object, ok := document.(map[string]any)
		if !ok {
			continue
		}
		id, _ := object["$id"].(string)
		id = stripFragment(id)
		if id == "" {
			continue
		}
		if previous, exists := r.idIndex[id]; exists && previous != file {
			return fmt.Errorf("duplicate schema $id %q in catalog: %s and %s", id, previous, file)
		}
		r.idIndex[id] = file
	}
	return nil
}

func (r *fileSchemaResolver) compiler() *jsonschema.Compiler {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	compiler.LoadURL = r.Load
	return compiler
}

// Load implements jsonschema's resource loader without any network fallback.
func (r *fileSchemaResolver) Load(rawURL string) (io.ReadCloser, error) {
	trimmed := stripFragment(rawURL)
	if trimmed == "" {
		return nil, fmt.Errorf("empty schema URL")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid schema URL %q: %w", rawURL, err)
	}

	var schemaPath string
	switch parsed.Scheme {
	case "file":
		schemaPath, err = r.resolveFileURL(parsed)
	case "http", "https":
		schemaPath, err = r.resolveCatalogURL(parsed)
	default:
		err = fmt.Errorf("unsupported schema URI scheme %q", parsed.Scheme)
	}
	if err != nil {
		return nil, err
	}

	data, err := LoadSchemaFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("load schema %s: %w", schemaPath, err)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (r *fileSchemaResolver) resolveFileURL(parsed *url.URL) (string, error) {
	if parsed.Host != "" && !strings.EqualFold(parsed.Host, "localhost") {
		return "", fmt.Errorf("unsupported non-local file host in schema URI: %s", parsed.String())
	}
	if parsed.Path == "" {
		return "", fmt.Errorf("empty file path in schema URI: %s", parsed.String())
	}
	candidate, err := url.PathUnescape(parsed.EscapedPath())
	if err != nil {
		return "", fmt.Errorf("decode file path: %w", err)
	}
	if !filepath.IsAbs(filepath.FromSlash(candidate)) {
		return "", fmt.Errorf("file schema URI must be absolute: %s", parsed.String())
	}
	candidatePath := filepath.Clean(filepath.FromSlash(candidate))
	canonical, err := canonicalExistingPath(candidatePath)
	if err != nil {
		if os.IsNotExist(err) {
			if !r.isLexicallyContained(candidatePath) {
				return "", fmt.Errorf("schema path not contained in catalog roots: %s", candidatePath)
			}
			return r.resolveContainedSuffix(candidatePath)
		}
		return "", fmt.Errorf("resolve schema path %s: %w", candidate, err)
	}
	if !r.isContained(canonical) {
		return "", fmt.Errorf("schema path not contained in catalog roots: %s", canonical)
	}
	return canonical, nil
}

func (r *fileSchemaResolver) resolveCatalogURL(parsed *url.URL) (string, error) {
	key := stripFragment(parsed.String())
	if r.resolution == FileSchemaResolutionPreferID {
		if schemaPath, ok := r.idIndex[key]; ok {
			return schemaPath, nil
		}
	}
	return r.resolveSuffix(parsed)
}

func (r *fileSchemaResolver) resolveSuffix(reference *url.URL) (string, error) {
	suffix, err := safeReferenceSuffix(reference.EscapedPath())
	if err != nil {
		return "", err
	}
	return r.resolveSuffixPath(suffix)
}

func (r *fileSchemaResolver) resolveContainedSuffix(candidate string) (string, error) {
	var suffixes []string
	for _, root := range r.allowedRoots {
		rel, err := filepath.Rel(root, candidate)
		if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
			continue
		}
		suffix, err := safeReferenceSuffix(filepath.ToSlash(rel))
		if err != nil {
			return "", err
		}
		suffixes = append(suffixes, suffix)
	}
	if len(suffixes) == 0 {
		return "", fmt.Errorf("schema path not contained in catalog roots: %s", candidate)
	}

	unique := make(map[string]struct{})
	for _, suffix := range suffixes {
		unique[suffix] = struct{}{}
	}
	var matches []string
	for suffix := range unique {
		match, err := r.resolveSuffixPath(suffix)
		if err == nil {
			matches = append(matches, match)
			continue
		}
		if !strings.Contains(err.Error(), "not found") {
			return "", err
		}
	}
	sort.Strings(matches)
	matches = compactStrings(matches)
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("schema reference not found in offline catalog: %s", candidate)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous schema path suffix for %s: %s", candidate, strings.Join(matches, ", "))
	}
}

func (r *fileSchemaResolver) resolveSuffixPath(suffix string) (string, error) {
	var matches []string
	for _, candidate := range r.files {
		candidatePath := filepath.ToSlash(candidate)
		if strings.HasSuffix(candidatePath, "/"+suffix) || filepath.Base(candidate) == suffix {
			matches = append(matches, candidate)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("schema reference not found in offline catalog: %s", suffix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous schema path suffix %q: %s", suffix, strings.Join(matches, ", "))
	}
}

func safeReferenceSuffix(escaped string) (string, error) {
	decoded, err := url.PathUnescape(escaped)
	if err != nil {
		return "", fmt.Errorf("decode schema path suffix: %w", err)
	}
	cleaned := path.Clean(strings.TrimPrefix(decoded, "/"))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("unsafe schema path suffix %q", decoded)
	}
	return cleaned, nil
}

func (r *fileSchemaResolver) isContained(candidate string) bool {
	for _, root := range r.allowedRoots {
		rel, err := filepath.Rel(root, candidate)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return true
		}
	}
	return false
}

func (r *fileSchemaResolver) isLexicallyContained(candidate string) bool {
	for _, root := range r.allowedRoots {
		rel, err := filepath.Rel(root, candidate)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return true
		}
	}
	return false
}

func (r *fileSchemaResolver) validateInMemoryReferences(schemaData []byte) error {
	var document any
	if err := json.Unmarshal(schemaData, &document); err != nil {
		return fmt.Errorf("parse in-memory schema: %w", err)
	}
	var refs []string
	collectReferences(document, &refs)
	for _, ref := range refs {
		withoutFragment := stripFragment(ref)
		if withoutFragment == "" {
			continue
		}
		parsed, err := url.Parse(withoutFragment)
		if err != nil {
			return fmt.Errorf("invalid schema reference %q: %w", ref, err)
		}
		switch parsed.Scheme {
		case "":
			if len(r.allowedRoots) == 0 {
				return fmt.Errorf("relative schema reference requires a FileSchemaOptions RefDirs entry: %s", ref)
			}
		case "file":
			if len(r.allowedRoots) == 0 {
				return fmt.Errorf("file schema reference requires a FileSchemaOptions RefDirs entry: %s", ref)
			}
		case "http", "https":
			// Absolute HTTP(S) identifiers are offline catalog keys.
		default:
			return fmt.Errorf("unsupported schema URI scheme %q", parsed.Scheme)
		}
	}
	return nil
}

func collectReferences(document any, refs *[]string) {
	switch value := document.(type) {
	case map[string]any:
		if ref, ok := value["$ref"].(string); ok {
			*refs = append(*refs, ref)
		}
		for _, child := range value {
			collectReferences(child, refs)
		}
	case []any:
		for _, child := range value {
			collectReferences(child, refs)
		}
	}
}

func canonicalExistingPath(value string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func isSchemaPath(value string) bool {
	extension := strings.ToLower(filepath.Ext(value))
	return extension == ".json" || extension == ".yaml" || extension == ".yml"
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	output := values[:1]
	for _, value := range values[1:] {
		if value != output[len(output)-1] {
			output = append(output, value)
		}
	}
	return output
}
