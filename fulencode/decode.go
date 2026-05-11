package fulencode

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func DecodeString(input string, format EncodingFormat, options *DecodeOptions) (DecodingResult, error) {
	return decodeInternal([]byte(input), format, options)
}

func DecodeBytes(input []byte, format EncodingFormat, options *DecodeOptions) (DecodingResult, error) {
	return decodeInternal(input, format, options)
}

func decodeInternal(input []byte, format EncodingFormat, options *DecodeOptions) (DecodingResult, error) {
	if err := ValidateEncodingFormat(format); err != nil {
		return DecodingResult{}, withInputFormat(newError(OperationDecode, "INVALID_FORMAT", err.Error()), format)
	}
	opts := applyDecodeDefaults(format, options)

	if opts.OnError == DecodeFallback {
		formats := append([]EncodingFormat{format}, opts.FallbackFormats...)
		var lastErr error
		for _, f := range formats {
			res, err := decodeInternal(input, f, &DecodeOptions{
				VerifyChecksum:    opts.VerifyChecksum,
				ComputeChecksum:   opts.ComputeChecksum,
				MaxDecodedSize:    opts.MaxDecodedSize,
				MaxExpansionRatio: opts.MaxExpansionRatio,
				OnError:           DecodeStrict,
				IgnoreWhitespace:  opts.IgnoreWhitespace,
				ValidatePadding:   opts.ValidatePadding,
			})
			if err == nil {
				res.Warnings = append(res.Warnings, "used_fallback_format")
				return res, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return DecodingResult{}, lastErr
		}
	}

	if opts.OnError != DecodeStrict {
		// Non-strict modes are only defined for character encodings (string->bytes)
		// and format fallback. For binary-to-text decoders, invalid input remains an error.
		switch format {
		case UTF8, UTF16LE, UTF16BE, ISO_8859_1, CP1252, ASCII:
			// ok
		default:
			return DecodingResult{}, withInputFormat(newError(OperationDecode, "INVALID_OPTIONS", "on_error is only supported for character encodings or fallback"), format)
		}
	}

	warnings := make([]string, 0)
	inputStr := string(input)
	normalized := inputStr
	if opts.IgnoreWhitespace != nil && *opts.IgnoreWhitespace {
		normalized = stripASCIIWhitespace(normalized)
	}

	maxDecoded := opts.MaxDecodedSize
	if maxDecoded < 1024 {
		maxDecoded = 1024
	}
	maxRatio := opts.MaxExpansionRatio
	if maxRatio < 1.0 {
		maxRatio = 1.0
	}

	var out []byte
	switch format {
	case BASE64:
		b, err := decodeBase64(normalized, false, opts.ValidatePadding != nil && *opts.ValidatePadding)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case BASE64URL:
		b, err := decodeBase64(normalized, true, opts.ValidatePadding != nil && *opts.ValidatePadding)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case BASE64_RAW:
		s := strings.TrimRight(normalized, "=")
		// Add padding to decode with standard encoding.
		pad := (4 - (len(s) % 4)) % 4
		s = s + strings.Repeat("=", pad)
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return DecodingResult{}, withInputFormat(newError(OperationDecode, "INVALID_ENCODING", err.Error()), format)
		}
		out = b
	case HEX:
		if len(normalized)%2 != 0 {
			return DecodingResult{}, withInputFormat(newError(OperationDecode, "INVALID_ENCODING", "hex input length must be even"), format)
		}
		b, err := hex.DecodeString(normalized)
		if err != nil {
			return DecodingResult{}, withInputFormat(newError(OperationDecode, "INVALID_ENCODING", err.Error()), format)
		}
		out = b
	case BASE32:
		b, err := decodeBase32(normalized, false, opts.ValidatePadding != nil && *opts.ValidatePadding)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case BASE32HEX:
		b, err := decodeBase32(normalized, true, opts.ValidatePadding != nil && *opts.ValidatePadding)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case UTF8:
		out = []byte(inputStr)
	case UTF16LE:
		b, _, err := encodeUTF16String(inputStr, true)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case UTF16BE:
		b, _, err := encodeUTF16String(inputStr, false)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
	case ISO_8859_1:
		b, corrections, err := encodeSingleByte(inputStr, charmap.ISO8859_1, opts.OnError)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
		if corrections > 0 {
			warnings = append(warnings, "corrections_applied")
		}
	case CP1252:
		b, corrections, err := encodeSingleByte(inputStr, charmap.Windows1252, opts.OnError)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
		if corrections > 0 {
			warnings = append(warnings, "corrections_applied")
		}
	case ASCII:
		b, corrections, err := encodeASCIIString(inputStr, opts.OnError)
		if err != nil {
			return DecodingResult{}, withInputFormat(err, format)
		}
		out = b
		if corrections > 0 {
			warnings = append(warnings, "corrections_applied")
		}
	default:
		return DecodingResult{}, withInputFormat(newError(OperationDecode, "UNSUPPORTED_FORMAT", fmt.Sprintf("unsupported decoding format: %s", format)), format)
	}

	// Expansion/bomb protection.
	inputSize := len(input)
	if opts.IgnoreWhitespace != nil && *opts.IgnoreWhitespace {
		inputSize = len(normalized)
	}
	if float64(len(out)) > float64(inputSize)*maxRatio {
		return DecodingResult{}, withInputFormat(addDetails(newError(OperationDecode, "OUTPUT_TOO_LARGE", "decoded output exceeds max_expansion_ratio"), map[string]any{"max_expansion_ratio": maxRatio, "input_size": inputSize, "output_size": len(out)}), format)
	}
	if len(out) > maxDecoded {
		return DecodingResult{}, withInputFormat(addDetails(newError(OperationDecode, "OUTPUT_TOO_LARGE", fmt.Sprintf("decoded output exceeds max_decoded_size (%d)", maxDecoded)), map[string]any{"max_decoded_size": maxDecoded, "output_size": len(out)}), format)
	}

	checksum, checksumAlg, err := computeChecksumIfRequested(out, opts.ComputeChecksum)
	if err != nil {
		return DecodingResult{}, withInputFormat(err, format)
	}

	return DecodingResult{
		Data:               out,
		Format:             format,
		InputSize:          len(input),
		OutputSize:         len(out),
		Checksum:           checksum,
		ChecksumAlgorithm:  checksumAlg,
		Warnings:           warnings,
		CorrectionsApplied: 0,
	}, nil
}

func decodeBase64(s string, urlsafe bool, validatePadding bool) ([]byte, *FulencodeError) {
	original := s

	// If padding is present, it must be trailing '=' and at most 2 chars.
	if idx := strings.IndexByte(original, '='); idx >= 0 {
		for i := idx; i < len(original); i++ {
			if original[i] != '=' {
				return nil, newError(OperationDecode, "INVALID_ENCODING", "invalid base64 padding")
			}
		}
		if len(original)-idx > 2 {
			return nil, newError(OperationDecode, "INVALID_ENCODING", "invalid base64 padding")
		}
	}

	// Always accept missing padding by normalizing to a multiple of 4.
	s = strings.TrimRight(s, "=")
	pad := (4 - (len(s) % 4)) % 4
	s = s + strings.Repeat("=", pad)

	if validatePadding {
		// After normalization, padding must decode cleanly.
		if len(s)%4 != 0 {
			return nil, newError(OperationDecode, "INVALID_ENCODING", "invalid base64 length")
		}
	}

	var enc *base64.Encoding
	if urlsafe {
		enc = base64.URLEncoding
	} else {
		enc = base64.StdEncoding
	}
	out, err := enc.DecodeString(s)
	if err != nil {
		return nil, newError(OperationDecode, "INVALID_ENCODING", err.Error())
	}
	return out, nil
}

func decodeBase32(s string, hexAlphabet bool, validatePadding bool) ([]byte, *FulencodeError) {
	var enc *base32.Encoding
	if hexAlphabet {
		enc = base32.HexEncoding
	} else {
		enc = base32.StdEncoding
	}

	if validatePadding {
		if len(s)%8 != 0 {
			return nil, newError(OperationDecode, "INVALID_ENCODING", "invalid base32 length")
		}
		if idx := strings.IndexByte(s, '='); idx >= 0 {
			for i := idx; i < len(s); i++ {
				if s[i] != '=' {
					return nil, newError(OperationDecode, "INVALID_ENCODING", "invalid base32 padding")
				}
			}
		}
	} else {
		s = strings.TrimRight(s, "=")
		enc = enc.WithPadding(base32.NoPadding)
	}

	out, err := enc.DecodeString(s)
	if err != nil {
		return nil, newError(OperationDecode, "INVALID_ENCODING", err.Error())
	}
	return out, nil
}

func encodeSingleByte(s string, enc encoding.Encoding, mode DecodeOnErrorMode) ([]byte, int, *FulencodeError) {
	// Fast path: strict encoding via transformer.
	if mode == "" || mode == DecodeStrict {
		out, _, err := transform.String(enc.NewEncoder(), s)
		if err != nil {
			return nil, 0, newError(OperationDecode, "INVALID_ENCODING", err.Error())
		}
		return []byte(out), 0, nil
	}

	if mode == DecodeReplace {
		out, _, err := transform.String(encoding.ReplaceUnsupported(enc.NewEncoder()), s)
		if err != nil {
			return nil, 0, newError(OperationDecode, "INVALID_ENCODING", err.Error())
		}
		return []byte(out), 1, nil
	}

	if mode == DecodeIgnore {
		// Encode rune-by-rune, dropping unsupported runes.
		tr := enc.NewEncoder()
		buf := make([]byte, 0, len(s))
		corrections := 0
		for _, r := range s {
			chunk := string(r)
			out, _, err := transform.String(tr, chunk)
			if err != nil {
				corrections++
				continue
			}
			buf = append(buf, out...)
		}
		return buf, corrections, nil
	}

	return nil, 0, newError(OperationDecode, "INVALID_OPTIONS", "unknown on_error mode")
}

func encodeASCIIString(s string, mode DecodeOnErrorMode) ([]byte, int, *FulencodeError) {
	corrections := 0
	buf := make([]byte, 0, len(s))
	for i, r := range s {
		if r <= 0x7F {
			buf = append(buf, s[i])
			continue
		}
		switch mode {
		case "", DecodeStrict:
			return nil, 0, addDetails(newError(OperationDecode, "INVALID_ENCODING", "character not representable in ascii"), map[string]any{"codepoint_offset": i})
		case DecodeReplace:
			buf = append(buf, '?')
			corrections++
		case DecodeIgnore:
			corrections++
		default:
			return nil, 0, newError(OperationDecode, "INVALID_OPTIONS", "unknown on_error mode")
		}
	}
	return buf, corrections, nil
}

// bytesToIntSlicePreview is used for error details (invalid_bytes). It avoids
// allocating huge slices for large inputs.
func bytesToIntSlicePreview(b []byte, max int) []int {
	if max <= 0 || len(b) <= max {
		out := make([]int, len(b))
		for i, v := range b {
			out[i] = int(v)
		}
		return out
	}
	out := make([]int, max)
	for i := 0; i < max; i++ {
		out[i] = int(b[i])
	}
	return out
}

// Avoid importing utf8 in other files.
var _ = utf8.RuneError
