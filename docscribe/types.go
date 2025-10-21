package docscribe

// DocumentInfo provides quick inspection results without full parsing.
// This is returned by InspectDocument for fast analysis of document structure
// and characteristics without the overhead of complete parsing.
type DocumentInfo struct {
	// HasFrontmatter indicates whether YAML frontmatter is present
	HasFrontmatter bool `json:"has_frontmatter"`

	// HeaderCount is the number of markdown headers detected
	HeaderCount int `json:"header_count"`

	// Format describes the detected content format
	// Possible values: "markdown", "yaml", "json", "toml", "text", "multi-yaml", "multi-markdown"
	Format string `json:"format"`

	// LineCount is the total number of lines in the document
	LineCount int `json:"line_count"`

	// EstimatedSections is a heuristic estimate of major document sections
	// based on header hierarchy and structure
	EstimatedSections int `json:"estimated_sections"`
}

// Header represents a markdown header with its metadata.
// This is returned by ExtractHeaders for navigation, TOC generation,
// and document structure analysis.
type Header struct {
	// Level is the header depth (1-6 for H1-H6)
	Level int `json:"level"`

	// Text is the header content without the # prefix or underline
	Text string `json:"text"`

	// Anchor is the URL-safe slug for linking (e.g., "my-header" from "My Header")
	Anchor string `json:"anchor"`

	// LineNumber is the 1-based line number where this header appears
	LineNumber int `json:"line_number"`
}

// Format constants for document format detection
const (
	FormatMarkdown      = "markdown"
	FormatYAML          = "yaml"
	FormatJSON          = "json"
	FormatTOML          = "toml"
	FormatText          = "text"
	FormatMultiYAML     = "multi-yaml"
	FormatMultiMarkdown = "multi-markdown"
)
