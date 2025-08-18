package handler_test

import (
	"context"
	"io"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockContactsService struct {
	mock.Mock
}

func (m *MockContactsService) GetContactsCountByUserID(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockContactsService) GetContactsPageByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.Contact), args.Error(1)
}

func (m *MockContactsService) GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsService) CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, contact)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsService) UpdateContact(ctx context.Context, userID, contactID int, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID, contact)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsService) DeleteContact(ctx context.Context, userID, contactID int) error {
	return m.Called(ctx, userID, contactID).Error(0)
}

type MockHealthCheckService struct {
	mock.Mock
}

func (m *MockHealthCheckService) HealthCheck(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockLoadContactsService struct {
	mock.Mock
}

func (m *MockLoadContactsService) ProcessUpload(ctx context.Context, userID int, filename string, payload io.ReadSeeker) error {
	return m.Called(ctx, userID, filename, payload).Error(0)
}

type MockLoginService struct {
	mock.Mock
}

func (m *MockLoginService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockLoginService) CompareCredentials(user *models.User, req *domain.LoginRequest) bool {
	return m.Called(user, req).Get(0).(bool)
}

type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

type MockRefreshTokenService struct {
	mock.Mock
}

func (m *MockRefreshTokenService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

type MockSendNotificationService struct {
	mock.Mock
}

func (m *MockSendNotificationService) SendNotification(ctx context.Context, userID, templateID int) error {
	return m.Called(ctx, userID, templateID).Error(0)
}

type MockSignupService struct {
	mock.Mock
}

func (m *MockSignupService) CreateUser(ctx context.Context, req *domain.SignupRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.User), args.Error(1)
}

type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) GetTemplatesCountByUserID(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockTemplateService) GetTemplatesPageByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Template, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplateByID(ctx context.Context, userID int, tmplID int) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) UpdateTemplate(ctx context.Context, userID int, tmplID int, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateService) DeleteTemplate(ctx context.Context, userID int, tmplID int) error {
	args := m.Called(ctx, userID, tmplID)
	return args.Error(0)
}
