package signals

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// SignalResolutionFixtures represents the test fixture file structure.
type SignalResolutionFixtures struct {
	Description           string                 `yaml:"description"`
	Version               string                 `yaml:"version"`
	ResolveSignalTests    []ResolveSignalTest    `yaml:"resolve_signal_tests"`
	ListSignalNamesTests  []ListSignalNamesTest  `yaml:"list_signal_names_tests"`
	MatchSignalNamesTests []MatchSignalNamesTest `yaml:"match_signal_names_tests"`
}

// ResolveSignalTest represents a single resolve_signal test case.
type ResolveSignalTest struct {
	Input       string  `yaml:"input"`
	ExpectName  *string `yaml:"expect_name"`
	Description string  `yaml:"description"`
}

// ListSignalNamesTest represents a single list_signal_names test case.
type ListSignalNamesTest struct {
	Description    string   `yaml:"description"`
	ExpectContains []string `yaml:"expect_contains"`
	ExpectMinCount int      `yaml:"expect_min_count"`
}

// MatchSignalNamesTest represents a single match_signal_names test case.
type MatchSignalNamesTest struct {
	Pattern                     string   `yaml:"pattern"`
	Description                 string   `yaml:"description"`
	ExpectContains              []string `yaml:"expect_contains"`
	ExpectNotContains           []string `yaml:"expect_not_contains"`
	ExpectCount                 *int     `yaml:"expect_count"`
	ExpectMinCount              *int     `yaml:"expect_min_count"`
	ExpectEmpty                 bool     `yaml:"expect_empty"`
	ExpectEqualsListSignalNames bool     `yaml:"expect_equals_list_signal_names"`
}

// loadFixtures loads the signal resolution test fixtures from the synced Crucible config.
func loadFixtures(t *testing.T) *SignalResolutionFixtures {
	t.Helper()

	data, err := os.ReadFile("../../config/crucible-go/library/foundry/signal-resolution-fixtures.yaml")
	require.NoError(t, err, "Failed to load signal resolution fixtures")

	var fixtures SignalResolutionFixtures
	err = yaml.Unmarshal(data, &fixtures)
	require.NoError(t, err, "Failed to parse signal resolution fixtures")

	return &fixtures
}

// TestResolveSignal_Fixtures runs all fixture test cases for ResolveSignal.
func TestResolveSignal_Fixtures(t *testing.T) {
	fixtures := loadFixtures(t)
	catalog := NewCatalog()

	for _, tc := range fixtures.ResolveSignalTests {
		t.Run(tc.Description, func(t *testing.T) {
			result := catalog.ResolveSignal(tc.Input)

			if tc.ExpectName == nil {
				assert.Nil(t, result, "Expected nil for input %q", tc.Input)
			} else {
				require.NotNil(t, result, "Expected non-nil result for input %q", tc.Input)
				assert.Equal(t, *tc.ExpectName, result.Name, "Signal name mismatch for input %q", tc.Input)
			}
		})
	}
}

// TestListSignalNames_Fixtures runs all fixture test cases for ListSignalNames.
func TestListSignalNames_Fixtures(t *testing.T) {
	fixtures := loadFixtures(t)
	catalog := NewCatalog()

	for _, tc := range fixtures.ListSignalNamesTests {
		t.Run(tc.Description, func(t *testing.T) {
			names, err := catalog.ListSignalNames()
			require.NoError(t, err, "ListSignalNames should not return an error")

			// Check minimum count
			if tc.ExpectMinCount > 0 {
				assert.GreaterOrEqual(t, len(names), tc.ExpectMinCount,
					"Expected at least %d signal names", tc.ExpectMinCount)
			}

			// Check required signals are present
			nameSet := make(map[string]bool)
			for _, name := range names {
				nameSet[name] = true
			}

			for _, expected := range tc.ExpectContains {
				assert.True(t, nameSet[expected],
					"Expected signal %q to be in list", expected)
			}
		})
	}
}

// TestMatchSignalNames_Fixtures runs all fixture test cases for MatchSignalNames.
func TestMatchSignalNames_Fixtures(t *testing.T) {
	fixtures := loadFixtures(t)
	catalog := NewCatalog()

	for _, tc := range fixtures.MatchSignalNamesTests {
		t.Run(tc.Description, func(t *testing.T) {
			matches, err := catalog.MatchSignalNames(tc.Pattern)
			require.NoError(t, err, "MatchSignalNames should not return an error")

			// Check if result should equal ListSignalNames
			if tc.ExpectEqualsListSignalNames {
				allNames, err := catalog.ListSignalNames()
				require.NoError(t, err)
				assert.ElementsMatch(t, allNames, matches,
					"Pattern %q should match all signal names", tc.Pattern)
				return
			}

			// Check empty expectation
			if tc.ExpectEmpty {
				assert.Empty(t, matches, "Pattern %q should match no signals", tc.Pattern)
				return
			}

			// Check exact count
			if tc.ExpectCount != nil {
				assert.Len(t, matches, *tc.ExpectCount,
					"Pattern %q should match exactly %d signals", tc.Pattern, *tc.ExpectCount)
			}

			// Check minimum count
			if tc.ExpectMinCount != nil {
				assert.GreaterOrEqual(t, len(matches), *tc.ExpectMinCount,
					"Pattern %q should match at least %d signals", tc.Pattern, *tc.ExpectMinCount)
			}

			// Check required matches
			matchSet := make(map[string]bool)
			for _, m := range matches {
				matchSet[m] = true
			}

			for _, expected := range tc.ExpectContains {
				assert.True(t, matchSet[expected],
					"Pattern %q should match %q", tc.Pattern, expected)
			}

			// Check exclusions
			for _, excluded := range tc.ExpectNotContains {
				assert.False(t, matchSet[excluded],
					"Pattern %q should NOT match %q", tc.Pattern, excluded)
			}
		})
	}
}

// TestResolveSignal_AdditionalCases tests edge cases not covered by fixtures.
func TestResolveSignal_AdditionalCases(t *testing.T) {
	catalog := NewCatalog()

	t.Run("returns same instance as strict lookup", func(t *testing.T) {
		resolved := catalog.ResolveSignal("SIGTERM")
		strict, err := catalog.GetSignalByName("SIGTERM")
		require.NoError(t, err)
		assert.Same(t, strict, resolved, "ResolveSignal should return same instance as GetSignalByName")
	})

	t.Run("handles tab and newline whitespace", func(t *testing.T) {
		result := catalog.ResolveSignal("\t\nSIGTERM\t\n")
		require.NotNil(t, result)
		assert.Equal(t, "SIGTERM", result.Name)
	})

	t.Run("zero is not a valid signal number", func(t *testing.T) {
		result := catalog.ResolveSignal("0")
		assert.Nil(t, result, "Signal number 0 should not resolve")
	})
}

// TestListSignalNames tests basic functionality.
func TestListSignalNames(t *testing.T) {
	catalog := NewCatalog()

	names, err := catalog.ListSignalNames()
	require.NoError(t, err)
	assert.NotEmpty(t, names, "Should return at least one signal name")

	// All names should start with SIG
	for _, name := range names {
		assert.True(t, len(name) > 3 && name[:3] == "SIG",
			"Signal name %q should start with SIG", name)
	}
}

// TestMatchSignalNames_GlobAlgorithm tests the glob matching algorithm.
func TestMatchSignalNames_GlobAlgorithm(t *testing.T) {
	catalog := NewCatalog()

	t.Run("empty pattern matches nothing", func(t *testing.T) {
		matches, err := catalog.MatchSignalNames("")
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("whitespace-only pattern matches nothing", func(t *testing.T) {
		matches, err := catalog.MatchSignalNames("   ")
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("pattern whitespace is trimmed", func(t *testing.T) {
		matches, err := catalog.MatchSignalNames("  SIGTERM  ")
		require.NoError(t, err)
		assert.Equal(t, []string{"SIGTERM"}, matches)
	})

	t.Run("literal pattern requires exact match", func(t *testing.T) {
		matches, err := catalog.MatchSignalNames("SIGTERM")
		require.NoError(t, err)
		assert.Equal(t, []string{"SIGTERM"}, matches)
	})

	t.Run("question mark matches single char", func(t *testing.T) {
		matches, err := catalog.MatchSignalNames("SIGUS?1")
		require.NoError(t, err)
		assert.Contains(t, matches, "SIGUSR1")
		assert.NotContains(t, matches, "SIGUSR2")
	})
}

// TestGlobMatch tests the internal glob matching function directly.
func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		// Basic matches
		{"", "", true},
		{"a", "a", true},
		{"a", "b", false},
		{"abc", "abc", true},
		{"abc", "abd", false},

		// Star wildcard
		{"*", "", true},
		{"*", "a", true},
		{"*", "abc", true},
		{"a*", "a", true},
		{"a*", "abc", true},
		{"a*", "b", false},
		{"*c", "c", true},
		{"*c", "abc", true},
		{"*c", "ab", false},
		{"a*c", "ac", true},
		{"a*c", "abc", true},
		{"a*c", "abbc", true},
		{"a*c", "ab", false},

		// Question mark wildcard
		{"?", "a", true},
		{"?", "", false},
		{"?", "ab", false},
		{"a?c", "abc", true},
		{"a?c", "ac", false},
		{"a?c", "abbc", false},
		{"???", "abc", true},
		{"???", "ab", false},
		{"???", "abcd", false},

		// Combined wildcards
		{"*?", "a", true},
		{"*?", "ab", true},
		{"*?", "", false},
		{"?*", "a", true},
		{"?*", "ab", true},
		{"?*", "", false},
		{"a*b?c", "abc", false},
		{"a*b?c", "abxc", true},
		{"a*b?c", "axxbxc", true},

		// Signal name examples
		{"sig*", "sigterm", true},
		{"sig*", "sigint", true},
		{"sig*", "term", false},
		{"*term", "sigterm", true},
		{"*term", "term", true},
		{"sig???", "sigint", true},
		{"sig???", "sighup", true},
		{"sig???", "sigterm", false},
		{"*usr*", "sigusr1", true},
		{"*usr*", "sigusr2", true},
		{"*usr*", "sigterm", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.text, func(t *testing.T) {
			got := globMatch(tt.pattern, tt.text)
			assert.Equal(t, tt.want, got, "globMatch(%q, %q)", tt.pattern, tt.text)
		})
	}
}
