package foundry

import "strings"

// Country represents an ISO 3166 country code from the Foundry catalog.
//
// Countries provide standardized codes for country identification across
// services. These are loaded from Crucible configuration.
type Country struct {
	// Alpha2 is the ISO 3166-1 alpha-2 two-letter country code (e.g., "US", "CA").
	Alpha2 string

	// Alpha3 is the ISO 3166-1 alpha-3 three-letter country code (e.g., "USA", "CAN").
	Alpha3 string

	// Numeric is the ISO 3166-1 numeric country code as a string (e.g., "840", "124").
	Numeric string

	// Name is the common English name of the country (e.g., "United States of America").
	Name string

	// OfficialName is the official name of the country (e.g., "United States of America").
	OfficialName string
}

// MatchesCode checks if the given code matches this country's Alpha2 or Alpha3 code.
//
// Matching is case-insensitive.
//
// Example:
//
//	country := &Country{Alpha2: "US", Alpha3: "USA"}
//	if country.MatchesCode("us") {  // true
//	    // Matched
//	}
//	if country.MatchesCode("USA") { // also true
//	    // Matched
//	}
func (c *Country) MatchesCode(code string) bool {
	upperCode := strings.ToUpper(code)
	return strings.ToUpper(c.Alpha2) == upperCode || strings.ToUpper(c.Alpha3) == upperCode
}

// GetCountry retrieves a country by its Alpha2 code from the default catalog.
//
// Returns nil if the country is not found or if an error occurs.
//
// Example:
//
//	country, err := GetCountry("US")
//	if err != nil {
//	    // Handle error
//	}
//	if country != nil {
//	    fmt.Println(country.Name) // "United States of America"
//	}
func GetCountry(alpha2 string) (*Country, error) {
	catalog := GetDefaultCatalog()
	return catalog.GetCountry(alpha2)
}

// GetCountryByAlpha3 retrieves a country by its Alpha3 code from the default catalog.
//
// Returns nil if the country is not found or if an error occurs.
//
// Example:
//
//	country, err := GetCountryByAlpha3("USA")
//	if err != nil {
//	    // Handle error
//	}
//	if country != nil {
//	    fmt.Println(country.Name) // "United States of America"
//	}
func GetCountryByAlpha3(alpha3 string) (*Country, error) {
	catalog := GetDefaultCatalog()
	return catalog.GetCountryByAlpha3(alpha3)
}

// GetCountryByNumeric retrieves a country by its numeric ISO 3166-1 code from the default catalog.
//
// The code is normalized to a zero-padded 3-digit string for consistent lookup.
// Accepts numeric codes with or without leading zeros.
// Returns nil if the country is not found or if an error occurs.
//
// Example:
//
//	country, err := GetCountryByNumeric("840")  // United States
//	country, err := GetCountryByNumeric("76")   // Brazil (normalized to "076")
//	if err != nil {
//	    // Handle error
//	}
//	if country != nil {
//	    fmt.Println(country.Name) // "Brazil" or "United States of America"
//	}
func GetCountryByNumeric(numeric string) (*Country, error) {
	catalog := GetDefaultCatalog()
	return catalog.GetCountryByNumeric(numeric)
}

// ValidateCountryCode checks if the given code (Alpha2, Alpha3, or Numeric) is valid.
//
// The code is normalized (uppercase for alpha codes, zero-padded for numeric)
// for case-insensitive lookup. Supports all three ISO 3166-1 formats.
//
// Returns true if the code matches a country in the catalog.
//
// Example:
//
//	if ValidateCountryCode("US") {      // Alpha2
//	    // Valid country code
//	}
//	if ValidateCountryCode("usa") {     // Alpha3 (case-insensitive)
//	    // Valid country code
//	}
//	if ValidateCountryCode("840") {     // Numeric
//	    // Valid country code
//	}
func ValidateCountryCode(code string) bool {
	if code == "" {
		return false
	}

	catalog := GetDefaultCatalog()

	// Try Alpha2 lookup (normalized to uppercase)
	country, _ := catalog.GetCountry(code)
	if country != nil {
		return true
	}

	// Try Alpha3 lookup (normalized to uppercase)
	country, _ = catalog.GetCountryByAlpha3(code)
	if country != nil {
		return true
	}

	// Try Numeric lookup (zero-padded to 3 digits)
	country, _ = catalog.GetCountryByNumeric(code)
	return country != nil
}

// ListCountries returns all countries from the default catalog.
//
// Returns a slice of Country instances.
//
// Example:
//
//	countries, err := ListCountries()
//	if err != nil {
//	    // Handle error
//	}
//	for _, country := range countries {
//	    fmt.Printf("%s: %s\n", country.Alpha2, country.Name)
//	}
func ListCountries() ([]*Country, error) {
	catalog := GetDefaultCatalog()
	return catalog.ListCountries()
}
