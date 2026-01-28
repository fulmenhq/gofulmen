package fulencode

import (
	"unicode/utf8"
)

func Detect(data []byte, options *DetectOptions) (DetectionResult, error) {
	opts := applyDetectDefaults(options)

	sample := data
	if opts.MaxSampleSize > 0 && len(sample) > opts.MaxSampleSize {
		sample = sample[:opts.MaxSampleSize]
	}

	// 1) BOM
	bom := DetectBOM(sample)
	if bom.BOMType != nil && bom.EncodingImplied != nil {
		enc := *bom.EncodingImplied
		return DetectionResult{
			Encoding:   &enc,
			Confidence: 1.0,
			Level:      ConfidenceHigh,
			Warnings:   []string{},
		}, nil
	}

	// 2) Multibase prefix (minimal)
	var multibase *string
	if opts.RecognizeMultibase != nil && *opts.RecognizeMultibase && len(sample) > 0 {
		// Minimal: 'f' indicates multibase base16 (hex)
		if sample[0] == 'f' {
			s := "f"
			multibase = &s
		}
	}

	// 3) UTF-16 null pattern
	if len(sample) >= 4 {
		evenNulls := 0
		oddNulls := 0
		for i := 0; i < len(sample); i++ {
			if sample[i] != 0 {
				continue
			}
			if i%2 == 0 {
				evenNulls++
			} else {
				oddNulls++
			}
		}
		// Heuristic: strong bias to one parity.
		if oddNulls > 0 && oddNulls >= evenNulls*4 && oddNulls >= len(sample)/8 {
			enc := string(UTF16LE)
			res := DetectionResult{Encoding: &enc, Confidence: 0.95, Level: ConfidenceHigh, MultibasePrefix: multibase, Warnings: []string{}}
			if res.Confidence < opts.MinConfidence {
				res.Encoding = nil
				res.Level = ConfidenceLow
				res.Warnings = append(res.Warnings, "below_min_confidence")
			}
			return res, nil
		}
		if evenNulls > 0 && evenNulls >= oddNulls*4 && evenNulls >= len(sample)/8 {
			enc := string(UTF16BE)
			res := DetectionResult{Encoding: &enc, Confidence: 0.95, Level: ConfidenceHigh, MultibasePrefix: multibase, Warnings: []string{}}
			if res.Confidence < opts.MinConfidence {
				res.Encoding = nil
				res.Level = ConfidenceLow
				res.Warnings = append(res.Warnings, "below_min_confidence")
			}
			return res, nil
		}
	}

	// 4) UTF-8 validation
	if utf8.Valid(sample) {
		allASCII := true
		for _, b := range sample {
			if b >= 0x80 {
				allASCII = false
				break
			}
		}
		enc := string(UTF8)
		confidence := 0.90
		level := ConfidenceHigh
		if allASCII {
			confidence = 0.20
			level = ConfidenceLow
		}
		res := DetectionResult{Encoding: &enc, Confidence: confidence, Level: level, MultibasePrefix: multibase, Warnings: []string{}}
		if res.Confidence < opts.MinConfidence {
			res.Encoding = nil
			res.Level = ConfidenceLow
			res.Warnings = append(res.Warnings, "below_min_confidence")
		}
		return res, nil
	}

	// Unknown.
	res := DetectionResult{Encoding: nil, Confidence: 0.0, Level: ConfidenceLow, MultibasePrefix: multibase, Warnings: []string{"unrecognized_encoding"}}
	if res.Confidence < opts.MinConfidence {
		res.Warnings = append(res.Warnings, "below_min_confidence")
	}
	return res, nil
}
