package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// MergeJSONSchemas merges JSON schema documents (supplied as bytes) and returns canonical JSON bytes.
// Later schemas in the list override earlier ones using a deep-merge strategy for objects and
// last-writer-wins for scalars/arrays.
func MergeJSONSchemas(base []byte, overlays ...[]byte) ([]byte, error) {
	accumulator, err := decodeSchemaDocument(base)
	if err != nil {
		return nil, fmt.Errorf("decode base schema: %w", err)
	}

	for i, overlayBytes := range overlays {
		if len(overlayBytes) == 0 {
			continue
		}
		overlay, err := decodeSchemaDocument(overlayBytes)
		if err != nil {
			return nil, fmt.Errorf("decode overlay %d: %w", i, err)
		}
		accumulator = mergeSchemaMap(accumulator, overlay)
	}

	return json.Marshal(accumulator)
}

// ExtendSchema loads the schema identified by baseID from the catalog and applies the
// provided extension map, returning canonical JSON bytes.
func ExtendSchema(catalog *Catalog, baseID string, extension map[string]any) ([]byte, error) {
	if catalog == nil {
		catalog = globalCatalog()
	}
	desc, err := catalog.GetSchema(baseID)
	if err != nil {
		return nil, err
	}
	baseBytes, err := os.ReadFile(desc.Path)
	if err != nil {
		return nil, fmt.Errorf("read base schema: %w", err)
	}
	base, err := decodeSchemaDocument(baseBytes)
	if err != nil {
		return nil, fmt.Errorf("decode base schema: %w", err)
	}
	merged := mergeSchemaMap(base, deepCopy(extension))
	return json.Marshal(merged)
}

// DiffSchemas provides a recursive diff between two schema documents represented as bytes.
// The result is used by Catalog.CompareSchema.
func DiffSchemas(original, other []byte) ([]SchemaDiff, error) {
	left, err := decodeSchemaDocument(original)
	if err != nil {
		return nil, fmt.Errorf("decode left schema: %w", err)
	}
	right, err := decodeSchemaDocument(other)
	if err != nil {
		return nil, fmt.Errorf("decode right schema: %w", err)
	}

	var diffs []SchemaDiff
	diffMaps(&diffs, "", left, right)
	return diffs, nil
}

func decodeSchemaDocument(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return make(map[string]any), nil
	}
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	obj, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema must decode to object")
	}
	return obj, nil
}

func mergeSchemaMap(base map[string]any, overlay map[string]any) map[string]any {
	if base == nil {
		base = make(map[string]any)
	}
	for key, value := range overlay {
		switch ov := value.(type) {
		case map[string]any:
			if existing, ok := base[key].(map[string]any); ok {
				base[key] = mergeSchemaMap(existing, ov)
			} else {
				base[key] = deepCopy(ov)
			}
		case []any:
			base[key] = deepCopySlice(ov)
		default:
			base[key] = ov
		}
	}
	return base
}

func deepCopy(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		switch tv := v.(type) {
		case map[string]any:
			out[k] = deepCopy(tv)
		case []any:
			out[k] = deepCopySlice(tv)
		default:
			out[k] = tv
		}
	}
	return out
}

func deepCopySlice(src []any) []any {
	if src == nil {
		return nil
	}
	out := make([]any, len(src))
	for i, val := range src {
		switch tv := val.(type) {
		case map[string]any:
			out[i] = deepCopy(tv)
		case []any:
			out[i] = deepCopySlice(tv)
		default:
			out[i] = tv
		}
	}
	return out
}

func diffMaps(diffs *[]SchemaDiff, path string, left, right map[string]any) {
	seen := make(map[string]struct{})
	for key := range left {
		seen[key] = struct{}{}
	}
	for key := range right {
		seen[key] = struct{}{}
	}

	for key := range seen {
		subPath := joinPath(path, key)
		lv, lok := left[key]
		rv, rok := right[key]
		switch {
		case !lok:
			*diffs = append(*diffs, SchemaDiff{Path: subPath, Message: "added"})
		case !rok:
			*diffs = append(*diffs, SchemaDiff{Path: subPath, Message: "removed"})
		default:
			compareValues(diffs, subPath, lv, rv)
		}
	}
}

func compareValues(diffs *[]SchemaDiff, path string, left, right any) {
	switch l := left.(type) {
	case map[string]any:
		if r, ok := right.(map[string]any); ok {
			diffMaps(diffs, path, l, r)
			return
		}
	case []any:
		if r, ok := right.([]any); ok {
			if len(l) != len(r) {
				*diffs = append(*diffs, SchemaDiff{Path: path, Message: fmt.Sprintf("array length %d -> %d", len(l), len(r))})
				return
			}
			for i := range l {
				compareValues(diffs, fmt.Sprintf("%s[%d]", path, i), l[i], r[i])
			}
			return
		}
	}
	if !valuesEqual(left, right) {
		*diffs = append(*diffs, SchemaDiff{Path: path, Message: fmt.Sprintf("changed from %v to %v", left, right)})
	}
}

func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func valuesEqual(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}
