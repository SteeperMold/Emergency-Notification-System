package phoneutils

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

// ISO3166Alpha2 represents a 2-letter ISO 3166-1 country code used as the default region
// when parsing phone numbers without an explicit country prefix.
type ISO3166Alpha2 string

const (
	// RegionRU specifies Russia as the default region for phone number parsing
	RegionRU ISO3166Alpha2 = "RU"
)

// FormatToE164 parses a raw phone number string using the provided default region,
// validates it, and returns the number formatted in E.164 standard (e.g. +1234567890).
// Returns the formatted E.164 string on success, or an error if parsing or validation fails.
func FormatToE164(rawPhoneNumber string, defaultRegion ISO3166Alpha2) (string, error) {
	num, err := phonenumbers.Parse(rawPhoneNumber, string(defaultRegion))
	if err != nil {
		return "", fmt.Errorf("invalid phone number")
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}
