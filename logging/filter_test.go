package logging

import "testing"

func TestSeverityFilter_GE(t *testing.T) {
	filter := NewSeverityFilter(GE, WARN)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, false}, // 0 < 30
		{DEBUG, false}, // 10 < 30
		{INFO, false},  // 20 < 30
		{WARN, true},   // 30 >= 30
		{ERROR, true},  // 40 >= 30
		{FATAL, true},  // 50 >= 30
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("GE filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestSeverityFilter_LE(t *testing.T) {
	filter := NewSeverityFilter(LE, INFO)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, true},  // 0 <= 20
		{DEBUG, true},  // 10 <= 20
		{INFO, true},   // 20 <= 20
		{WARN, false},  // 30 > 20
		{ERROR, false}, // 40 > 20
		{FATAL, false}, // 50 > 20
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("LE filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestSeverityFilter_EQ(t *testing.T) {
	filter := NewSeverityFilter(EQ, ERROR)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, false},
		{DEBUG, false},
		{INFO, false},
		{WARN, false},
		{ERROR, true}, // Only ERROR matches
		{FATAL, false},
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("EQ filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestSeverityFilter_NE(t *testing.T) {
	filter := NewSeverityFilter(NE, INFO)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, true},
		{DEBUG, true},
		{INFO, false}, // INFO does not match != INFO
		{WARN, true},
		{ERROR, true},
		{FATAL, true},
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("NE filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestSeverityFilter_GT(t *testing.T) {
	filter := NewSeverityFilter(GT, WARN)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, false}, // 0 not > 30
		{DEBUG, false}, // 10 not > 30
		{INFO, false},  // 20 not > 30
		{WARN, false},  // 30 not > 30
		{ERROR, true},  // 40 > 30
		{FATAL, true},  // 50 > 30
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("GT filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestSeverityFilter_LT(t *testing.T) {
	filter := NewSeverityFilter(LT, ERROR)

	tests := []struct {
		severity Severity
		expected bool
	}{
		{TRACE, true},  // 0 < 40
		{DEBUG, true},  // 10 < 40
		{INFO, true},   // 20 < 40
		{WARN, true},   // 30 < 40
		{ERROR, false}, // 40 not < 40
		{FATAL, false}, // 50 not < 40
	}

	for _, tt := range tests {
		result := filter.Matches(tt.severity)
		if result != tt.expected {
			t.Errorf("LT filter: severity %s, expected %v, got %v", tt.severity, tt.expected, result)
		}
	}
}

func TestMinLevel(t *testing.T) {
	filter := MinLevel(WARN)

	if filter.Operator != GE {
		t.Errorf("Expected GE operator, got %s", filter.Operator)
	}

	if filter.Level != WARN {
		t.Errorf("Expected WARN level, got %s", filter.Level)
	}

	// Should match WARN and above
	if !filter.Matches(WARN) {
		t.Error("MinLevel(WARN) should match WARN")
	}
	if !filter.Matches(ERROR) {
		t.Error("MinLevel(WARN) should match ERROR")
	}
	if filter.Matches(INFO) {
		t.Error("MinLevel(WARN) should not match INFO")
	}
}

func TestMaxLevel(t *testing.T) {
	filter := MaxLevel(INFO)

	if filter.Operator != LE {
		t.Errorf("Expected LE operator, got %s", filter.Operator)
	}

	if filter.Level != INFO {
		t.Errorf("Expected INFO level, got %s", filter.Level)
	}

	// Should match INFO and below
	if !filter.Matches(INFO) {
		t.Error("MaxLevel(INFO) should match INFO")
	}
	if !filter.Matches(DEBUG) {
		t.Error("MaxLevel(INFO) should match DEBUG")
	}
	if filter.Matches(WARN) {
		t.Error("MaxLevel(INFO) should not match WARN")
	}
}

func TestOnlyLevel(t *testing.T) {
	filter := OnlyLevel(ERROR)

	if filter.Operator != EQ {
		t.Errorf("Expected EQ operator, got %s", filter.Operator)
	}

	if filter.Level != ERROR {
		t.Errorf("Expected ERROR level, got %s", filter.Level)
	}

	// Should match only ERROR
	if !filter.Matches(ERROR) {
		t.Error("OnlyLevel(ERROR) should match ERROR")
	}
	if filter.Matches(WARN) {
		t.Error("OnlyLevel(ERROR) should not match WARN")
	}
	if filter.Matches(FATAL) {
		t.Error("OnlyLevel(ERROR) should not match FATAL")
	}
}
