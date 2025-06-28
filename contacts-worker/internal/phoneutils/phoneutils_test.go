package phoneutils_test

import (
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/phoneutils"
	"github.com/stretchr/testify/assert"
)

func TestFormatToE164(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		region    phoneutils.ISO3166Alpha2
		want      string
		expectErr bool
	}{
		{
			name:      "Valid Russian 10‑digit starting with 8",
			raw:       "8 (912) 345-6789",
			region:    phoneutils.RegionRU,
			want:      "+79123456789",
			expectErr: false,
		},
		{
			name:      "Valid Russian with country code",
			raw:       "+7 912 345 6789",
			region:    phoneutils.RegionRU,
			want:      "+79123456789",
			expectErr: false,
		},
		{
			name:      "Invalid too short",
			raw:       "12345",
			region:    phoneutils.RegionRU,
			expectErr: true,
		},
		{
			name:      "Malformed characters",
			raw:       "phone123",
			region:    phoneutils.RegionRU,
			expectErr: true,
		},
		{
			name:      "Valid format but invalid number",
			raw:       "+7 000 000 0000",
			region:    phoneutils.RegionRU,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := phoneutils.FormatToE164(tc.raw, tc.region)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
