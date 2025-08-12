package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockDBConn struct {
	mock.Mock
}

func (m *MockDBConn) Query(ctx context.Context, q string, queryArgs ...any) (pgx.Rows, error) {
	args := m.Called(ctx, q, queryArgs)
	return args.Get(0).(pgx.Rows), args.Error(1)
}

func (m *MockDBConn) QueryRow(ctx context.Context, q string, queryArgs ...any) pgx.Row {
	return m.Called(ctx, q, queryArgs).Get(0).(pgx.Row)
}

func (m *MockDBConn) Exec(ctx context.Context, q string, queryArgs ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, q, queryArgs)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockDBConn) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockKafkaFactory struct {
	mock.Mock
}

func (m *MockKafkaFactory) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockKafkaFactory) NewWriter(topic string, opts ...bootstrap.WriterOption) *kafka.Writer {
	return m.Called(topic, opts).Get(0).(*kafka.Writer)
}

type MockContactsRepository struct {
	mock.Mock
}

func (m *MockContactsRepository) GetAllContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Contact), args.Error(1)
}

func (m *MockContactsRepository) GetContactsByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error) {
	args := m.Called(ctx, userID, limit, offset)
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

func (m *MockTemplateRepository) GetTemplatesByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Template, error) {
	args := m.Called(ctx, userID, limit, offset)
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
