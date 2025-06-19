package phoneutils

import (
	"fmt"
	"github.com/nyaruka/phonenumbers"
)

type ISO3166Alpha2 string

const (
	RegionRU ISO3166Alpha2 = "RU"
)

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
