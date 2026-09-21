package config_test

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/config"
)

func TestMergeMapsRecursivelyMergesWithoutAliasingInputs(t *testing.T) {
	base := map[string]any{
		"keep": "base",
		"object": map[string]any{
			"base":   "kept",
			"shared": []any{map[string]any{"source": "base"}},
		},
		"array": []any{"base"},
	}
	overlay := map[string]any{
		"object": map[string]any{
			"overlay": []any{map[string]any{"source": "overlay"}},
		},
		"array": []any{"overlay"},
	}

	merged, err := config.MergeMaps(base, overlay)
	if err != nil {
		t.Fatalf("MergeMaps returned error: %v", err)
	}

	want := map[string]any{
		"keep": "base",
		"object": map[string]any{
			"base":    "kept",
			"shared":  []any{map[string]any{"source": "base"}},
			"overlay": []any{map[string]any{"source": "overlay"}},
		},
		"array": []any{"overlay"},
	}
	if !reflect.DeepEqual(merged, want) {
		t.Fatalf("MergeMaps result = %#v, want %#v", merged, want)
	}

	mergedObject := merged["object"].(map[string]any)
	mergedObject["base"] = "changed"
	mergedObject["shared"].([]any)[0].(map[string]any)["source"] = "changed"
	mergedObject["overlay"].([]any)[0].(map[string]any)["source"] = "changed"
	merged["array"].([]any)[0] = "changed"

	if got := base["object"].(map[string]any)["base"]; got != "kept" {
		t.Fatalf("base object was mutated: %v", got)
	}
	if got := base["object"].(map[string]any)["shared"].([]any)[0].(map[string]any)["source"]; got != "base" {
		t.Fatalf("base nested value was mutated: %v", got)
	}
	if got := overlay["object"].(map[string]any)["overlay"].([]any)[0].(map[string]any)["source"]; got != "overlay" {
		t.Fatalf("overlay nested value was mutated: %v", got)
	}
	if got := overlay["array"].([]any)[0]; got != "overlay" {
		t.Fatalf("overlay array was mutated: %v", got)
	}
}

func TestMergeMapsDeletesWithUntypedNilAndHandlesTypedNilContainers(t *testing.T) {
	var nilMap map[string]any
	var nilSlice []any
	base := map[string]any{
		"delete": map[string]any{"nested": "value"},
		"object": map[string]any{"keep": "value"},
		"slice":  []any{"value"},
	}
	overlay := map[string]any{
		"delete": nil,
		"object": nilMap,
		"slice":  nilSlice,
	}

	merged, err := config.MergeMaps(base, overlay)
	if err != nil {
		t.Fatalf("MergeMaps returned error: %v", err)
	}
	if _, ok := merged["delete"]; ok {
		t.Fatal("untyped nil overlay did not delete key")
	}
	if got := merged["object"].(map[string]any)["keep"]; got != "value" {
		t.Fatalf("typed-nil map did not merge as empty object: %v", got)
	}
	if slice, ok := merged["slice"].([]any); !ok || slice != nil {
		t.Fatalf("typed-nil slice = %#v, want typed nil []any", merged["slice"])
	}

	merged, err = config.MergeMaps(nil, map[string]any{"object": nilMap})
	if err != nil {
		t.Fatalf("MergeMaps returned error: %v", err)
	}
	if object, ok := merged["object"].(map[string]any); !ok || object == nil || len(object) != 0 {
		t.Fatalf("typed-nil object on an absent key = %#v, want non-nil empty object", merged["object"])
	}
}

func TestMergeMapsNilInputsAndNestedDeletion(t *testing.T) {
	merged, err := config.MergeMaps(nil, nil)
	if err != nil {
		t.Fatalf("MergeMaps(nil, nil) returned error: %v", err)
	}
	if merged == nil || len(merged) != 0 {
		t.Fatalf("MergeMaps(nil, nil) = %#v, want non-nil empty map", merged)
	}

	merged, err = config.MergeMaps(
		map[string]any{"object": map[string]any{"remove": "value", "keep": "value"}},
		map[string]any{"object": map[string]any{"remove": nil, "added": "value"}},
	)
	if err != nil {
		t.Fatalf("MergeMaps returned error: %v", err)
	}
	want := map[string]any{"object": map[string]any{"keep": "value", "added": "value"}}
	if !reflect.DeepEqual(merged, want) {
		t.Fatalf("MergeMaps result = %#v, want %#v", merged, want)
	}
}

func TestDeepCopyMapDoesNotAliasAndPreservesNil(t *testing.T) {
	var nilMap map[string]any
	var nilSlice []any
	src := map[string]any{
		"map":      map[string]any{"slice": []any{map[string]any{"value": "original"}}},
		"nilMap":   nilMap,
		"nilSlice": nilSlice,
	}

	copy, err := config.DeepCopyMap(src)
	if err != nil {
		t.Fatalf("DeepCopyMap returned error: %v", err)
	}
	copy["map"].(map[string]any)["slice"].([]any)[0].(map[string]any)["value"] = "changed"
	if got := src["map"].(map[string]any)["slice"].([]any)[0].(map[string]any)["value"]; got != "original" {
		t.Fatalf("source was mutated: %v", got)
	}
	if value, ok := copy["nilMap"].(map[string]any); !ok || value != nil {
		t.Fatalf("copied nil map = %#v", copy["nilMap"])
	}
	if value, ok := copy["nilSlice"].([]any); !ok || value != nil {
		t.Fatalf("copied nil slice = %#v", copy["nilSlice"])
	}

	copy, err = config.DeepCopyMap(nil)
	if err != nil || copy != nil {
		t.Fatalf("DeepCopyMap(nil) = %#v, %v; want nil, nil", copy, err)
	}
}

func TestConfigPrimitivesPreserveSupportedScalars(t *testing.T) {
	values := map[string]any{
		"string":  "value",
		"bool":    true,
		"int":     int(-1),
		"int8":    int8(-1),
		"int16":   int16(-1),
		"int32":   int32(-1),
		"int64":   int64(-1),
		"uint":    uint(1),
		"uint8":   uint8(1),
		"uint16":  uint16(1),
		"uint32":  uint32(1),
		"uint64":  uint64(1),
		"float32": float32(1.5),
		"float64": 1.5,
		"number":  json.Number("123.45e-6"),
	}

	copy, err := config.DeepCopyMap(values)
	if err != nil {
		t.Fatalf("DeepCopyMap returned error: %v", err)
	}
	if !reflect.DeepEqual(copy, values) {
		t.Fatalf("DeepCopyMap result = %#v, want %#v", copy, values)
	}
}

func TestConfigPrimitivesRejectUnsupportedValuesWithoutLeakingValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "named type", value: namedString("sensitive-value")},
		{name: "typed map", value: map[string]string{"key": "sensitive-value"}},
		{name: "typed slice", value: []string{"sensitive-value"}},
		{name: "pointer", value: new(string)},
		{name: "uintptr", value: uintptr(1)},
		{name: "invalid number", value: json.Number("not-a-number")},
		{name: "infinite float", value: math.Inf(1)},
		{name: "nan float", value: math.NaN()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			src := map[string]any{"bad": test.value}
			copy, err := config.DeepCopyMap(src)
			if err == nil || copy != nil {
				t.Fatalf("DeepCopyMap = %#v, %v; want nil, error", copy, err)
			}
			if strings.Contains(err.Error(), "sensitive-value") {
				t.Fatalf("error leaked input value: %q", err)
			}
			if len(src) != 1 || reflect.TypeOf(src["bad"]) != reflect.TypeOf(test.value) {
				t.Fatalf("input was modified: %#v", src)
			}
		})
	}
}

func TestConfigPrimitivesRejectCyclesAndPermitSharedChildren(t *testing.T) {
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	if copy, err := config.DeepCopyMap(cyclicMap); err == nil || copy != nil {
		t.Fatalf("DeepCopyMap(cyclic map) = %#v, %v; want nil, error", copy, err)
	}

	cyclicSlice := make([]any, 1)
	cyclicSlice[0] = cyclicSlice
	if copy, err := config.DeepCopyMap(map[string]any{"slice": cyclicSlice}); err == nil || copy != nil {
		t.Fatalf("DeepCopyMap(cyclic slice) = %#v, %v; want nil, error", copy, err)
	}

	shared := map[string]any{"value": "shared"}
	copy, err := config.DeepCopyMap(map[string]any{"first": shared, "second": shared})
	if err != nil {
		t.Fatalf("DeepCopyMap(shared child) returned error: %v", err)
	}
	first := copy["first"].(map[string]any)
	second := copy["second"].(map[string]any)
	if reflect.ValueOf(first).Pointer() == reflect.ValueOf(second).Pointer() {
		t.Fatal("DeepCopyMap preserved mutable alias topology")
	}
}

type namedString string

func TestMergeMapsValidatesBothGraphsBeforeReturningAResult(t *testing.T) {
	base := map[string]any{"nested": map[string]any{"value": "base"}}
	overlay := map[string]any{
		"nested": map[string]any{"value": "overlay"},
		"bad":    []byte("sensitive-value"),
	}

	merged, err := config.MergeMaps(base, overlay)
	if err == nil || merged != nil {
		t.Fatalf("MergeMaps = %#v, %v; want nil, error", merged, err)
	}
	if strings.Contains(err.Error(), "sensitive-value") {
		t.Fatalf("error leaked input value: %q", err)
	}
	if got := base["nested"].(map[string]any)["value"]; got != "base" {
		t.Fatalf("base was mutated: %v", got)
	}
	if got := overlay["nested"].(map[string]any)["value"]; got != "overlay" {
		t.Fatalf("overlay was mutated: %v", got)
	}
}

func TestDeepCopyMapAcceptsAliasesOfSupportedTypes(t *testing.T) {
	type stringAlias = string
	type intAlias = int

	copy, err := config.DeepCopyMap(map[string]any{
		"string": stringAlias("value"),
		"int":    intAlias(1),
	})
	if err != nil {
		t.Fatalf("DeepCopyMap returned error: %v", err)
	}
	if copy["string"] != "value" || copy["int"] != int(1) {
		t.Fatalf("DeepCopyMap result = %#v", copy)
	}
}
