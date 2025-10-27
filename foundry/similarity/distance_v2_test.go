package similarity

import (
	"testing"
)

// TestDistanceWithAlgorithm_Levenshtein verifies Levenshtein algorithm
func TestDistanceWithAlgorithm_Levenshtein(t *testing.T) {
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
				got, err := DistanceWithAlgorithm(tc.InputA, tc.InputB, AlgorithmLevenshtein)
				if err != nil {
					t.Fatalf("DistanceWithAlgorithm returned error: %v", err)
				}

				if got != tc.ExpectedDistance {
					t.Errorf("DistanceWithAlgorithm(%q, %q, Levenshtein) = %d, want %d",
						tc.InputA, tc.InputB, got, tc.ExpectedDistance)
				}
			})
		}
	}
}

// TestDistanceWithAlgorithm_DamerauOSA verifies Damerau OSA algorithm
func TestDistanceWithAlgorithm_DamerauOSA(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "damerau_osa" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				got, err := DistanceWithAlgorithm(tc.InputA, tc.InputB, AlgorithmDamerauOSA)
				if err != nil {
					t.Fatalf("DistanceWithAlgorithm returned error: %v", err)
				}

				if got != tc.ExpectedDistance {
					t.Errorf("DistanceWithAlgorithm(%q, %q, DamerauOSA) = %d, want %d",
						tc.InputA, tc.InputB, got, tc.ExpectedDistance)
				}
			})
		}
	}
}

// TestDistanceWithAlgorithm_DamerauUnrestricted verifies unrestricted Damerau algorithm
func TestDistanceWithAlgorithm_DamerauUnrestricted(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "damerau_unrestricted" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				got, err := DistanceWithAlgorithm(tc.InputA, tc.InputB, AlgorithmDamerauUnrestricted)
				if err != nil {
					t.Fatalf("DistanceWithAlgorithm returned error: %v", err)
				}

				if got != tc.ExpectedDistance {
					t.Errorf("DistanceWithAlgorithm(%q, %q, DamerauUnrestricted) = %d, want %d",
						tc.InputA, tc.InputB, got, tc.ExpectedDistance)
				}
			})
		}
	}
}

// TestDistanceWithAlgorithm_CriticalDistinction verifies OSA vs unrestricted distinction
func TestDistanceWithAlgorithm_CriticalDistinction(t *testing.T) {
	// Critical test case from fixtures: "CA" vs "ABC"
	// OSA: requires 3 operations (delete C, insert A, insert B)
	// Unrestricted: allows 2 operations (transpose + insert)

	osaDistance, err := DistanceWithAlgorithm("CA", "ABC", AlgorithmDamerauOSA)
	if err != nil {
		t.Fatalf("OSA returned error: %v", err)
	}

	unrestrictedDistance, err := DistanceWithAlgorithm("CA", "ABC", AlgorithmDamerauUnrestricted)
	if err != nil {
		t.Fatalf("Unrestricted returned error: %v", err)
	}

	if osaDistance != 3 {
		t.Errorf("OSA distance for 'CA' vs 'ABC' = %d, want 3", osaDistance)
	}

	if unrestrictedDistance != 2 {
		t.Errorf("Unrestricted distance for 'CA' vs 'ABC' = %d, want 2", unrestrictedDistance)
	}

	if osaDistance == unrestrictedDistance {
		t.Errorf("OSA and unrestricted should differ for 'CA' vs 'ABC', both returned %d", osaDistance)
	}

	t.Logf("✓ Critical distinction verified: OSA=%d, Unrestricted=%d", osaDistance, unrestrictedDistance)
}

// TestScoreWithAlgorithm_Levenshtein verifies Levenshtein score calculation
func TestScoreWithAlgorithm_Levenshtein(t *testing.T) {
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
				got, err := ScoreWithAlgorithm(tc.InputA, tc.InputB, AlgorithmLevenshtein, nil)
				if err != nil {
					t.Fatalf("ScoreWithAlgorithm returned error: %v", err)
				}

				if !floatNearlyEqual(got, tc.ExpectedScore, 0.0001) {
					t.Errorf("ScoreWithAlgorithm(%q, %q, Levenshtein) = %.16f, want %.16f (diff: %.16f)",
						tc.InputA, tc.InputB, got, tc.ExpectedScore, got-tc.ExpectedScore)
				}
			})
		}
	}
}

// TestScoreWithAlgorithm_DamerauOSA verifies Damerau OSA score calculation
func TestScoreWithAlgorithm_DamerauOSA(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "damerau_osa" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				got, err := ScoreWithAlgorithm(tc.InputA, tc.InputB, AlgorithmDamerauOSA, nil)
				if err != nil {
					t.Fatalf("ScoreWithAlgorithm returned error: %v", err)
				}

				if !floatNearlyEqual(got, tc.ExpectedScore, 0.0001) {
					t.Errorf("ScoreWithAlgorithm(%q, %q, DamerauOSA) = %.16f, want %.16f (diff: %.16f)",
						tc.InputA, tc.InputB, got, tc.ExpectedScore, got-tc.ExpectedScore)
				}
			})
		}
	}
}

// TestScoreWithAlgorithm_DamerauUnrestricted verifies unrestricted Damerau score calculation
func TestScoreWithAlgorithm_DamerauUnrestricted(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "damerau_unrestricted" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				got, err := ScoreWithAlgorithm(tc.InputA, tc.InputB, AlgorithmDamerauUnrestricted, nil)
				if err != nil {
					t.Fatalf("ScoreWithAlgorithm returned error: %v", err)
				}

				if !floatNearlyEqual(got, tc.ExpectedScore, 0.0001) {
					t.Errorf("ScoreWithAlgorithm(%q, %q, DamerauUnrestricted) = %.16f, want %.16f (diff: %.16f)",
						tc.InputA, tc.InputB, got, tc.ExpectedScore, got-tc.ExpectedScore)
				}
			})
		}
	}
}

// TestScoreWithAlgorithm_JaroWinkler verifies Jaro-Winkler score calculation
func TestScoreWithAlgorithm_JaroWinkler(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, group := range fixtures.TestCases {
		if group.Category != "jaro_winkler" {
			continue
		}

		for _, tc := range group.Cases {
			name := tc.Description
			if name == "" {
				name = tc.InputA + "_" + tc.InputB
			}

			t.Run(name, func(t *testing.T) {
				opts := DefaultScoreOptions()
				if tc.PrefixScale > 0 {
					opts.JaroPrefixScale = tc.PrefixScale
				}
				if tc.MaxPrefix > 0 {
					opts.JaroMaxPrefix = tc.MaxPrefix
				}

				got, err := ScoreWithAlgorithm(tc.InputA, tc.InputB, AlgorithmJaroWinkler, opts)
				if err != nil {
					t.Fatalf("ScoreWithAlgorithm returned error: %v", err)
				}

				if !floatNearlyEqual(got, tc.ExpectedScore, 0.0001) {
					t.Errorf("ScoreWithAlgorithm(%q, %q, JaroWinkler) = %.16f, want %.16f (diff: %.16f)",
						tc.InputA, tc.InputB, got, tc.ExpectedScore, got-tc.ExpectedScore)
				}
			})
		}
	}
}

// TestDistanceWithAlgorithm_ErrorCases verifies error handling
func TestDistanceWithAlgorithm_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		algorithm Algorithm
		wantError string
	}{
		{
			name:      "jaro_winkler_distance",
			algorithm: AlgorithmJaroWinkler,
			wantError: "jaro_winkler metric produces similarity scores",
		},
		{
			name:      "substring_distance",
			algorithm: AlgorithmSubstring,
			wantError: "substring metric does not produce distances",
		},
		{
			name:      "invalid_algorithm",
			algorithm: "invalid",
			wantError: "invalid algorithm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DistanceWithAlgorithm("test", "test", tt.algorithm)
			if err == nil {
				t.Errorf("DistanceWithAlgorithm(%q) expected error, got nil", tt.algorithm)
				return
			}

			if err.Error() == "" || len(err.Error()) < len(tt.wantError) {
				t.Errorf("DistanceWithAlgorithm(%q) error too short: %v", tt.algorithm, err)
			}
		})
	}
}

// TestSubstringMatch verifies substring matching
func TestSubstringMatch(t *testing.T) {
	tests := []struct {
		name          string
		needle        string
		haystack      string
		expectedStart int
		expectedEnd   int
		expectedValid bool
		expectedScore float64
	}{
		{
			name:          "prefix_match",
			needle:        "hello",
			haystack:      "hello world",
			expectedStart: 0,
			expectedEnd:   5,
			expectedValid: true,
			expectedScore: 0.4545454545454545,
		},
		{
			name:          "suffix_match",
			needle:        "world",
			haystack:      "hello world",
			expectedStart: 6,
			expectedEnd:   11,
			expectedValid: true,
			expectedScore: 0.4545454545454545,
		},
		{
			name:          "no_match",
			needle:        "xyz",
			haystack:      "abcdef",
			expectedValid: false,
			expectedScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, score := SubstringMatch(tt.needle, tt.haystack)

			if match.Valid != tt.expectedValid {
				t.Errorf("SubstringMatch(%q, %q).Valid = %v, want %v",
					tt.needle, tt.haystack, match.Valid, tt.expectedValid)
			}

			if tt.expectedValid {
				if match.Start != tt.expectedStart {
					t.Errorf("SubstringMatch(%q, %q).Start = %d, want %d",
						tt.needle, tt.haystack, match.Start, tt.expectedStart)
				}
				if match.End != tt.expectedEnd {
					t.Errorf("SubstringMatch(%q, %q).End = %d, want %d",
						tt.needle, tt.haystack, match.End, tt.expectedEnd)
				}
			}

			if !floatNearlyEqual(score, tt.expectedScore, 0.0001) {
				t.Errorf("SubstringMatch(%q, %q) score = %.16f, want %.16f",
					tt.needle, tt.haystack, score, tt.expectedScore)
			}
		})
	}
}
