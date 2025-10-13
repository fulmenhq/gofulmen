package logging

// ComparisonOperator defines severity comparison operations
type ComparisonOperator string

const (
	GE ComparisonOperator = "GE" // Greater than or equal
	LE ComparisonOperator = "LE" // Less than or equal
	GT ComparisonOperator = "GT" // Greater than
	LT ComparisonOperator = "LT" // Less than
	EQ ComparisonOperator = "EQ" // Equal
	NE ComparisonOperator = "NE" // Not equal
)

// SeverityFilter represents severity-based filtering with comparison operators
type SeverityFilter struct {
	Operator ComparisonOperator `json:"operator"`
	Level    Severity           `json:"level"`
}

// Matches checks if a severity passes this filter
func (f *SeverityFilter) Matches(severity Severity) bool {
	currentLevel := severity.Level()
	filterLevel := f.Level.Level()

	switch f.Operator {
	case GE:
		return currentLevel >= filterLevel
	case LE:
		return currentLevel <= filterLevel
	case GT:
		return currentLevel > filterLevel
	case LT:
		return currentLevel < filterLevel
	case EQ:
		return currentLevel == filterLevel
	case NE:
		return currentLevel != filterLevel
	default:
		return false
	}
}

// NewSeverityFilter creates a filter with the given operator and level
func NewSeverityFilter(operator ComparisonOperator, level Severity) *SeverityFilter {
	return &SeverityFilter{
		Operator: operator,
		Level:    level,
	}
}

// MinLevel creates a filter for minimum severity (level >= threshold)
func MinLevel(level Severity) *SeverityFilter {
	return &SeverityFilter{
		Operator: GE,
		Level:    level,
	}
}

// MaxLevel creates a filter for maximum severity (level <= threshold)
func MaxLevel(level Severity) *SeverityFilter {
	return &SeverityFilter{
		Operator: LE,
		Level:    level,
	}
}

// OnlyLevel creates a filter for exact severity match
func OnlyLevel(level Severity) *SeverityFilter {
	return &SeverityFilter{
		Operator: EQ,
		Level:    level,
	}
}
