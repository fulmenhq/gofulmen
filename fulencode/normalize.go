package fulencode

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func Normalize(text string, profile NormalizationProfile, options *NormalizeOptions) (NormalizationResult, error) {
	if err := ValidateNormalizationProfile(profile); err != nil {
		return NormalizationResult{}, newError(OperationNormalize, "INVALID_PROFILE", err.Error())
	}
	opts := applyNormalizeDefaults(options)

	inputRunes := utf8.RuneCountInString(text)
	var out string
	transformations := make([]string, 0)
	warnings := make([]string, 0)

	switch profile {
	case NFC:
		out = norm.NFC.String(text)
		transformations = append(transformations, "nfc")
	case NFD:
		out = norm.NFD.String(text)
		transformations = append(transformations, "nfd")
	case NFKC:
		out = norm.NFKC.String(text)
		transformations = append(transformations, "nfkc")
		if opts.WarnSemanticChange != nil && *opts.WarnSemanticChange {
			warnings = append(warnings, "semantic_change_possible")
		}
	case NFKD:
		out = norm.NFKD.String(text)
		transformations = append(transformations, "nfkd")
		if opts.WarnSemanticChange != nil && *opts.WarnSemanticChange {
			warnings = append(warnings, "semantic_change_possible")
		}
	case TEXT_SAFE:
		res, err := normalizeTextSafe(text, opts)
		if err != nil {
			return NormalizationResult{}, err
		}
		out = res
		transformations = append(transformations, "nfc", "text_safe")
	default:
		return NormalizationResult{}, newError(OperationNormalize, "UNSUPPORTED_PROFILE", "normalization profile not implemented")
	}

	return NormalizationResult{
		Text:                   out,
		Profile:                string(profile),
		InputLength:            inputRunes,
		OutputLength:           utf8.RuneCountInString(out),
		TransformationsApplied: transformations,
		SemanticChanges:        []SemanticChange{},
		Warnings:               warnings,
	}, nil
}

func normalizeTextSafe(text string, opts *NormalizeOptions) (string, error) {
	// Step 1: NFC
	normalized := norm.NFC.String(text)

	// Step 2-3: reject forbidden characters
	for idx, r := range normalized {
		// Control characters (general category Cc)
		if unicode.IsControl(r) {
			return "", addDetails(newError(OperationNormalize, "INVALID_ENCODING", "control character not allowed"), map[string]any{"codepoint_offset": idx})
		}
		if opts.RejectBidiControls != nil && *opts.RejectBidiControls {
			if isBidiControl(r) {
				return "", addDetails(newError(OperationNormalize, "BIDI_CONTROL_CHARACTER", "bidi control character not allowed"), map[string]any{"codepoint_offset": idx})
			}
		}
		if opts.RejectZeroWidth != nil && *opts.RejectZeroWidth {
			if isZeroWidth(r) {
				return "", addDetails(newError(OperationNormalize, "ZERO_WIDTH_CHARACTER", "zero-width character not allowed"), map[string]any{"codepoint_offset": idx})
			}
		}
	}

	// Step 4: combining mark cap
	maxMarks := opts.MaxCombiningMarks
	if maxMarks <= 0 {
		maxMarks = DefaultMaxCombiningMarks
	}
	marks := 0
	for idx, r := range normalized {
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r) {
			marks++
			if marks > maxMarks {
				return "", addDetails(newError(OperationNormalize, "EXCESSIVE_COMBINING_MARKS", "too many combining marks"), map[string]any{"codepoint_offset": idx, "max_combining_marks": maxMarks})
			}
			continue
		}
		marks = 0
	}

	return normalized, nil
}

func isBidiControl(r rune) bool {
	if r >= 0x202A && r <= 0x202E {
		return true
	}
	if r >= 0x2066 && r <= 0x2069 {
		return true
	}
	switch r {
	case 0x200E, 0x200F, 0x061C:
		return true
	default:
		return false
	}
}

func isZeroWidth(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0xFEFF:
		return true
	default:
		return false
	}
}
