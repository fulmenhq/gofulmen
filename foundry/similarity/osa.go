package similarity

// osaDistance implements Optimal String Alignment distance.
//
// OSA is a variant of Damerau-Levenshtein distance with a restriction:
// cannot edit the same substring more than once. This restriction makes
// OSA suitable for detecting common typing errors and simple typos.
//
// Allowed operations:
// - Insertions: insert a character
// - Deletions: delete a character
// - Substitutions: replace a character
// - Adjacent Transpositions: swap two adjacent characters
//
// Based on rapidfuzz-cpp OSA implementation:
// https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/OSA.hpp
//
// Copyright notice from rapidfuzz-cpp:
// SPDX-License-Identifier: MIT
// Copyright © 2022-present Max Bachmann
//
// Ported to Go with Unicode (rune) handling and optimizations from
// gofulmen Levenshtein implementation (see similarity.go).
//
// Time complexity: O(m×n) where m, n are string lengths
// Space complexity: O(min(m,n)) using three-row optimization
//
// Examples:
//
//	osaDistance("abcd", "abdc") returns 1 (transposition of cd->dc)
//	osaDistance("hello", "ehllo") returns 1 (transposition of he->eh)
//	osaDistance("CA", "ABC") returns 3 (OSA restriction applies)
//
// The critical distinction from unrestricted Damerau-Levenshtein:
// OSA enforces "edit once" restriction, which prevents certain
// transformation sequences that would be optimal in unrestricted variant.
func osaDistance(a, b string) int {
	// Convert to runes for Unicode support
	runesA := []rune(a)
	runesB := []rune(b)

	lenA := len(runesA)
	lenB := len(runesB)

	// Edge cases: empty strings
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Ensure we iterate over the shorter string for better cache locality
	// If B is shorter, swap so A is always the shorter one
	if lenB < lenA {
		runesA, runesB = runesB, runesA
		lenA, lenB = lenB, lenA
	}

	// OSA requires THREE rows for dynamic programming:
	// - prevPrevRow: stores values from two iterations ago (for transposition detection)
	// - prevRow: stores values from previous iteration
	// - currRow: stores current iteration values
	//
	// This is one more row than Levenshtein (which uses two rows).
	prevPrevRow := make([]int, lenA+1)
	prevRow := make([]int, lenA+1)
	currRow := make([]int, lenA+1)

	// Initialize first row: distance from empty string to prefixes of A
	for i := 0; i <= lenA; i++ {
		prevRow[i] = i
	}

	// Fill the matrix row by row
	for j := 1; j <= lenB; j++ {
		// First column: distance from empty string to prefix of B
		currRow[0] = j

		for i := 1; i <= lenA; i++ {
			// Cost of substitution: 0 if characters match, 1 otherwise
			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
			}

			// Calculate minimum of three standard operations (same as Levenshtein):
			// 1. Deletion: currRow[i-1] + 1
			// 2. Insertion: prevRow[i] + 1
			// 3. Substitution: prevRow[i-1] + cost
			deletion := currRow[i-1] + 1
			insertion := prevRow[i] + 1
			substitution := prevRow[i-1] + cost

			// Take minimum of standard operations
			currRow[i] = deletion
			if insertion < currRow[i] {
				currRow[i] = insertion
			}
			if substitution < currRow[i] {
				currRow[i] = substitution
			}

			// OSA TRANSPOSITION CHECK (key difference from Levenshtein):
			// Check if we can perform an adjacent transposition.
			// This requires:
			// - We're at least at position (2,2) in the matrix (i > 1 && j > 1)
			// - Current character in A matches character before current in B (runesA[i-1] == runesB[j-2])
			// - Character before current in A matches current character in B (runesA[i-2] == runesB[j-1])
			//
			// Example: transforming "ab" to "ba"
			// - Position (2,2): A[1]='b' matches B[0]='b', A[0]='a' matches B[1]='a'
			// - This is an adjacent transposition, cost is prevPrevRow[i-2] + 1
			if i > 1 && j > 1 &&
				runesA[i-1] == runesB[j-2] &&
				runesA[i-2] == runesB[j-1] {
				// Adjacent transposition detected
				transpose := prevPrevRow[i-2] + 1
				if transpose < currRow[i] {
					currRow[i] = transpose
				}
			}
		}

		// Rotate rows for next iteration:
		// - Current becomes previous
		// - Previous becomes prev-previous
		// - Prev-previous becomes current (will be overwritten)
		prevPrevRow, prevRow, currRow = prevRow, currRow, prevPrevRow
	}

	// Result is in the last column of the last processed row (now prevRow)
	return prevRow[lenA]
}
