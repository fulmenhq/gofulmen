package foundry

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestNewCountryCode_Valid tests creating CountryCode with valid inputs
func TestNewCountryCode_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // normalized form
	}{
		// Alpha-2 codes
		{"Alpha2_US", "US", "US"},
		{"Alpha2_CA", "CA", "CA"},
		{"Alpha2_JP", "JP", "JP"},
		{"Alpha2_DE", "DE", "DE"},
		{"Alpha2_BR", "BR", "BR"},

		// Alpha-3 codes
		{"Alpha3_USA", "USA", "USA"},
		{"Alpha3_CAN", "CAN", "CAN"},
		{"Alpha3_JPN", "JPN", "JPN"},
		{"Alpha3_DEU", "DEU", "DEU"},
		{"Alpha3_BRA", "BRA", "BRA"},

		// Numeric codes
		{"Numeric_840", "840", "840"},
		{"Numeric_124", "124", "124"},
		{"Numeric_392", "392", "392"},
		{"Numeric_276", "276", "276"},
		{"Numeric_076", "076", "076"},
		{"Numeric_76", "76", "076"}, // canonicalized to 3 digits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := NewCountryCode(tt.input)
			if err != nil {
				t.Fatalf("NewCountryCode(%q) returned error: %v", tt.input, err)
			}

			if string(code) != tt.expected {
				t.Errorf("NewCountryCode(%q) = %q, want %q", tt.input, code, tt.expected)
			}

			if !code.IsValid() {
				t.Errorf("Expected code %q to be valid", code)
			}
		})
	}
}

// TestNewCountryCode_CaseInsensitive tests case-insensitive validation
func TestNewCountryCode_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"us", "US"},
		{"Us", "US"},
		{"uS", "US"},
		{"usa", "USA"},
		{"Usa", "USA"},
		{"UsA", "USA"},
		{"can", "CAN"},
		{"CaN", "CAN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			code, err := NewCountryCode(tt.input)
			if err != nil {
				t.Fatalf("NewCountryCode(%q) returned error: %v", tt.input, err)
			}

			if string(code) != tt.expected {
				t.Errorf("NewCountryCode(%q) = %q, want %q (case normalization failed)", tt.input, code, tt.expected)
			}
		})
	}
}

// TestNewCountryCode_Invalid tests creating CountryCode with invalid inputs
func TestNewCountryCode_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Invalid_Alpha2", "XX"},
		{"Invalid_Alpha3", "XXX"},
		{"Invalid_Numeric", "999"},
		{"TooLong", "USAA"},
		{"TooShort", "U"},
		{"SpecialChars", "US!"},
		{"Spaces", "US "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCountryCode(tt.input)
			if err == nil {
				t.Errorf("NewCountryCode(%q) should return error for invalid input", tt.input)
			}
		})
	}
}

// TestMustCountryCode_Success tests MustCountryCode with valid inputs
func TestMustCountryCode_Success(t *testing.T) {
	tests := []string{"US", "USA", "840"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MustCountryCode(%q) panicked: %v", input, r)
				}
			}()

			code := MustCountryCode(input)
			if !code.IsValid() {
				t.Errorf("MustCountryCode(%q) returned invalid code", input)
			}
		})
	}
}

// TestMustCountryCode_Panic tests MustCountryCode panics on invalid input
func TestMustCountryCode_Panic(t *testing.T) {
	tests := []string{"", "XX", "XXX", "999"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustCountryCode(%q) should panic for invalid input", input)
				}
			}()

			MustCountryCode(input)
		})
	}
}

// TestCountryCode_Validate tests the Validate method
func TestCountryCode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		code    CountryCode
		wantErr bool
	}{
		{"Valid_US", "US", false},
		{"Valid_USA", "USA", false},
		{"Valid_840", "840", false},
		{"Empty", "", true},
		{"Invalid", "XX", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CountryCode(%q).Validate() error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

// TestCountryCode_IsValid tests the IsValid convenience method
func TestCountryCode_IsValid(t *testing.T) {
	tests := []struct {
		code  CountryCode
		valid bool
	}{
		{"US", true},
		{"USA", true},
		{"840", true},
		{"", false},
		{"XX", false},
		{"XXX", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.IsValid(); got != tt.valid {
				t.Errorf("CountryCode(%q).IsValid() = %v, want %v", tt.code, got, tt.valid)
			}
		})
	}
}

// TestCountryCode_String tests the String method
func TestCountryCode_String(t *testing.T) {
	code := CountryCode("US")
	if got := code.String(); got != "US" {
		t.Errorf("CountryCode.String() = %q, want %q", got, "US")
	}
}

// TestCountryCode_Country tests retrieving full country metadata
func TestCountryCode_Country(t *testing.T) {
	tests := []struct {
		code         CountryCode
		expectedName string
	}{
		{"US", "United States of America"},
		{"CA", "Canada"},
		{"JP", "Japan"},
		{"DE", "Germany"},
		{"BR", "Brazil"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			country, err := tt.code.Country()
			if err != nil {
				t.Fatalf("CountryCode(%q).Country() returned error: %v", tt.code, err)
			}

			if country == nil {
				t.Fatalf("CountryCode(%q).Country() returned nil", tt.code)
			}

			if country.Name != tt.expectedName {
				t.Errorf("CountryCode(%q).Country().Name = %q, want %q", tt.code, country.Name, tt.expectedName)
			}
		})
	}
}

// TestCountryCode_Country_Invalid tests Country() with invalid code
func TestCountryCode_Country_Invalid(t *testing.T) {
	code := CountryCode("XX")
	_, err := code.Country()
	if err == nil {
		t.Error("Expected error for invalid country code")
	}
}

// TestCountryCode_Country_Numeric tests Country() with numeric codes
func TestCountryCode_Country_Numeric(t *testing.T) {
	tests := []struct {
		code         CountryCode
		expectedName string
	}{
		{"840", "United States of America"}, // US
		{"076", "Brazil"},                   // Brazil (zero-padded)
		{"76", "Brazil"},                    // Brazil (without leading zero)
		{"392", "Japan"},                    // Japan
		{"124", "Canada"},                   // Canada
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			country, err := tt.code.Country()
			if err != nil {
				t.Fatalf("CountryCode(%q).Country() returned error: %v", tt.code, err)
			}

			if country == nil {
				t.Fatalf("CountryCode(%q).Country() returned nil", tt.code)
			}

			if country.Name != tt.expectedName {
				t.Errorf("CountryCode(%q).Country().Name = %q, want %q", tt.code, country.Name, tt.expectedName)
			}
		})
	}
}

// TestCountryCode_JSONMarshal tests JSON marshaling
func TestCountryCode_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		code     CountryCode
		expected string
	}{
		{"Alpha2", "US", `"US"`},
		{"Alpha3", "USA", `"USA"`},
		{"Numeric", "840", `"840"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.code)
			if err != nil {
				t.Fatalf("json.Marshal(%q) returned error: %v", tt.code, err)
			}

			if string(data) != tt.expected {
				t.Errorf("json.Marshal(%q) = %s, want %s", tt.code, data, tt.expected)
			}
		})
	}
}

// TestCountryCode_MarshalText_Error tests MarshalText with invalid codes
func TestCountryCode_MarshalText_Error(t *testing.T) {
	tests := []struct {
		name string
		code CountryCode
	}{
		{"Empty", ""},
		{"Invalid_Alpha2", "XX"},
		{"Invalid_Alpha3", "XXX"},
		{"Invalid_Numeric", "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.code.MarshalText()
			if err == nil {
				t.Errorf("MarshalText(%q) should return error for invalid code", tt.code)
			}
		})
	}
}

// TestCountryCode_JSONUnmarshal tests JSON unmarshaling
func TestCountryCode_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected CountryCode
		wantErr  bool
	}{
		{"Alpha2", `"US"`, "US", false},
		{"Alpha3", `"USA"`, "USA", false},
		{"Numeric", `"840"`, "840", false},
		{"CaseInsensitive", `"us"`, "US", false},
		{"Invalid", `"XX"`, "", true},
		{"Empty", `""`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var code CountryCode
			err := json.Unmarshal([]byte(tt.input), &code)

			if (err != nil) != tt.wantErr {
				t.Fatalf("json.Unmarshal(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && code != tt.expected {
				t.Errorf("json.Unmarshal(%s) = %q, want %q", tt.input, code, tt.expected)
			}
		})
	}
}

// TestCountryCode_JSONRoundTrip tests JSON marshal/unmarshal round-trip
func TestCountryCode_JSONRoundTrip(t *testing.T) {
	type TestStruct struct {
		Country CountryCode `json:"country"`
	}

	tests := []struct {
		name  string
		input TestStruct
	}{
		{"Alpha2", TestStruct{Country: "US"}},
		{"Alpha3", TestStruct{Country: "USA"}},
		{"Numeric", TestStruct{Country: "840"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}

			// Unmarshal
			var output TestStruct
			err = json.Unmarshal(data, &output)
			if err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}

			// Compare
			if output.Country != tt.input.Country {
				t.Errorf("Round-trip failed: got %q, want %q", output.Country, tt.input.Country)
			}

			// Validate
			if !output.Country.IsValid() {
				t.Errorf("Round-trip produced invalid code: %q", output.Country)
			}
		})
	}
}

// TestCountryCode_YAMLMarshal tests YAML marshaling
func TestCountryCode_YAMLMarshal(t *testing.T) {
	type Config struct {
		Country CountryCode `yaml:"country"`
	}

	tests := []struct {
		name     string
		input    Config
		contains string
	}{
		{"Alpha2", Config{Country: "US"}, "country: US"},
		{"Alpha3", Config{Country: "USA"}, "country: USA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.input)
			if err != nil {
				t.Fatalf("yaml.Marshal() error: %v", err)
			}

			output := string(data)
			if !containsString(output, tt.contains) {
				t.Errorf("yaml.Marshal() = %q, want to contain %q", output, tt.contains)
			}
		})
	}
}

// TestCountryCode_YAMLUnmarshal tests YAML unmarshaling
func TestCountryCode_YAMLUnmarshal(t *testing.T) {
	type Config struct {
		Country CountryCode `yaml:"country"`
	}

	tests := []struct {
		name     string
		input    string
		expected CountryCode
		wantErr  bool
	}{
		{"Alpha2", "country: US\n", "US", false},
		{"Alpha3", "country: USA\n", "USA", false},
		{"CaseInsensitive", "country: us\n", "US", false},
		{"Invalid", "country: XX\n", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			err := yaml.Unmarshal([]byte(tt.input), &config)

			if (err != nil) != tt.wantErr {
				t.Fatalf("yaml.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && config.Country != tt.expected {
				t.Errorf("yaml.Unmarshal() Country = %q, want %q", config.Country, tt.expected)
			}
		})
	}
}

// TestCountryCode_DatabaseValue tests database/sql Value interface
func TestCountryCode_DatabaseValue(t *testing.T) {
	tests := []struct {
		name     string
		code     CountryCode
		expected string
		wantErr  bool
	}{
		{"Valid_US", "US", "US", false},
		{"Valid_USA", "USA", "USA", false},
		{"Invalid", "XX", "", true},
		{"Empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.code.Value()

			if (err != nil) != tt.wantErr {
				t.Fatalf("CountryCode(%q).Value() error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}

			if !tt.wantErr {
				strValue, ok := value.(string)
				if !ok {
					t.Fatalf("CountryCode.Value() returned non-string type: %T", value)
				}
				if strValue != tt.expected {
					t.Errorf("CountryCode(%q).Value() = %q, want %q", tt.code, strValue, tt.expected)
				}
			}
		})
	}
}

// TestCountryCode_DatabaseScan tests database/sql Scanner interface
func TestCountryCode_DatabaseScan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected CountryCode
		wantErr  bool
	}{
		{"String_US", "US", "US", false},
		{"String_USA", "USA", "USA", false},
		{"Bytes_US", []byte("US"), "US", false},
		{"Bytes_USA", []byte("USA"), "USA", false},
		{"CaseInsensitive", "us", "US", false},
		{"Nil", nil, "", false},
		{"Invalid_String", "XX", "", true},
		{"Invalid_Type", 123, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var code CountryCode
			err := code.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CountryCode.Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && code != tt.expected {
				t.Errorf("CountryCode.Scan(%v) = %q, want %q", tt.input, code, tt.expected)
			}
		})
	}
}

// TestCountryCode_DatabaseRoundTrip tests Value/Scan round-trip
func TestCountryCode_DatabaseRoundTrip(t *testing.T) {
	tests := []CountryCode{"US", "USA", "840"}

	for _, original := range tests {
		t.Run(string(original), func(t *testing.T) {
			// Convert to driver.Value
			value, err := original.Value()
			if err != nil {
				t.Fatalf("Value() error: %v", err)
			}

			// Scan back
			var scanned CountryCode
			err = scanned.Scan(value)
			if err != nil {
				t.Fatalf("Scan() error: %v", err)
			}

			// Compare
			if scanned != original {
				t.Errorf("Round-trip failed: got %q, want %q", scanned, original)
			}
		})
	}
}

// TestCountryCode_IntegrationExample demonstrates real-world usage
func TestCountryCode_IntegrationExample(t *testing.T) {
	type User struct {
		Name    string      `json:"name" yaml:"name" db:"name"`
		Country CountryCode `json:"country" yaml:"country" db:"country"`
	}

	// Create user
	user := User{
		Name:    "Alice",
		Country: MustCountryCode("US"),
	}

	// JSON round-trip
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var jsonUser User
	err = json.Unmarshal(jsonData, &jsonUser)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if jsonUser.Country != user.Country {
		t.Errorf("JSON round-trip failed: got %q, want %q", jsonUser.Country, user.Country)
	}

	// YAML round-trip
	yamlData, err := yaml.Marshal(user)
	if err != nil {
		t.Fatalf("YAML marshal error: %v", err)
	}

	var yamlUser User
	err = yaml.Unmarshal(yamlData, &yamlUser)
	if err != nil {
		t.Fatalf("YAML unmarshal error: %v", err)
	}

	if yamlUser.Country != user.Country {
		t.Errorf("YAML round-trip failed: got %q, want %q", yamlUser.Country, user.Country)
	}

	// Database value/scan
	dbValue, err := user.Country.Value()
	if err != nil {
		t.Fatalf("Database Value() error: %v", err)
	}

	var dbCountry CountryCode
	err = dbCountry.Scan(dbValue)
	if err != nil {
		t.Fatalf("Database Scan() error: %v", err)
	}

	if dbCountry != user.Country {
		t.Errorf("Database round-trip failed: got %q, want %q", dbCountry, user.Country)
	}

	// Validate
	if err := user.Country.Validate(); err != nil {
		t.Errorf("Country validation failed: %v", err)
	}

	// Get full metadata
	country, err := user.Country.Country()
	if err != nil {
		t.Fatalf("Country lookup error: %v", err)
	}

	if country.Name != "United States of America" {
		t.Errorf("Country name = %q, want %q", country.Name, "United States of America")
	}
}

// Benchmarks

func BenchmarkNewCountryCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewCountryCode("US")
	}
}

func BenchmarkCountryCode_Validate(b *testing.B) {
	code := CountryCode("US")
	for i := 0; i < b.N; i++ {
		_ = code.Validate()
	}
}

func BenchmarkCountryCode_JSONMarshal(b *testing.B) {
	code := CountryCode("US")
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(code)
	}
}

func BenchmarkCountryCode_JSONUnmarshal(b *testing.B) {
	data := []byte(`"US"`)
	for i := 0; i < b.N; i++ {
		var code CountryCode
		_ = json.Unmarshal(data, &code)
	}
}

func BenchmarkCountryCode_DatabaseValue(b *testing.B) {
	code := CountryCode("US")
	for i := 0; i < b.N; i++ {
		_, _ = code.Value()
	}
}

func BenchmarkCountryCode_DatabaseScan(b *testing.B) {
	value := driver.Value("US")
	for i := 0; i < b.N; i++ {
		var code CountryCode
		_ = code.Scan(value)
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
