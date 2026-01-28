package fulencode

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func Encode(data []byte, format EncodingFormat, options *EncodeOptions) (EncodingResult, error) {
	if err := ValidateEncodingFormat(format); err != nil {
		return EncodingResult{}, withOutputFormat(newError(OperationEncode, "INVALID_FORMAT", err.Error()), format)
	}
	opts := applyEncodeDefaults(options)
	if opts.EmbedChecksum != nil && *opts.EmbedChecksum {
		return EncodingResult{}, withOutputFormat(newError(OperationEncode, "UNSUPPORTED_FEATURE", "embed_checksum is not implemented"), format)
	}

	warnings := make([]string, 0)
	var out string

	switch format {
	case BASE64:
		enc := base64.StdEncoding
		if opts.Padding != nil && !*opts.Padding {
			enc = base64.RawStdEncoding
		}
		out = enc.EncodeToString(data)
	case BASE64URL:
		enc := base64.URLEncoding
		if opts.Padding != nil && !*opts.Padding {
			enc = base64.RawURLEncoding
		}
		out = enc.EncodeToString(data)
	case BASE64_RAW:
		out = base64.RawStdEncoding.EncodeToString(data)
	case HEX:
		out = hex.EncodeToString(data)
		if opts.Case == "upper" {
			out = strings.ToUpper(out)
		}
	case BASE32:
		enc := base32.StdEncoding
		if opts.Padding != nil && !*opts.Padding {
			enc = enc.WithPadding(base32.NoPadding)
		}
		out = enc.EncodeToString(data)
	case BASE32HEX:
		enc := base32.HexEncoding
		if opts.Padding != nil && !*opts.Padding {
			enc = enc.WithPadding(base32.NoPadding)
		}
		out = enc.EncodeToString(data)
	case UTF8:
		s, err := decodeUTF8Bytes(data, opts.OnError)
		if err != nil {
			return EncodingResult{}, withOutputFormat(err, format)
		}
		out = s
	case UTF16LE:
		s, err := decodeUTF16Bytes(data, true, opts.OnError)
		if err != nil {
			return EncodingResult{}, withOutputFormat(err, format)
		}
		out = s
	case UTF16BE:
		s, err := decodeUTF16Bytes(data, false, opts.OnError)
		if err != nil {
			return EncodingResult{}, withOutputFormat(err, format)
		}
		out = s
	case ISO_8859_1:
		decoded, err := charmap.ISO8859_1.NewDecoder().Bytes(data)
		if err != nil {
			return EncodingResult{}, withOutputFormat(newError(OperationEncode, "INVALID_ENCODING", err.Error()), format)
		}
		out = string(decoded)
	case CP1252:
		decoded, err := charmap.Windows1252.NewDecoder().Bytes(data)
		if err != nil {
			return EncodingResult{}, withOutputFormat(newError(OperationEncode, "INVALID_ENCODING", err.Error()), format)
		}
		out = string(decoded)
	case ASCII:
		s, err := decodeASCIIBytes(data, opts.OnError)
		if err != nil {
			return EncodingResult{}, withOutputFormat(err, format)
		}
		out = s
	default:
		return EncodingResult{}, withOutputFormat(newError(OperationEncode, "UNSUPPORTED_FORMAT", fmt.Sprintf("unsupported encoding format: %s", format)), format)
	}

	if opts.LineLength != nil {
		out = wrapLines(out, *opts.LineLength, opts.LineEnding)
	}
	if len(out) > opts.MaxEncodedSize {
		return EncodingResult{}, withOutputFormat(addDetails(newError(OperationEncode, "OUTPUT_TOO_LARGE", fmt.Sprintf("encoded output exceeds max_encoded_size (%d)", opts.MaxEncodedSize)), map[string]any{"max_encoded_size": opts.MaxEncodedSize, "output_size": len(out)}), format)
	}

	checksum, checksumAlg, err := computeChecksumIfRequested(data, opts.ComputeChecksum)
	if err != nil {
		return EncodingResult{}, withOutputFormat(err, format)
	}

	return EncodingResult{
		Data:              out,
		Format:            format,
		InputSize:         len(data),
		OutputSize:        len(out),
		Checksum:          checksum,
		ChecksumAlgorithm: checksumAlg,
		Warnings:          warnings,
	}, nil
}

func decodeUTF8Bytes(data []byte, mode OnErrorMode) (string, *FulencodeError) {
	if utf8.Valid(data) {
		return string(data), nil
	}

	switch mode {
	case Strict, "":
		return "", addDetails(newError(OperationEncode, "INVALID_ENCODING", "invalid utf-8"), map[string]any{"invalid_bytes": bytesToIntSlicePreview(data, 16)})
	case Replace:
		return string(bytes.ToValidUTF8(data, []byte("\uFFFD"))), nil
	case Ignore:
		// Drop invalid sequences (byte-by-byte for malformed runs).
		out := make([]byte, 0, len(data))
		for len(data) > 0 {
			r, size := utf8.DecodeRune(data)
			if r == utf8.RuneError && size == 1 {
				data = data[1:]
				continue
			}
			out = append(out, data[:size]...)
			data = data[size:]
		}
		return string(out), nil
	default:
		return "", newError(OperationEncode, "INVALID_OPTIONS", "unknown on_error mode")
	}
}

func decodeASCIIBytes(data []byte, mode OnErrorMode) (string, *FulencodeError) {
	// ASCII is a strict subset of UTF-8.
	for i, b := range data {
		if b <= 0x7F {
			continue
		}
		switch mode {
		case Strict, "":
			return "", addDetails(newError(OperationEncode, "INVALID_ENCODING", "invalid ascii byte"), map[string]any{"byte_offset": i, "invalid_bytes": []int{int(b)}})
		case Replace:
			out := make([]rune, 0, len(data))
			for _, bb := range data {
				if bb <= 0x7F {
					out = append(out, rune(bb))
				} else {
					out = append(out, utf8.RuneError)
				}
			}
			return string(out), nil
		case Ignore:
			out := make([]byte, 0, len(data))
			for _, bb := range data {
				if bb <= 0x7F {
					out = append(out, bb)
				}
			}
			return string(out), nil
		default:
			return "", newError(OperationEncode, "INVALID_OPTIONS", "unknown on_error mode")
		}
	}
	return string(data), nil
}
