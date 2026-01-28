package fulencode

import (
	"encoding/binary"
	"unicode/utf16"
)

func decodeUTF16Bytes(data []byte, littleEndian bool, mode OnErrorMode) (string, *FulencodeError) {
	if len(data)%2 != 0 {
		switch mode {
		case Ignore:
			data = data[:len(data)-1]
		case Replace:
			data = append(data, 0)
		default:
			return "", newError(OperationEncode, "INVALID_ENCODING", "utf-16 byte length must be even")
		}
	}

	var bo binary.ByteOrder = binary.BigEndian
	if littleEndian {
		bo = binary.LittleEndian
	}

	codeUnits := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		codeUnits = append(codeUnits, bo.Uint16(data[i:i+2]))
	}

	// Manual decode to support strict/replace/ignore on unpaired surrogates.
	runes := make([]rune, 0, len(codeUnits))
	for i := 0; i < len(codeUnits); i++ {
		u := codeUnits[i]
		switch {
		case u >= 0xD800 && u <= 0xDBFF: // high surrogate
			if i+1 < len(codeUnits) {
				n := codeUnits[i+1]
				if n >= 0xDC00 && n <= 0xDFFF {
					runes = append(runes, utf16.DecodeRune(rune(u), rune(n)))
					i++
					continue
				}
			}
			// Unpaired high surrogate
			switch mode {
			case Strict, "":
				return "", addDetails(newError(OperationEncode, "INVALID_UTF16", "unpaired high surrogate"), map[string]any{"byte_offset": i * 2})
			case Replace:
				runes = append(runes, '\uFFFD')
			case Ignore:
				continue
			default:
				return "", newError(OperationEncode, "INVALID_OPTIONS", "unknown on_error mode")
			}
		case u >= 0xDC00 && u <= 0xDFFF: // low surrogate without high
			switch mode {
			case Strict, "":
				return "", addDetails(newError(OperationEncode, "INVALID_UTF16", "unpaired low surrogate"), map[string]any{"byte_offset": i * 2})
			case Replace:
				runes = append(runes, '\uFFFD')
			case Ignore:
				continue
			default:
				return "", newError(OperationEncode, "INVALID_OPTIONS", "unknown on_error mode")
			}
		default:
			runes = append(runes, rune(u))
		}
	}

	return string(runes), nil
}

func encodeUTF16String(s string, littleEndian bool) ([]byte, int, *FulencodeError) {
	var bo binary.ByteOrder = binary.BigEndian
	if littleEndian {
		bo = binary.LittleEndian
	}
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units)*2)
	for i, u := range units {
		bo.PutUint16(out[i*2:(i+1)*2], u)
	}
	return out, 0, nil
}
