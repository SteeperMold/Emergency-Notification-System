package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestContactsService_GetContactsByUserID(t *testing.T) {
	contacts := []*models.Contact{
		{ID: 1, UserID: 123, Name: "Alice", Phone: "+79123456789"},
		{ID: 2, UserID: 123, Name: "Bob", Phone: "+79129876543"},
	}

	m := new(MockContactsRepository)
	m.
		On("GetContactsByUserID", mock.Anything, 123).
		Return(contacts, nil).
		Once()
	svc := service.NewContactsService(m)

	res, err := svc.GetContactsByUserID(context.Background(), 123)
	assert.NoError(t, err)
	assert.Equal(t, contacts, res)
	m.AssertExpectations(t)
}

func TestContactsService_GetContactByID(t *testing.T) {
	contact := &models.Contact{ID: 456, UserID: 123, Name: "Alice", Phone: "+79123456789"}

	m := new(MockContactsRepository)
	m.
		On("GetContactByID", mock.Anything, 123, 456).
		Return(contact, nil).
		Once()
	svc := service.NewContactsService(m)

	res, err := svc.GetContactByID(context.Background(), 123, 456)
	assert.NoError(t, err)
	assert.Equal(t, contact, res)
	m.AssertExpectations(t)
}

func TestContactsService_CreateContact(t *testing.T) {
	type args struct {
		ctx     context.Context
		contact *models.Contact
	}
	tests := []struct {
		name       string
		mockSetup  func(*MockContactsRepository)
		args       args
		wantResult *models.Contact
		wantErr    error
	}{
		{
			name: "success",
			mockSetup: func(m *MockContactsRepository) {
				m.
					On("CreateContact", mock.Anything, &models.Contact{UserID: 123, Name: "Alice", Phone: "+79123456789"}).
					Return(&models.Contact{ID: 1, UserID: 123, Name: "Alice", Phone: "+79123456789"}, nil).
					Once()
			},
			args: args{
				context.Background(),
				&models.Contact{UserID: 123, Name: "Alice", Phone: "+79123456789"},
			},
			wantResult: &models.Contact{ID: 1, UserID: 123, Name: "Alice", Phone: "+79123456789"},
			wantErr:    nil,
		},
		{
			name:      "name too short",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				context.Background(),
				&models.Contact{ID: 1, UserID: 123, Name: "", Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:      "name too long",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				context.Background(),
				&models.Contact{ID: 1, UserID: 123, Name: strings.Repeat("A", 33), Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:      "invalid phone",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				ctx:     context.Background(),
				contact: &models.Contact{UserID: 123, Name: "Alice", Phone: "not-a-phone"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactPhone,
		},
		{
			name: "repository error",
			mockSetup: func(m *MockContactsRepository) {
				m.
					On("CreateContact", mock.Anything, &models.Contact{UserID: 123, Name: "Alice", Phone: "+79123456789"}).
					Return((*models.Contact)(nil), assert.AnError).
					Once()
			},
			args: args{
				ctx:     context.Background(),
				contact: &models.Contact{UserID: 123, Name: "Alice", Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    assert.AnError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsRepository)
			tc.mockSetup(m)
			svc := service.NewContactsService(m)

			res, err := svc.CreateContact(tc.args.ctx, tc.args.contact)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantResult, res)
			m.AssertExpectations(t)
		})
	}
}

func TestContactsService_UpdateContact(t *testing.T) {
	type args struct {
		ctx            context.Context
		userID, cid    int
		updatedContact *models.Contact
	}

	tests := []struct {
		name       string
		mockSetup  func(*MockContactsRepository)
		args       args
		wantResult *models.Contact
		wantErr    error
	}{
		{
			name: "success",
			mockSetup: func(m *MockContactsRepository) {
				normalized := "+79123456789"
				input := &models.Contact{UserID: 123, Name: "Alice", Phone: normalized}
				output := &models.Contact{ID: 42, UserID: 123, Name: "Alice", Phone: normalized}
				m.
					On("UpdateContact", mock.Anything, 123, 42, input).
					Return(output, nil).
					Once()
			},
			args: args{
				ctx:            context.Background(),
				userID:         123,
				cid:            42,
				updatedContact: &models.Contact{UserID: 123, Name: "Alice", Phone: "8 (912) 345-6789"},
			},
			wantResult: &models.Contact{ID: 42, UserID: 123, Name: "Alice", Phone: "+79123456789"},
			wantErr:    nil,
		},
		{
			name:      "name too short",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				ctx:            context.Background(),
				userID:         123,
				cid:            42,
				updatedContact: &models.Contact{UserID: 123, Name: "", Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:      "name too long",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				ctx:            context.Background(),
				userID:         123,
				cid:            42,
				updatedContact: &models.Contact{UserID: 123, Name: strings.Repeat("A", 33), Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:      "invalid phone",
			mockSetup: func(m *MockContactsRepository) {},
			args: args{
				ctx:            context.Background(),
				userID:         123,
				cid:            42,
				updatedContact: &models.Contact{UserID: 123, Name: "Alice", Phone: "not-a-number"},
			},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactPhone,
		},
		{
			name: "repository error",
			mockSetup: func(m *MockContactsRepository) {
				normalized := "+79123456789"
				input := &models.Contact{UserID: 123, Name: "Alice", Phone: normalized}

				m.
					On("UpdateContact", mock.Anything, 123, 42, input).
					Return((*models.Contact)(nil), assert.AnError).
					Once()
			},
			args: args{
				ctx:            context.Background(),
				userID:         123,
				cid:            42,
				updatedContact: &models.Contact{UserID: 123, Name: "Alice", Phone: "+79123456789"},
			},
			wantResult: nil,
			wantErr:    assert.AnError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsRepository)
			tc.mockSetup(m)
			svc := service.NewContactsService(m)

			res, err := svc.UpdateContact(tc.args.ctx, tc.args.userID, tc.args.cid, tc.args.updatedContact)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, res)
			}

			m.AssertExpectations(t)
		})
	}
}

func TestContactsService_DeleteContact(t *testing.T) {
	m := new(MockContactsRepository)
	m.
		On("DeleteContact", mock.Anything, 123, 42).
		Return(nil).
		Once()
	svc := service.NewContactsService(m)

	err := svc.DeleteContact(context.Background(), 123, 42)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
