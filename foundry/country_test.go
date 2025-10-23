package foundry

import (
	"testing"
)

func TestGetCountry(t *testing.T) {
	tests := []struct {
		alpha2       string
		expectedName string
	}{
		{"US", "United States of America"},
		{"CA", "Canada"},
		{"JP", "Japan"},
		{"DE", "Germany"},
		{"BR", "Brazil"},
	}

	for _, tt := range tests {
		t.Run(tt.alpha2, func(t *testing.T) {
			country, err := GetCountry(tt.alpha2)
			if err != nil {
				t.Fatalf("Failed to get country: %v", err)
			}

			if country == nil {
				t.Fatalf("Expected non-nil country for %q", tt.alpha2)
			}

			if country.Alpha2 != tt.alpha2 {
				t.Errorf("Expected Alpha2 %q, got %q", tt.alpha2, country.Alpha2)
			}

			if country.Name != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, country.Name)
			}
		})
	}
}

func TestGetCountry_NotFound(t *testing.T) {
	country, err := GetCountry("XX")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if country != nil {
		t.Error("Expected nil country for non-existent code")
	}
}

func TestGetCountryByAlpha3(t *testing.T) {
	tests := []struct {
		alpha3       string
		expectedName string
	}{
		{"USA", "United States of America"},
		{"CAN", "Canada"},
		{"JPN", "Japan"},
		{"DEU", "Germany"},
		{"BRA", "Brazil"},
	}

	for _, tt := range tests {
		t.Run(tt.alpha3, func(t *testing.T) {
			country, err := GetCountryByAlpha3(tt.alpha3)
			if err != nil {
				t.Fatalf("Failed to get country: %v", err)
			}

			if country == nil {
				t.Fatalf("Expected non-nil country for %q", tt.alpha3)
			}

			if country.Alpha3 != tt.alpha3 {
				t.Errorf("Expected Alpha3 %q, got %q", tt.alpha3, country.Alpha3)
			}

			if country.Name != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, country.Name)
			}
		})
	}
}

func TestGetCountryByAlpha3_NotFound(t *testing.T) {
	country, err := GetCountryByAlpha3("XXX")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if country != nil {
		t.Error("Expected nil country for non-existent code")
	}
}

func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		// Alpha2 codes
		{"US", true},
		{"us", true}, // case-insensitive
		{"CA", true},
		{"ca", true}, // case-insensitive
		// Alpha3 codes
		{"USA", true},
		{"usa", true}, // case-insensitive
		{"CAN", true},
		{"can", true}, // case-insensitive
		// Numeric codes
		{"840", true}, // US numeric
		{"124", true}, // CA numeric
		{"076", true}, // BR numeric
		{"76", true},  // BR numeric (without leading zero)
		{"392", true}, // JP numeric
		{"276", true}, // DE numeric
		// Invalid codes
		{"XX", false},
		{"XXX", false},
		{"999", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := ValidateCountryCode(tt.code)
			if result != tt.valid {
				t.Errorf("ValidateCountryCode(%q) = %v, want %v", tt.code, result, tt.valid)
			}
		})
	}
}

func TestListCountries(t *testing.T) {
	countries, err := ListCountries()
	if err != nil {
		t.Fatalf("Failed to list countries: %v", err)
	}

	if len(countries) == 0 {
		t.Fatal("Expected at least one country")
	}

	// Verify expected countries are in the list
	expectedCodes := map[string]bool{
		"US": false,
		"CA": false,
		"JP": false,
		"DE": false,
		"BR": false,
	}

	for _, country := range countries {
		if _, exists := expectedCodes[country.Alpha2]; exists {
			expectedCodes[country.Alpha2] = true
		}
	}

	for code, found := range expectedCodes {
		if !found {
			t.Errorf("Expected country %q to be in the list", code)
		}
	}
}

func TestCountry_MatchesCode(t *testing.T) {
	country := &Country{
		Alpha2: "US",
		Alpha3: "USA",
		Name:   "United States of America",
	}

	tests := []struct {
		code    string
		matches bool
	}{
		{"US", true},
		{"us", true},
		{"USA", true},
		{"usa", true},
		{"CA", false},
		{"CAN", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := country.MatchesCode(tt.code)
			if result != tt.matches {
				t.Errorf("MatchesCode(%q) = %v, want %v", tt.code, result, tt.matches)
			}
		})
	}
}

func TestCatalog_GetCountry(t *testing.T) {
	catalog := GetDefaultCatalog()

	country, err := catalog.GetCountry("US")
	if err != nil {
		t.Fatalf("Failed to get country: %v", err)
	}

	if country == nil {
		t.Fatal("Expected non-nil country")
	}

	if country.Alpha2 != "US" {
		t.Errorf("Expected Alpha2 'US', got %q", country.Alpha2)
	}
}

func TestCatalog_ListCountries(t *testing.T) {
	catalog := GetDefaultCatalog()

	countries, err := catalog.ListCountries()
	if err != nil {
		t.Fatalf("Failed to list countries: %v", err)
	}

	if len(countries) < 5 {
		t.Errorf("Expected at least 5 countries, got %d", len(countries))
	}
}

func BenchmarkGetCountry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetCountry("US") //nolint:errcheck // benchmark ignores return values
	}
}

func BenchmarkGetCountryByAlpha3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetCountryByAlpha3("USA") //nolint:errcheck
	}
}

func BenchmarkValidateCountryCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetCountry("US") //nolint:errcheck
	}
}
