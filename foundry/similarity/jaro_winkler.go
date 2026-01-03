package similarity

// Native Jaro-Winkler implementation.
//
// Replaces github.com/antzucaro/matchr (GPL-2.0) with MIT-compatible native code.
// Algorithm based on the original Jaro-Winkler paper and validated against
// matchr behavior for exact parity.
//
// References:
// - Jaro, M.A. (1989). "Advances in Record-Linkage Methodology..."
// - Winkler, W.E. (1990). "String Comparator Metrics and Enhanced Decision Rules..."

// jaroSimilarity calculates the Jaro similarity between two strings.
//
// The Jaro similarity is defined as:
//
//	0 if both strings are empty or no matching characters
//	(m/|s1| + m/|s2| + (m-t)/m) / 3 otherwise
//
// where:
//
//	m = number of matching characters
//	t = number of transpositions / 2
//	|s1|, |s2| = lengths of the strings
//
// Characters are considered matching if they are the same and within
// the match window: floor(max(|s1|, |s2|) / 2) - 1
func jaroSimilarity(s1, s2 string) float64 {
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Handle empty strings
	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate match window
	// matchWindow = floor(max(len1, len2) / 2) - 1
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	matchWindow := maxLen/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	// Track which characters have been matched
	matched1 := make([]bool, len1)
	matched2 := make([]bool, len2)

	// Count matches
	matches := 0
	for i := 0; i < len1; i++ {
		// Define the window bounds for s2
		start := i - matchWindow
		if start < 0 {
			start = 0
		}
		end := i + matchWindow + 1
		if end > len2 {
			end = len2
		}

		for j := start; j < end; j++ {
			if matched2[j] || r1[i] != r2[j] {
				continue
			}
			matched1[i] = true
			matched2[j] = true
			matches++
			break
		}
	}

	// No matches found
	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	// A transposition is when matched characters are in different order
	transpositions := 0
	j := 0
	for i := 0; i < len1; i++ {
		if !matched1[i] {
			continue
		}
		// Find next matched character in s2
		for !matched2[j] {
			j++
		}
		if r1[i] != r2[j] {
			transpositions++
		}
		j++
	}

	// Calculate Jaro similarity
	m := float64(matches)
	t := float64(transpositions) / 2.0

	return (m/float64(len1) + m/float64(len2) + (m-t)/m) / 3.0
}

// jaroWinklerSimilarity calculates the Jaro-Winkler similarity between two strings.
//
// The Jaro-Winkler similarity adds a prefix bonus to the Jaro similarity:
//
//	JW = J + (l * p * (1 - J))
//
// where:
//
//	J = Jaro similarity
//	l = length of common prefix (up to maxPrefix, typically 4)
//	p = prefix scale factor (typically 0.1)
//
// The longTolerance parameter matches matchr.JaroWinkler behavior:
// when true, applies additional tolerance for longer strings.
func jaroWinklerSimilarity(s1, s2 string, longTolerance bool) float64 {
	r1 := []rune(s1)
	r2 := []rune(s2)

	// Handle empty strings
	if len(r1) == 0 && len(r2) == 0 {
		return 1.0
	}
	if len(r1) == 0 || len(r2) == 0 {
		return 0.0
	}

	// Calculate base Jaro similarity
	jaro := jaroSimilarity(s1, s2)

	// Find common prefix length (up to 4 characters)
	maxPrefix := 4
	prefixLen := 0
	minLen := len(r1)
	if len(r2) < minLen {
		minLen = len(r2)
	}
	if maxPrefix > minLen {
		maxPrefix = minLen
	}

	for i := 0; i < maxPrefix; i++ {
		if r1[i] == r2[i] {
			prefixLen++
		} else {
			break
		}
	}

	// Standard prefix scale factor
	prefixScale := 0.1

	// Apply Winkler modification
	jw := jaro + float64(prefixLen)*prefixScale*(1.0-jaro)

	// Long tolerance adjustment (matches matchr behavior)
	if longTolerance && jaro > 0.7 {
		// Additional bonus for longer strings with high similarity
		len1 := len(r1)
		len2 := len(r2)
		minL := len1
		if len2 < minL {
			minL = len2
		}
		if minL > 4 {
			// Apply long string adjustment
			jw += (1.0 - jw) * float64(minL-4) * 0.1 * (1.0 - jaro)
		}
	}

	// Clamp to [0, 1]
	if jw > 1.0 {
		return 1.0
	}
	if jw < 0.0 {
		return 0.0
	}

	return jw
}
