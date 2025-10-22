package similarity

import (
	"testing"
)

// TestDefaultSuggestOptions tests the default options constructor
func TestDefaultSuggestOptions(t *testing.T) {
	opts := DefaultSuggestOptions()

	if opts.MinScore != 0.6 {
		t.Errorf("DefaultSuggestOptions().MinScore = %f, want 0.6", opts.MinScore)
	}
	if opts.MaxSuggestions != 3 {
		t.Errorf("DefaultSuggestOptions().MaxSuggestions = %d, want 3", opts.MaxSuggestions)
	}
	if !opts.Normalize {
		t.Errorf("DefaultSuggestOptions().Normalize = %v, want true", opts.Normalize)
	}
}

// TestSuggest_Basic tests basic suggestion functionality
func TestSuggest_Basic(t *testing.T) {
	candidates := []string{"docscribe", "crucible", "foundry", "similarity"}

	suggestions := Suggest("docscrib", candidates, DefaultSuggestOptions())

	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion, got none")
	}

	// First suggestion should be "docscribe" with high score
	if suggestions[0].Value != "docscribe" {
		t.Errorf("Top suggestion = %q, want %q", suggestions[0].Value, "docscribe")
	}

	// Score should be high (1 character difference out of 9)
	expectedScore := 1.0 - 1.0/9.0 // 0.888...
	if !floatNearlyEqual(suggestions[0].Score, expectedScore, 0.01) {
		t.Errorf("Top suggestion score = %f, want ~%f", suggestions[0].Score, expectedScore)
	}
}

// TestSuggest_EmptyCandidates tests with no candidates
func TestSuggest_EmptyCandidates(t *testing.T) {
	suggestions := Suggest("test", []string{}, DefaultSuggestOptions())

	if len(suggestions) != 0 {
		t.Errorf("Suggest with empty candidates = %d results, want 0", len(suggestions))
	}
}

// TestSuggest_NoMatches tests when no candidates meet threshold
func TestSuggest_NoMatches(t *testing.T) {
	candidates := []string{"abc", "def", "ghi"}
	opts := DefaultSuggestOptions()

	suggestions := Suggest("xyz", candidates, opts)

	if len(suggestions) != 0 {
		t.Errorf("Suggest with no matches = %d results, want 0", len(suggestions))
	}
}

// TestSuggest_MaxSuggestions tests the max suggestions limit
func TestSuggest_MaxSuggestions(t *testing.T) {
	candidates := []string{"test1", "test2", "test3", "test4", "test5"}
	opts := SuggestOptions{
		MinScore:       0.6,
		MaxSuggestions: 2,
		Normalize:      true,
	}

	suggestions := Suggest("test", candidates, opts)

	if len(suggestions) != 2 {
		t.Errorf("Suggest with MaxSuggestions=2 returned %d results, want 2", len(suggestions))
	}

	// Should return test1 and test2 (alphabetically first in tie)
	if suggestions[0].Value != "test1" || suggestions[1].Value != "test2" {
		t.Errorf("Top 2 suggestions = [%q, %q], want [%q, %q]",
			suggestions[0].Value, suggestions[1].Value, "test1", "test2")
	}
}

// TestSuggest_Threshold tests the minimum score threshold
func TestSuggest_Threshold(t *testing.T) {
	candidates := []string{"hello", "help", "world"}
	opts := SuggestOptions{
		MinScore:       0.8, // High threshold
		MaxSuggestions: 3,
		Normalize:      true,
	}

	suggestions := Suggest("hell", candidates, opts)

	// Only "hello" (score 0.8) should meet the threshold
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion")
	}

	// All returned suggestions should have score >= 0.8
	for _, s := range suggestions {
		if s.Score < 0.8 {
			t.Errorf("Suggestion %q has score %f, want >= 0.8", s.Value, s.Score)
		}
	}
}

// TestSuggest_TieBreaking tests alphabetical tie-breaking
func TestSuggest_TieBreaking(t *testing.T) {
	// All candidates have same edit distance from "test"
	candidates := []string{"test3", "test1", "test2"}
	opts := DefaultSuggestOptions()

	suggestions := Suggest("test", candidates, opts)

	// Should be sorted by score first, then alphabetically
	if len(suggestions) < 3 {
		t.Fatalf("Expected 3 suggestions, got %d", len(suggestions))
	}

	// All have same score, should be alphabetically sorted
	if suggestions[0].Value != "test1" {
		t.Errorf("First suggestion = %q, want %q (alphabetical order)", suggestions[0].Value, "test1")
	}
	if suggestions[1].Value != "test2" {
		t.Errorf("Second suggestion = %q, want %q (alphabetical order)", suggestions[1].Value, "test2")
	}
	if suggestions[2].Value != "test3" {
		t.Errorf("Third suggestion = %q, want %q (alphabetical order)", suggestions[2].Value, "test3")
	}
}

// TestSuggest_CaseInsensitive tests case-insensitive matching
func TestSuggest_CaseInsensitive(t *testing.T) {
	candidates := []string{"DocScribe", "Crucible", "Foundry"}
	opts := SuggestOptions{
		MinScore:       0.9,
		MaxSuggestions: 3,
		Normalize:      true, // Case-insensitive
	}

	suggestions := Suggest("DOCSCRIBE", candidates, opts)

	if len(suggestions) == 0 {
		t.Fatal("Expected suggestions for case-insensitive match")
	}

	// Should match "DocScribe" despite case difference
	if suggestions[0].Value != "DocScribe" {
		t.Errorf("Top suggestion = %q, want %q", suggestions[0].Value, "DocScribe")
	}

	// Score should be 1.0 (exact match when normalized)
	if !floatNearlyEqual(suggestions[0].Score, 1.0, 0.01) {
		t.Errorf("Score for exact case-insensitive match = %f, want 1.0", suggestions[0].Score)
	}
}

// TestSuggest_CaseSensitive tests case-sensitive matching
func TestSuggest_CaseSensitive(t *testing.T) {
	candidates := []string{"DocScribe", "Crucible", "Foundry"}
	opts := SuggestOptions{
		MinScore:       0.9,
		MaxSuggestions: 3,
		Normalize:      false, // Case-sensitive
	}

	suggestions := Suggest("DOCSCRIBE", candidates, opts)

	// With case-sensitive matching, "DOCSCRIBE" vs "DocScribe" has distance 9
	// Score would be much lower than 0.9
	if len(suggestions) > 0 {
		t.Errorf("Expected no suggestions for case-sensitive mismatch, got %d", len(suggestions))
	}
}

// TestSuggest_ScoreOrdering tests that results are ordered by score
func TestSuggest_ScoreOrdering(t *testing.T) {
	candidates := []string{"format", "validate", "build", "test"}
	opts := DefaultSuggestOptions()

	suggestions := Suggest("formatt", candidates, opts)

	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion")
	}

	// "format" should be first (highest score)
	if suggestions[0].Value != "format" {
		t.Errorf("Top suggestion = %q, want %q", suggestions[0].Value, "format")
	}

	// Verify scores are in descending order
	for i := 1; i < len(suggestions); i++ {
		if suggestions[i].Score > suggestions[i-1].Score {
			t.Errorf("Suggestions not sorted by score: position %d has higher score than %d", i, i-1)
		}
	}
}

// TestSuggest_DefaultBehavior tests default option handling
func TestSuggest_DefaultBehavior(t *testing.T) {
	candidates := []string{"hello", "help", "world"}

	// Empty options - should use defaults
	opts := SuggestOptions{}

	suggestions := Suggest("helo", candidates, opts)

	// With MinScore=0.6 default, should get matches
	if len(suggestions) == 0 {
		t.Error("Expected suggestions with default options")
	}

	// Should return at most 3 (default MaxSuggestions)
	if len(suggestions) > 3 {
		t.Errorf("Returned %d suggestions, want max 3 (default)", len(suggestions))
	}
}

// TestSuggest_ExactMatch tests behavior with exact matches
func TestSuggest_ExactMatch(t *testing.T) {
	candidates := []string{"exact", "similar", "different"}
	opts := DefaultSuggestOptions()

	suggestions := Suggest("exact", candidates, opts)

	if len(suggestions) == 0 {
		t.Fatal("Expected suggestions including exact match")
	}

	// Exact match should be first with score 1.0
	if suggestions[0].Value != "exact" {
		t.Errorf("Top suggestion = %q, want %q", suggestions[0].Value, "exact")
	}

	if !floatNearlyEqual(suggestions[0].Score, 1.0, 0.001) {
		t.Errorf("Exact match score = %f, want 1.0", suggestions[0].Score)
	}
}

// TestSuggest_LongCandidates tests with longer candidate strings
// Tests realistic scenario: partial input matching longer similar terms
func TestSuggest_LongCandidates(t *testing.T) {
	candidates := []string{
		"schema_definition",
		"schema_validation",
		"database_schema",
		"schema_migration",
		"configuration",
		"schedule_task",
	}
	opts := SuggestOptions{
		MinScore:       0.3, // Realistic threshold for partial input vs longer candidates
		MaxSuggestions: 3,
		Normalize:      true,
	}

	suggestions := Suggest("schem", candidates, opts)

	// Should get suggestions for partial matches with "schema" prefix
	if len(suggestions) == 0 {
		t.Error("Expected suggestions for candidates with 'schema' prefix")
	}

	// Verify all returned suggestions meet the threshold
	for _, s := range suggestions {
		if s.Score < 0.3 {
			t.Errorf("Suggestion %q has score %f, below threshold 0.3", s.Value, s.Score)
		}
	}

	// Verify scores are in descending order
	for i := 1; i < len(suggestions); i++ {
		if suggestions[i].Score > suggestions[i-1].Score {
			t.Errorf("Suggestions not sorted: position %d score %f > position %d score %f",
				i, suggestions[i].Score, i-1, suggestions[i-1].Score)
		}
	}

	// Verify at least one suggestion starts with "schema"
	foundSchemaMatch := false
	for _, s := range suggestions {
		if len(s.Value) >= 6 && s.Value[:6] == "schema" {
			foundSchemaMatch = true
			break
		}
	}
	if !foundSchemaMatch {
		t.Error("Expected at least one suggestion starting with 'schema'")
	}
}
