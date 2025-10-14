package foundry

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// CountryCode is a validated ISO 3166-1 country code.
//
// Supports Alpha-2 (US), Alpha-3 (USA), and Numeric (840) codes with
// automatic normalization. Implements standard Go interfaces for seamless
// integration with JSON, YAML, TOML, and SQL databases.
//
// The zero value is an invalid country code. Use NewCountryCode or
// MustCountryCode to create valid instances.
//
// Example:
//
//	type User struct {
//	    Name    string      `json:"name"`
//	    Country CountryCode `json:"country" db:"country"`
//	}
//
//	user := User{Name: "Alice", Country: MustCountryCode("US")}
//	json.Marshal(user) // {"name":"Alice","country":"US"}
type CountryCode string

// NewCountryCode creates a validated CountryCode from any ISO 3166-1 format.
//
// Accepts Alpha-2 (US, us), Alpha-3 (USA, usa), or Numeric (840, 76) codes.
// Returns an error if the code is invalid.
//
// Example:
//
//	code, err := NewCountryCode("US")    // Alpha-2
//	code, err := NewCountryCode("usa")   // Alpha-3 (case-insensitive)
//	code, err := NewCountryCode("840")   // Numeric
func NewCountryCode(code string) (CountryCode, error) {
	if code == "" {
		return "", fmt.Errorf("country code cannot be empty")
	}

	// Validate using catalog
	if !ValidateCountryCode(code) {
		return "", fmt.Errorf("invalid country code: %s", code)
	}

	// Normalize to uppercase for consistency
	return CountryCode(strings.ToUpper(code)), nil
}

// MustCountryCode creates a CountryCode or panics if invalid.
//
// Use this for compile-time constants or when the code is known to be valid.
//
// Example:
//
//	const DefaultCountry = MustCountryCode("US")
func MustCountryCode(code string) CountryCode {
	c, err := NewCountryCode(code)
	if err != nil {
		panic(err)
	}
	return c
}

// String returns the country code as a string.
func (c CountryCode) String() string {
	return string(c)
}

// Validate checks if the country code is valid.
//
// Returns an error if the code is not a recognized ISO 3166-1 code.
func (c CountryCode) Validate() error {
	if c == "" {
		return fmt.Errorf("country code is empty")
	}

	if !ValidateCountryCode(string(c)) {
		return fmt.Errorf("invalid country code: %s", c)
	}

	return nil
}

// IsValid returns true if the country code is valid.
func (c CountryCode) IsValid() bool {
	return c.Validate() == nil
}

// Country retrieves the full Country metadata from the catalog.
//
// Returns an error if the code is invalid or the catalog cannot be loaded.
//
// Example:
//
//	code := MustCountryCode("US")
//	country, err := code.Country()
//	if err == nil {
//	    fmt.Println(country.Name) // "United States of America"
//	}
func (c CountryCode) Country() (*Country, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	codeStr := string(c)

	// Try Alpha-2 lookup first
	country, err := GetCountry(codeStr)
	if err != nil {
		return nil, err
	}
	if country != nil {
		return country, nil
	}

	// Try Alpha-3 lookup
	country, err = GetCountryByAlpha3(codeStr)
	if err != nil {
		return nil, err
	}
	if country != nil {
		return country, nil
	}

	return nil, fmt.Errorf("country not found for code: %s", c)
}

// MarshalText implements encoding.TextMarshaler for JSON, YAML, TOML support.
//
// The country code is marshaled as-is (uppercase normalized).
func (c CountryCode) MarshalText() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return []byte(c), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON, YAML, TOML support.
//
// Validates and normalizes the country code on unmarshal.
// Accepts Alpha-2, Alpha-3, or Numeric codes in any case.
func (c *CountryCode) UnmarshalText(text []byte) error {
	code, err := NewCountryCode(string(text))
	if err != nil {
		return err
	}
	*c = code
	return nil
}

// Value implements database/sql/driver.Valuer for database integration.
//
// The country code is stored as a string (VARCHAR/TEXT column).
func (c CountryCode) Value() (driver.Value, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return string(c), nil
}

// Scan implements database/sql.Scanner for database integration.
//
// Reads country codes from VARCHAR/TEXT columns with validation.
func (c *CountryCode) Scan(src interface{}) error {
	if src == nil {
		*c = ""
		return nil
	}

	var code string
	switch v := src.(type) {
	case string:
		code = v
	case []byte:
		code = string(v)
	default:
		return fmt.Errorf("cannot scan %T into CountryCode", src)
	}

	parsed, err := NewCountryCode(code)
	if err != nil {
		return err
	}

	*c = parsed
	return nil
}
