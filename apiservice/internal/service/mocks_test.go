package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockContactsRepository struct {
	mock.Mock
}

func (m *MockContactsRepository) GetContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Contact), args.Error(1)
}

func (m *MockContactsRepository) GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsRepository) CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, contact)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsRepository) UpdateContact(ctx context.Context, userID, contactID int, contact *models.Contact) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID, contact)
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactsRepository) DeleteContact(ctx context.Context, userID, contactID int) error {
	return m.Called(ctx, userID, contactID).Error(0)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

type MockTemplateRepository struct {
	mock.Mock
}

func (m *MockTemplateRepository) GetTemplatesByUserID(ctx context.Context, userID int) ([]*models.Template, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Template), args.Error(1)
}

func (m *MockTemplateRepository) GetTemplateByID(ctx context.Context, userID, tmplID int) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateRepository) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateRepository) UpdateTemplate(ctx context.Context, userID, tmplID int, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID, tmpl)
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateRepository) DeleteTemplate(ctx context.Context, userID, tmplID int) error {
	return m.Called(ctx, userID, tmplID).Error(0)
}
