package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIngestAndSave(t *testing.T) {
	type setupFn func(m *MockContactsRepository)
	providerFromRows := func(rows [][]string, errToReturn error) rowProvider {
		return func(ch chan<- []string) error {
			for _, r := range rows {
				ch <- r
			}
			return errToReturn
		}
	}

	tests := []struct {
		name        string
		batchSize   int
		rows        [][]string
		providerErr error   // error returned by provider
		setupMock   setupFn // expectations on SaveContacts
		wantTotal   int
		wantErr     bool
	}{
		{
			name:      "happy path, exact batch",
			batchSize: 2,
			rows: [][]string{
				{"Alice", "+79123456789"},
				{"Bob", "+79123456788"},
			},
			providerErr: nil,
			setupMock: func(m *MockContactsRepository) {
				// expect save of two contacts
				// we’ll get two batches, one per worker that got a row
				m.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(nil).
					Twice()
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "mix valid+invalid, smaller final batch",
			batchSize: 2,
			rows: [][]string{
				{"Alice", "+79123456789"},
				{"", "+79123456789"}, // invalid name
				{"Charlie", "+79120001111"},
			},
			providerErr: nil,
			setupMock: func(m *MockContactsRepository) {
				m.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "repository error on first batch",
			batchSize: 1,
			rows: [][]string{
				{"Dave", "+79125551234"}, // batch size=1 → one batch
			},
			providerErr: nil,
			setupMock: func(m *MockContactsRepository) {
				m.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantTotal: 0, // nothing saved
			wantErr:   true,
		},
		{
			name:        "provider returns error",
			batchSize:   3,
			rows:        [][]string{},
			providerErr: assert.AnError,
			setupMock:   func(m *MockContactsRepository) {},
			wantTotal:   0,
			wantErr:     true,
		},
		{
			name:      "all invalid rows",
			batchSize: 2,
			rows: [][]string{
				{"", "+79120000000"},   // invalid name
				{"Eve", "not-a-phone"}, // invalid phone
			},
			providerErr: nil,
			setupMock:   func(m *MockContactsRepository) {},
			wantTotal:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(MockContactsRepository)
			tt.setupMock(m)

			svc := &ContactsService{
				repository: m,
				batchSize:  tt.batchSize,
			}

			gotTotal, err := svc.ingestAndSave(context.Background(), 42, providerFromRows(tt.rows, tt.providerErr))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantTotal, gotTotal)

			m.AssertExpectations(t)
		})
	}
}

func TestValidateName(t *testing.T) {
	svc := &ContactsService{}
	tests := []struct {
		name string
		ok   bool
	}{
		{"John", true},
		{"", false},
		{strings.Repeat("A", 33), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := svc.validateName(tt.name)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
