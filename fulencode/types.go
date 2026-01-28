package fulencode

import cruciblefulencode "github.com/fulmenhq/crucible/fulencode"

// Re-export Crucible generated enums (SSOT).
//
// NOTE: Do not hand-write these enums; import them from Crucible.

type EncodingFormat = cruciblefulencode.EncodingFormat

const (
	BASE64     = cruciblefulencode.EncodingFormatBase64
	BASE64URL  = cruciblefulencode.EncodingFormatBase64url
	BASE64_RAW = cruciblefulencode.EncodingFormatBase64Raw
	BASE32     = cruciblefulencode.EncodingFormatBase32
	BASE32HEX  = cruciblefulencode.EncodingFormatBase32hex
	HEX        = cruciblefulencode.EncodingFormatHex
	UTF8       = cruciblefulencode.EncodingFormatUtf8
	UTF16LE    = cruciblefulencode.EncodingFormatUtf16le
	UTF16BE    = cruciblefulencode.EncodingFormatUtf16be
	ISO_8859_1 = cruciblefulencode.EncodingFormatIso88591
	CP1252     = cruciblefulencode.EncodingFormatCp1252
	ASCII      = cruciblefulencode.EncodingFormatAscii
)

func ValidateEncodingFormat(v EncodingFormat) error {
	return cruciblefulencode.ValidateEncodingFormat(v)
}

type NormalizationProfile = cruciblefulencode.NormalizationProfile

const (
	NFC               = cruciblefulencode.NormalizationProfileNfc
	NFD               = cruciblefulencode.NormalizationProfileNfd
	NFKC              = cruciblefulencode.NormalizationProfileNfkc
	NFKD              = cruciblefulencode.NormalizationProfileNfkd
	SAFE_IDENTIFIERS  = cruciblefulencode.NormalizationProfileSafeIdentifiers
	SEARCH_OPTIMIZED  = cruciblefulencode.NormalizationProfileSearchOptimized
	FILENAME_SAFE     = cruciblefulencode.NormalizationProfileFilenameSafe
	TEXT_SAFE         = cruciblefulencode.NormalizationProfileTextSafe
	LEGACY_COMPATIBLE = cruciblefulencode.NormalizationProfileLegacyCompatible
)

func ValidateNormalizationProfile(v NormalizationProfile) error {
	return cruciblefulencode.ValidateNormalizationProfile(v)
}

type ConfidenceLevel = cruciblefulencode.ConfidenceLevel

const (
	ConfidenceHigh   = cruciblefulencode.ConfidenceLevelHigh
	ConfidenceMedium = cruciblefulencode.ConfidenceLevelMedium
	ConfidenceLow    = cruciblefulencode.ConfidenceLevelLow
)

func ValidateConfidenceLevel(v ConfidenceLevel) error {
	return cruciblefulencode.ValidateConfidenceLevel(v)
}
