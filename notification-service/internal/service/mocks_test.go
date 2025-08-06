package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

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
