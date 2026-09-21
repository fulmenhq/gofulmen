package config

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
)

var (
	errUnsupportedConfigValue = errors.New("unsupported config value")
	errConfigValueCycle       = errors.New("cyclic config value")
	jsonNumberPattern         = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)
)

// MergeMaps returns a deep-copied merge of base and overlay.
//
// Objects represented by map[string]any are merged recursively. Arrays
// represented by []any replace the corresponding base value, and an untyped nil
// overlay value deletes its key. A nil overlay map is an empty object; therefore
// MergeMaps(nil, nil) returns a non-nil empty map.
//
// The supported value graph contains map[string]any, []any, nil, string, bool,
// built-in signed and unsigned integer types (except uintptr), finite float32
// and float64 values, and valid json.Number values. Other values and cyclic
// graphs are rejected. On error, neither input is modified and no result is
// returned.
func MergeMaps(base, overlay map[string]any) (map[string]any, error) {
	if err := validateConfigMap(base, make(map[configVisit]bool)); err != nil {
		return nil, err
	}
	if err := validateConfigMap(overlay, make(map[configVisit]bool)); err != nil {
		return nil, err
	}

	return mergeValidatedMaps(base, overlay), nil
}

// DeepCopyMap returns a deep copy of src.
//
// It accepts the same bounded value graph as MergeMaps. Unsupported values and
// cyclic graphs return an error and no result. DeepCopyMap(nil) returns nil.
func DeepCopyMap(src map[string]any) (map[string]any, error) {
	if err := validateConfigMap(src, make(map[configVisit]bool)); err != nil {
		return nil, err
	}

	return copyValidatedMap(src), nil
}

type configVisit struct {
	kind reflect.Kind
	ptr  uintptr
	len  int
	cap  int
}

func validateConfigMap(src map[string]any, active map[configVisit]bool) error {
	if src == nil {
		return nil
	}

	visit := configVisit{kind: reflect.Map, ptr: reflect.ValueOf(src).Pointer()}
	if active[visit] {
		return errConfigValueCycle
	}
	active[visit] = true
	defer delete(active, visit)

	for _, value := range src {
		if err := validateConfigValue(value, active); err != nil {
			return err
		}
	}
	return nil
}

func validateConfigSlice(src []any, active map[configVisit]bool) error {
	if src == nil {
		return nil
	}

	value := reflect.ValueOf(src)
	visit := configVisit{
		kind: reflect.Slice,
		ptr:  value.Pointer(),
		len:  value.Len(),
		cap:  value.Cap(),
	}
	if active[visit] {
		return errConfigValueCycle
	}
	active[visit] = true
	defer delete(active, visit)

	for _, value := range src {
		if err := validateConfigValue(value, active); err != nil {
			return err
		}
	}
	return nil
}

func validateConfigValue(value any, active map[configVisit]bool) error {
	switch typed := value.(type) {
	case nil, string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		if !isFiniteFloat(typed) {
			return errUnsupportedConfigValue
		}
		return nil
	case json.Number:
		if !jsonNumberPattern.MatchString(string(typed)) {
			return errUnsupportedConfigValue
		}
		return nil
	case map[string]any:
		return validateConfigMap(typed, active)
	case []any:
		return validateConfigSlice(typed, active)
	default:
		return errUnsupportedConfigValue
	}
}

func isFiniteFloat(value any) bool {
	switch number := value.(type) {
	case float32:
		return !isNaN32(number) && !isInf32(number)
	case float64:
		return !isNaN64(number) && !isInf64(number)
	default:
		return true
	}
}

func isNaN32(value float32) bool { return value != value }
func isNaN64(value float64) bool { return value != value }
func isInf32(value float32) bool { return value > maxFloat32 || value < -maxFloat32 }
func isInf64(value float64) bool { return value > maxFloat64 || value < -maxFloat64 }

const (
	maxFloat32 = float32(0x1p127 * (2 - 0x1p-23))
	maxFloat64 = float64(0x1p1023 * (2 - 0x1p-52))
)

func mergeValidatedMaps(base, overlay map[string]any) map[string]any {
	result := copyValidatedMap(base)
	if result == nil {
		result = make(map[string]any)
	}

	for key, overlayValue := range overlay {
		if overlayValue == nil {
			delete(result, key)
			continue
		}

		if overlayMap, ok := overlayValue.(map[string]any); ok {
			if baseMap, ok := base[key].(map[string]any); ok {
				result[key] = mergeValidatedMaps(baseMap, overlayMap)
			} else {
				result[key] = mergeValidatedMaps(nil, overlayMap)
			}
			continue
		}

		result[key] = copyValidatedValue(overlayValue)
	}
	return result
}

func copyValidatedMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	result := make(map[string]any, len(src))
	for key, value := range src {
		result[key] = copyValidatedValue(value)
	}
	return result
}

func copyValidatedSlice(src []any) []any {
	if src == nil {
		return nil
	}

	result := make([]any, len(src))
	for index, value := range src {
		result[index] = copyValidatedValue(value)
	}
	return result
}

func copyValidatedValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return copyValidatedMap(typed)
	case []any:
		return copyValidatedSlice(typed)
	default:
		return value
	}
}
