package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/models"
	"github.com/aws/aws-sdk-go/aws"
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

func (m *MockDBConn) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
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

func (m *MockKafkaFactory) NewReader(topic string, groupID string) *kafka.Reader {
	return m.Called(topic, groupID).Get(0).(*kafka.Reader)
}

type MockContactsRepository struct {
	mock.Mock
}

func (m *MockContactsRepository) SaveContacts(ctx context.Context, contacts []*models.Contact) error {
	return m.Called(ctx, contacts).Error(0)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}
