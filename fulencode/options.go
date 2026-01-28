package fulencode

// NOTE: Options mirror Crucible schemas (snake_case JSON tags) but are
// intended for Go callers. Nil pointers mean "use default".

type OnErrorMode string

const (
	Strict  OnErrorMode = "strict"
	Replace OnErrorMode = "replace"
	Ignore  OnErrorMode = "ignore"
)

type DecodeOnErrorMode string

const (
	DecodeStrict   DecodeOnErrorMode = "strict"
	DecodeReplace  DecodeOnErrorMode = "replace"
	DecodeIgnore   DecodeOnErrorMode = "ignore"
	DecodeFallback DecodeOnErrorMode = "fallback"
)

type EncodeOptions struct {
	Padding        *bool  `json:"padding,omitempty"`
	Case           string `json:"case,omitempty"` // "upper" | "lower" (hex)
	LineLength     *int   `json:"line_length,omitempty"`
	LineEnding     string `json:"line_ending,omitempty"` // "\n" | "\r\n"
	MaxEncodedSize int    `json:"max_encoded_size,omitempty"`

	ComputeChecksum string `json:"compute_checksum,omitempty"`
	EmbedChecksum   *bool  `json:"embed_checksum,omitempty"`

	OnError OnErrorMode `json:"on_error,omitempty"`
}

type DecodeOptions struct {
	VerifyChecksum    *bool             `json:"verify_checksum,omitempty"`
	ComputeChecksum   string            `json:"compute_checksum,omitempty"`
	MaxDecodedSize    int               `json:"max_decoded_size,omitempty"`
	MaxExpansionRatio float64           `json:"max_expansion_ratio,omitempty"`
	OnError           DecodeOnErrorMode `json:"on_error,omitempty"`
	FallbackFormats   []EncodingFormat  `json:"fallback_formats,omitempty"`
	IgnoreWhitespace  *bool             `json:"ignore_whitespace,omitempty"`
	ValidatePadding   *bool             `json:"validate_padding,omitempty"`
}

type DetectOptions struct {
	MaxSampleSize      int     `json:"max_sample_size,omitempty"`
	MinConfidence      float64 `json:"min_confidence,omitempty"`
	RecognizeMultibase *bool   `json:"recognize_multibase,omitempty"`
}

type NormalizeOptions struct {
	WarnSemanticChange *bool `json:"warn_semantic_change,omitempty"`
	RejectZeroWidth    *bool `json:"reject_zero_width,omitempty"`
	RejectBidiControls *bool `json:"reject_bidi_controls,omitempty"`
	MaxCombiningMarks  int   `json:"max_combining_marks,omitempty"`
	StripAccents       *bool `json:"strip_accents,omitempty"`
	CaseFold           *bool `json:"case_fold,omitempty"`
	RemovePunctuation  *bool `json:"remove_punctuation,omitempty"`
	CompressWhitespace *bool `json:"compress_whitespace,omitempty"`
}
