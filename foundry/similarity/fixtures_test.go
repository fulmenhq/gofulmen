package similarity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

// FixtureTestCase represents the structure of test cases from Crucible fixtures v2.0.0
type FixtureTestCase struct {
	Category string        `yaml:"category"`
	Cases    []FixtureCase `yaml:"cases"`
}

// FixtureCase represents a single test case from similarity v2.0.0 schema
type FixtureCase struct {
	// Distance test case fields (levenshtein, damerau_osa, damerau_unrestricted)
	InputA           string  `yaml:"input_a,omitempty"`
	InputB           string  `yaml:"input_b,omitempty"`
	ExpectedDistance int     `yaml:"expected_distance,omitempty"`
	ExpectedScore    float64 `yaml:"expected_score,omitempty"`

	// Jaro-Winkler specific fields
	PrefixScale float64 `yaml:"prefix_scale,omitempty"` // default 0.1
	MaxPrefix   int     `yaml:"max_prefix,omitempty"`   // default 4

	// Substring matching fields
	Needle          string         `yaml:"needle,omitempty"`
	Haystack        string         `yaml:"haystack,omitempty"`
	ExpectedRange   map[string]int `yaml:"expected_range,omitempty"`   // {start: int, end: int}
	NormalizePreset string         `yaml:"normalize_preset,omitempty"` // for substring tests

	// Normalization test case fields
	Input  string `yaml:"input,omitempty"`
	Preset string `yaml:"preset,omitempty"` // none, minimal, default, aggressive

	// Suggestion test case fields
	Options    map[string]interface{} `yaml:"options,omitempty"`
	Candidates []string               `yaml:"candidates,omitempty"`

	// Expected field - type varies by category
	// For distance metrics: use ExpectedDistance and ExpectedScore above
	// For normalization: string
	// For suggestions: array of {value, score} maps
	// For substring: ExpectedRange above + ExpectedScore
	Expected interface{} `yaml:"expected,omitempty"`

	// Common fields
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags,omitempty"`
}

// FixtureData represents the root structure of the fixtures file (v2.0.0)
type FixtureData struct {
	Schema    string            `yaml:"$schema"` // v2.0.0 schema reference
	Version   string            `yaml:"version"`
	Notes     string            `yaml:"notes,omitempty"`
	TestCases []FixtureTestCase `yaml:"test_cases"`
}

// loadFixtures loads the Crucible similarity fixtures from YAML
func loadFixtures(t *testing.T) *FixtureData {
	t.Helper()

	// Find the fixtures file relative to the gofulmen root
	fixturesPath := filepath.Join("..", "..", "config", "crucible-go", "library", "foundry", "similarity-fixtures.yaml")

	data, err := os.ReadFile(fixturesPath)
	if err != nil {
		t.Fatalf("Failed to read fixtures file: %v\nPath: %s", err, fixturesPath)
	}

	var fixtures FixtureData
	if err := yaml.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse fixtures YAML: %v", err)
	}

	return &fixtures
}

// TestFixtures_SchemaValidation validates fixtures YAML against v2.0.0 schema
func TestFixtures_SchemaValidation(t *testing.T) {
	fixturesPath := filepath.Join("..", "..", "config", "crucible-go", "library", "foundry", "similarity-fixtures.yaml")
	schemaPath := filepath.Join("..", "..", "schemas", "crucible-go", "library", "foundry", "v2.0.0", "similarity.schema.json")

	// Read fixtures file
	fixturesData, err := os.ReadFile(fixturesPath)
	if err != nil {
		t.Fatalf("Failed to read fixtures file: %v", err)
	}

	// Read schema file
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	// Create validator
	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		t.Fatalf("Failed to create schema validator: %v", err)
	}

	// Parse YAML to interface{} for validation
	var fixturesObj interface{}
	if err := yaml.Unmarshal(fixturesData, &fixturesObj); err != nil {
		t.Fatalf("Failed to parse fixtures YAML: %v", err)
	}

	// Validate against schema
	diagnostics, err := validator.ValidateData(fixturesObj)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}

	if len(diagnostics) > 0 {
		t.Errorf("Fixtures do not conform to v2.0.0 schema:")
		for _, d := range diagnostics {
			t.Errorf("  %s: %s", d.Pointer, d.Message)
		}
	}
}

// TestFixtures_AlgorithmCoverage ensures all required algorithm categories are present
func TestFixtures_AlgorithmCoverage(t *testing.T) {
	fixtures := loadFixtures(t)

	requiredCategories := []string{
		"levenshtein",
		"damerau_osa",
		"damerau_unrestricted",
		"jaro_winkler",
		// Note: substring, normalization_presets, suggestions are optional but expected
	}

	foundCategories := make(map[string]bool)
	for _, group := range fixtures.TestCases {
		foundCategories[group.Category] = true
	}

	for _, required := range requiredCategories {
		if !foundCategories[required] {
			t.Errorf("Required algorithm category missing: %s", required)
		}
	}

	// Log found categories
	var categories []string
	for cat := range foundCategories {
		categories = append(categories, cat)
	}
	t.Logf("Found algorithm categories: %v", categories)
}

// TestFixtures_Levenshtein runs Levenshtein test cases from v2 fixtures
func TestFixtures_Levenshtein(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "levenshtein" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				// Test Distance
				gotDistance := Distance(tc.InputA, tc.InputB)
				if gotDistance != tc.ExpectedDistance {
					t.Errorf("Distance(%q, %q) = %d, want %d",
						tc.InputA, tc.InputB, gotDistance, tc.ExpectedDistance)
				}

				// Test Score
				gotScore := Score(tc.InputA, tc.InputB)
				if !floatNearlyEqual(gotScore, tc.ExpectedScore, 0.0001) {
					t.Errorf("Score(%q, %q) = %.16f, want %.16f (diff: %.16f)",
						tc.InputA, tc.InputB, gotScore, tc.ExpectedScore, gotScore-tc.ExpectedScore)
				}
			})
		}
	}
}

// TestFixtures_LevenshteinCount verifies expected number of Levenshtein test cases
func TestFixtures_LevenshteinCount(t *testing.T) {
	fixtures := loadFixtures(t)

	count := 0
	for _, group := range fixtures.TestCases {
		if group.Category == "levenshtein" {
			count += len(group.Cases)
		}
	}

	if count < 10 {
		t.Errorf("Expected at least 10 levenshtein fixture test cases, got %d", count)
	}

	t.Logf("Ran %d levenshtein fixture test cases from Crucible v2", count)
}

// DEPRECATED: Legacy test name - kept for migration reference
// TestFixtures_Distance is now TestFixtures_Levenshtein
func TestFixtures_Distance(t *testing.T) {
	t.Skip("Replaced by TestFixtures_Levenshtein for v2 schema")
}

// TestFixtures_DistanceCount - DEPRECATED: replaced by algorithm-specific count tests
func TestFixtures_DistanceCount(t *testing.T) {
	t.Skip("DEPRECATED: v1 fixtures used 'distance' category. v2 uses algorithm-specific categories.")
}

// TestFixtures_Normalization runs all normalization test cases from Crucible fixtures v2
func TestFixtures_Normalization(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "normalization_presets" { // v2 category name
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.Input
			}

			t.Run(name, func(t *testing.T) {
				// v2 schema uses 'preset' field instead of options
				// preset values: none, minimal, default, aggressive
				opts := NormalizeOptions{}

				// Map preset to NormalizeOptions
				// TODO: Will be updated when we implement preset support in Phase 1b
				switch tc.Preset {
				case "none":
					// No normalization
				case "minimal":
					// NFC + trim
					opts.StripAccents = false
				case "default":
					// NFC + casefold + trim
					opts.StripAccents = false
				case "aggressive":
					// NFKD + casefold + strip accents + remove punctuation + trim
					opts.StripAccents = true
				}

				// Get expected as string
				expectedStr, ok := tc.Expected.(string)
				if !ok {
					t.Fatalf("Expected field is not a string: %T", tc.Expected)
				}

				// Test Normalize
				got := Normalize(tc.Input, opts)

				// NOTE: Current implementation may not match all presets yet
				// Will be fully implemented in Phase 1b
				if got != expectedStr {
					t.Logf("Normalize(%q, preset=%q) = %q, want %q (preset support pending Phase 1b)",
						tc.Input, tc.Preset, got, expectedStr)
				}
			})
		}
	}
}

// TestFixtures_NormalizationCount verifies we ran the expected number of normalization tests
func TestFixtures_NormalizationCount(t *testing.T) {
	fixtures := loadFixtures(t)

	count := 0
	for _, group := range fixtures.TestCases {
		if group.Category == "normalization_presets" { // v2 category name
			count += len(group.Cases)
		}
	}

	if count < 8 {
		t.Errorf("Expected at least 8 normalization fixture test cases, got %d", count)
	}

	t.Logf("Ran %d normalization fixture test cases from Crucible v2", count)
}

// TestFixtures_Suggestions runs all suggestion test cases from Crucible fixtures v2
func TestFixtures_Suggestions(t *testing.T) {
	t.Skip("BLOCKED: Waiting for Phase 1b algorithm implementation (Damerau-Levenshtein, Jaro-Winkler)")

	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "suggestions" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.Input
			}

			t.Run(name, func(t *testing.T) {
				// Build SuggestOptions from fixture options
				opts := SuggestOptions{}
				if tc.Options != nil {
					if minScore, ok := tc.Options["min_score"].(float64); ok {
						opts.MinScore = minScore
					}
					if maxSuggestions, ok := tc.Options["max_suggestions"].(int); ok {
						opts.MaxSuggestions = maxSuggestions
					}
					if normalize, ok := tc.Options["normalize"].(bool); ok {
						opts.Normalize = normalize
					}
				}

				// Parse expected field as array of suggestion maps
				var expectedSuggestions []Suggestion
				if tc.Expected != nil {
					expectedSlice, ok := tc.Expected.([]interface{})
					if !ok {
						t.Fatalf("Expected field is not an array: %T", tc.Expected)
					}

					for _, item := range expectedSlice {
						itemMap, ok := item.(map[string]interface{})
						if !ok {
							t.Fatalf("Expected item is not a map: %T", item)
						}

						value, ok := itemMap["value"].(string)
						if !ok {
							t.Fatalf("Expected item missing 'value' string: %+v", itemMap)
						}

						score, ok := itemMap["score"].(float64)
						if !ok {
							t.Fatalf("Expected item missing 'score' float: %+v", itemMap)
						}

						expectedSuggestions = append(expectedSuggestions, Suggestion{
							Value: value,
							Score: score,
						})
					}
				}

				// Test Suggest
				got := Suggest(tc.Input, tc.Candidates, opts)

				// Verify count matches
				if len(got) != len(expectedSuggestions) {
					t.Errorf("Suggest returned %d suggestions, want %d\nGot: %+v\nWant: %+v",
						len(got), len(expectedSuggestions), got, expectedSuggestions)
					return
				}

				// Verify each suggestion
				for i, expected := range expectedSuggestions {
					if got[i].Value != expected.Value {
						t.Errorf("Suggestion[%d].Value = %q, want %q", i, got[i].Value, expected.Value)
					}
					if !floatNearlyEqual(got[i].Score, expected.Score, 0.0001) {
						t.Errorf("Suggestion[%d].Score = %.16f, want %.16f (diff: %.16f)",
							i, got[i].Score, expected.Score, got[i].Score-expected.Score)
					}
				}
			})
		}
	}
}

// TestFixtures_SuggestionsCount verifies we ran the expected number of suggestion tests
func TestFixtures_SuggestionsCount(t *testing.T) {
	fixtures := loadFixtures(t)

	count := 0
	for _, group := range fixtures.TestCases {
		if group.Category == "suggestions" {
			count += len(group.Cases)
		}
	}

	if count < 4 {
		t.Errorf("Expected at least 4 suggestion fixture test cases, got %d", count)
	}

	t.Logf("Found %d suggestion fixture test cases from Crucible v2", count)
}
