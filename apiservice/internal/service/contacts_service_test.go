package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/phoneutils"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContactsRepo is a testify/mock implementation of domain.ContactsRepository.
type MockContactsRepo struct {
	mock.Mock
}

func (m *MockContactsRepo) GetContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error) {
	args := m.Called(ctx, userID)
	if cs := args.Get(0); cs != nil {
		return cs.([]*models.Contact), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContactsRepo) GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID)
	if c := args.Get(0); c != nil {
		return c.(*models.Contact), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContactsRepo) CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, contact)
	if c := args.Get(0); c != nil {
		return c.(*models.Contact), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContactsRepo) UpdateContact(ctx context.Context, userID, contactID int, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID, contact)
	if c := args.Get(0); c != nil {
		return c.(*models.Contact), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContactsRepo) DeleteContact(ctx context.Context, userID, contactID int) error {
	return m.Called(ctx, userID, contactID).Error(0)
}

func TestContactsService(t *testing.T) {
	// a sample valid E.164 number
	validRaw := "8 (912) 345-6789"
	validE164, _ := phoneutils.FormatToE164(validRaw, phoneutils.RegionRU)

	errDB := errors.New("db failure")

	tests := []struct {
		name       string
		method     string
		args       []interface{}
		mockSetup  func(*MockContactsRepo)
		wantResult interface{}
		wantErr    error
	}{
		{
			name:   "GetContactsByUserID success",
			method: "GetContactsByUserID",
			args:   []interface{}{context.Background(), 42},
			mockSetup: func(m *MockContactsRepo) {
				contacts := []*models.Contact{
					{ID: 1, UserID: 42, Name: "A", Phone: "+79123456789"},
				}
				m.On("GetContactsByUserID", mock.Anything, 42).
					Return(contacts, nil).Once()
			},
			wantResult: []*models.Contact{{ID: 1, UserID: 42, Name: "A", Phone: "+79123456789"}},
			wantErr:    nil,
		},
		{
			name:   "GetContactsByUserID error",
			method: "GetContactsByUserID",
			args:   []interface{}{context.Background(), 42},
			mockSetup: func(m *MockContactsRepo) {
				m.On("GetContactsByUserID", mock.Anything, 42).
					Return(nil, errDB).Once()
			},
			wantResult: nil,
			wantErr:    errDB,
		},
		{
			name:   "GetContactByID success",
			method: "GetContactByID",
			args:   []interface{}{context.Background(), 42, 7},
			mockSetup: func(m *MockContactsRepo) {
				contact := &models.Contact{ID: 7, UserID: 42, Name: "Bob", Phone: "+71234567890"}
				m.On("GetContactByID", mock.Anything, 42, 7).
					Return(contact, nil).Once()
			},
			wantResult: &models.Contact{ID: 7, UserID: 42, Name: "Bob", Phone: "+71234567890"},
			wantErr:    nil,
		},
		{
			name:   "GetContactByID error",
			method: "GetContactByID",
			args:   []interface{}{context.Background(), 42, 7},
			mockSetup: func(m *MockContactsRepo) {
				m.On("GetContactByID", mock.Anything, 42, 7).
					Return(nil, errDB).Once()
			},
			wantResult: nil,
			wantErr:    errDB,
		},
		{
			name:       "CreateContact invalid name",
			method:     "CreateContact",
			args:       []interface{}{context.Background(), &models.Contact{UserID: 1, Name: "", Phone: validRaw}},
			mockSetup:  func(m *MockContactsRepo) {},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:       "CreateContact invalid phone",
			method:     "CreateContact",
			args:       []interface{}{context.Background(), &models.Contact{UserID: 1, Name: "Alice", Phone: "bad"}},
			mockSetup:  func(m *MockContactsRepo) {},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactPhone,
		},
		{
			name:   "CreateContact success",
			method: "CreateContact",
			args:   []interface{}{context.Background(), &models.Contact{UserID: 1, Name: "Alice", Phone: validRaw}},
			mockSetup: func(m *MockContactsRepo) {
				m.On("CreateContact", mock.Anything, mock.MatchedBy(func(c *models.Contact) bool {
					return c.UserID == 1 && c.Name == "Alice" && c.Phone == validE164
				})).Return(&models.Contact{ID: 5, UserID: 1, Name: "Alice", Phone: validE164}, nil).Once()
			},
			wantResult: &models.Contact{ID: 5, UserID: 1, Name: "Alice", Phone: validE164},
			wantErr:    nil,
		},
		{
			name:   "CreateContact repo error",
			method: "CreateContact",
			args:   []interface{}{context.Background(), &models.Contact{UserID: 2, Name: "Joe", Phone: validRaw}},
			mockSetup: func(m *MockContactsRepo) {
				m.On("CreateContact", mock.Anything, mock.Anything).
					Return(nil, errDB).Once()
			},
			wantResult: nil,
			wantErr:    errDB,
		},
		{
			name:       "UpdateContact invalid name",
			method:     "UpdateContact",
			args:       []interface{}{context.Background(), 7, 3, &models.Contact{Name: "", Phone: validRaw}},
			mockSetup:  func(m *MockContactsRepo) {},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactName,
		},
		{
			name:       "UpdateContact invalid phone",
			method:     "UpdateContact",
			args:       []interface{}{context.Background(), 7, 3, &models.Contact{Name: "Joe", Phone: "123"}},
			mockSetup:  func(m *MockContactsRepo) {},
			wantResult: nil,
			wantErr:    domain.ErrInvalidContactPhone,
		},
		{
			name:   "UpdateContact success",
			method: "UpdateContact",
			args:   []interface{}{context.Background(), 7, 3, &models.Contact{Name: "Joe", Phone: validRaw}},
			mockSetup: func(m *MockContactsRepo) {
				m.On("UpdateContact", mock.Anything, 7, 3, mock.MatchedBy(func(c *models.Contact) bool {
					return c.Name == "Joe" && c.Phone == validE164
				})).Return(&models.Contact{ID: 3, UserID: 7, Name: "Joe", Phone: validE164}, nil).Once()
			},
			wantResult: &models.Contact{ID: 3, UserID: 7, Name: "Joe", Phone: validE164},
			wantErr:    nil,
		},
		{
			name:   "UpdateContact repo error",
			method: "UpdateContact",
			args:   []interface{}{context.Background(), 7, 3, &models.Contact{Name: "Sam", Phone: validRaw}},
			mockSetup: func(m *MockContactsRepo) {
				m.On("UpdateContact", mock.Anything, 7, 3, mock.Anything).
					Return(nil, errDB).Once()
			},
			wantResult: nil,
			wantErr:    errDB,
		},
		{
			name:   "DeleteContact success",
			method: "DeleteContact",
			args:   []interface{}{context.Background(), 9, 4},
			mockSetup: func(m *MockContactsRepo) {
				m.On("DeleteContact", mock.Anything, 9, 4).Return(nil).Once()
			},
			wantResult: nil,
			wantErr:    nil,
		},
		{
			name:   "DeleteContact error",
			method: "DeleteContact",
			args:   []interface{}{context.Background(), 9, 4},
			mockSetup: func(m *MockContactsRepo) {
				m.On("DeleteContact", mock.Anything, 9, 4).Return(errDB).Once()
			},
			wantResult: nil,
			wantErr:    errDB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockContactsRepo)
			svc := service.NewContactsService(mockRepo)
			tc.mockSetup(mockRepo)

			var (
				res interface{}
				err error
			)
			ctx := tc.args[0].(context.Context)

			switch tc.method {
			case "GetContactsByUserID":
				res, err = svc.GetContactsByUserID(ctx, tc.args[1].(int))
			case "GetContactByID":
				res, err = svc.GetContactByID(ctx, tc.args[1].(int), tc.args[2].(int))
			case "CreateContact":
				res, err = svc.CreateContact(ctx, tc.args[1].(*models.Contact))
			case "UpdateContact":
				res, err = svc.UpdateContact(ctx, tc.args[1].(int), tc.args[2].(int), tc.args[3].(*models.Contact))
			case "DeleteContact":
				err = svc.DeleteContact(ctx, tc.args[1].(int), tc.args[2].(int))
			}

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantResult != nil {
				assert.Equal(t, tc.wantResult, res)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
