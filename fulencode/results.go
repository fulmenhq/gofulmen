package fulencode

// EncodingResult is the result payload for Encode().
type EncodingResult struct {
	Data              string         `json:"data"`
	Format            EncodingFormat `json:"format"`
	InputSize         int            `json:"input_size"`
	OutputSize        int            `json:"output_size"`
	Checksum          string         `json:"checksum,omitempty"`
	ChecksumAlgorithm string         `json:"checksum_algorithm,omitempty"`
	Warnings          []string       `json:"warnings"`
}

// DecodingResult is the result payload for Decode().
type DecodingResult struct {
	Data               []byte         `json:"-"`
	Format             EncodingFormat `json:"format"`
	InputSize          int            `json:"input_size"`
	OutputSize         int            `json:"output_size"`
	Checksum           string         `json:"checksum,omitempty"`
	ChecksumVerified   *bool          `json:"checksum_verified,omitempty"`
	ChecksumAlgorithm  string         `json:"checksum_algorithm,omitempty"`
	Warnings           []string       `json:"warnings"`
	CorrectionsApplied int            `json:"corrections_applied"`
}

type DetectionResult struct {
	Encoding        *string         `json:"encoding"`
	Confidence      float64         `json:"confidence"`
	Level           ConfidenceLevel `json:"level"`
	MultibasePrefix *string         `json:"multibase_prefix"`
	Warnings        []string        `json:"warnings"`
}

type SemanticChange struct {
	Position   int    `json:"position"`
	Original   string `json:"original"`
	Normalized string `json:"normalized"`
	Reason     string `json:"reason"`
}

type NormalizationResult struct {
	Text                   string           `json:"text"`
	Profile                string           `json:"profile"`
	InputLength            int              `json:"input_length"`
	OutputLength           int              `json:"output_length"`
	TransformationsApplied []string         `json:"transformations_applied"`
	SemanticChanges        []SemanticChange `json:"semantic_changes"`
	Warnings               []string         `json:"warnings"`
}

type BOMResult struct {
	BOMType         *string `json:"bom_type"`
	ByteLength      int     `json:"byte_length"`
	EncodingImplied *string `json:"encoding_implied"`
}
