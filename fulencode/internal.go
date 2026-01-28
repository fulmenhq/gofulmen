package fulencode

import "strings"

const (
	DefaultMaxDecodedSizeBytes = 100 * 1024 * 1024
	DefaultMaxEncodedSizeBytes = 500 * 1024 * 1024
	DefaultMaxExpansionRatio   = 10.0
	DefaultDetectMaxSampleSize = 8192
	DefaultDetectMinConfidence = 0.0
	DefaultMaxCombiningMarks   = 10
)

func boolPtr(b bool) *bool { return &b }

func applyEncodeDefaults(opts *EncodeOptions) *EncodeOptions {
	if opts == nil {
		opts = &EncodeOptions{}
	}
	if opts.Padding == nil {
		opts.Padding = boolPtr(true)
	}
	if opts.LineEnding == "" {
		opts.LineEnding = "\n"
	}
	if opts.MaxEncodedSize == 0 {
		opts.MaxEncodedSize = DefaultMaxEncodedSizeBytes
	}
	if opts.OnError == "" {
		opts.OnError = Strict
	}
	if opts.Case == "" {
		opts.Case = "lower"
	}
	return opts
}

func applyDecodeDefaults(format EncodingFormat, opts *DecodeOptions) *DecodeOptions {
	if opts == nil {
		opts = &DecodeOptions{}
	}
	if opts.MaxDecodedSize == 0 {
		opts.MaxDecodedSize = DefaultMaxDecodedSizeBytes
	}
	if opts.MaxExpansionRatio == 0 {
		opts.MaxExpansionRatio = DefaultMaxExpansionRatio
	}
	if opts.OnError == "" {
		opts.OnError = DecodeStrict
	}
	if opts.IgnoreWhitespace == nil {
		ignore := format == BASE64 || format == BASE64URL || format == BASE64_RAW || format == BASE32 || format == BASE32HEX || format == HEX
		opts.IgnoreWhitespace = boolPtr(ignore)
	}
	if opts.ValidatePadding == nil {
		validate := format == BASE64 || format == BASE64URL || format == BASE32 || format == BASE32HEX
		opts.ValidatePadding = boolPtr(validate)
	}
	return opts
}

func applyDetectDefaults(opts *DetectOptions) *DetectOptions {
	if opts == nil {
		opts = &DetectOptions{}
	}
	if opts.MaxSampleSize == 0 {
		opts.MaxSampleSize = DefaultDetectMaxSampleSize
	}
	if opts.MinConfidence == 0 {
		opts.MinConfidence = DefaultDetectMinConfidence
	}
	if opts.RecognizeMultibase == nil {
		opts.RecognizeMultibase = boolPtr(false)
	}
	return opts
}

func applyNormalizeDefaults(opts *NormalizeOptions) *NormalizeOptions {
	if opts == nil {
		opts = &NormalizeOptions{}
	}
	if opts.MaxCombiningMarks == 0 {
		opts.MaxCombiningMarks = DefaultMaxCombiningMarks
	}
	if opts.RejectZeroWidth == nil {
		opts.RejectZeroWidth = boolPtr(true)
	}
	if opts.RejectBidiControls == nil {
		opts.RejectBidiControls = boolPtr(true)
	}
	if opts.WarnSemanticChange == nil {
		opts.WarnSemanticChange = boolPtr(false)
	}
	return opts
}

func stripASCIIWhitespace(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func wrapLines(s string, lineLength int, lineEnding string) string {
	if lineLength <= 0 || len(s) <= lineLength {
		return s
	}
	if lineEnding == "" {
		lineEnding = "\n"
	}
	var b strings.Builder
	b.Grow(len(s) + (len(s)/lineLength)*len(lineEnding))
	for i := 0; i < len(s); i += lineLength {
		end := i + lineLength
		if end > len(s) {
			end = len(s)
		}
		b.WriteString(s[i:end])
		if end < len(s) {
			b.WriteString(lineEnding)
		}
	}
	return b.String()
}
