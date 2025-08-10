package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
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

func (m *MockDBConn) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
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

func (m *MockKafkaFactory) NewReader(topic string, groupID string) *kafka.Reader {
	return m.Called(topic, groupID).Get(0).(*kafka.Reader)
}

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) CreateMultipleNotifications(ctx context.Context, notifications []*models.Notification) error {
	return m.Called(ctx, notifications).Error(0)
}

func (m *MockNotificationRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) ChangeNotificationStatus(ctx context.Context, id uuid.UUID, newStatus models.NotificationStatus) error {
	return m.Called(ctx, id, newStatus).Error(0)
}

type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}
