package consumers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

type MockNotificationRequestsService struct {
	mock.Mock
}

func (m *MockNotificationRequestsService) SaveNotifications(ctx context.Context, ntfs *[]*models.Notification) error {
	return m.Called(ctx, ntfs).Error(0)
}

func TestNotificationRequestsConsumer_StartConsumer(t *testing.T) {
	t.Run("valid notification triggers flush on batch size", func(t *testing.T) {
		userID := 1
		mockKR := new(MockKafkaReader)
		mockSvc := new(MockNotificationRequestsService)
		logger := zaptest.NewLogger(t)

		nr := domain.NotificationRequest{
			UserID:   userID,
			Template: "Hello",
			Contacts: []*models.SlimContact{
				{Phone: "123", Name: "Alice"},
				{Phone: "456", Name: "Ben"},
			},
		}
		raw, _ := json.Marshal(nr)
		msg := kafka.Message{Value: raw}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		mockKR.
			On("FetchMessage", mock.Anything).
			Return(msg, nil).
			Twice()
		mockKR.
			On("CommitMessages", mock.Anything, mock.Anything).
			Return(nil).
			Twice()
		mockSvc.
			On("SaveNotifications", mock.Anything, mock.AnythingOfType("*[]*models.Notification")).
			Return(nil).
			Once()
		mockKR.
			On("FetchMessage", mock.Anything).
			Return(kafka.Message{}, context.Canceled).
			Once()

		c := consumers.NewNotificationRequestsConsumer(mockSvc, mockKR, logger, time.Second, 4, 500*time.Millisecond)

		go func() {
			_ = c.StartConsumer(ctx)
		}()

		<-ctx.Done()
		mockKR.AssertExpectations(t)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid JSON should be skipped", func(t *testing.T) {
		mockKR := new(MockKafkaReader)
		mockSvc := new(MockNotificationRequestsService)
		logger := zaptest.NewLogger(t)

		badMsg := kafka.Message{Value: []byte("invalid-json")}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		mockKR.
			On("FetchMessage", mock.Anything).
			Return(badMsg, nil).
			Once()
		mockKR.
			On("CommitMessages", mock.Anything, mock.Anything).
			Return(nil).
			Once()
		mockKR.
			On("FetchMessage", mock.Anything).
			Return(kafka.Message{}, context.Canceled).
			Once()

		c := consumers.NewNotificationRequestsConsumer(mockSvc, mockKR, logger, time.Second, 2, 500*time.Millisecond)

		go func() {
			_ = c.StartConsumer(ctx)
		}()

		<-ctx.Done()
		mockKR.AssertExpectations(t)
		mockSvc.AssertExpectations(t)
	})

	t.Run("fetch error should exit consumer", func(t *testing.T) {
		mockKR := new(MockKafkaReader)
		mockSvc := new(MockNotificationRequestsService)
		logger := zaptest.NewLogger(t)

		ctx := context.Background()

		mockKR.On("FetchMessage", ctx).Return(kafka.Message{}, errors.New("fetch failed")).Once()

		c := consumers.NewNotificationRequestsConsumer(mockSvc, mockKR, logger, time.Second, 2, time.Second)

		err := c.StartConsumer(ctx)
		assert.EqualError(t, err, "fetch failed")

		mockKR.AssertExpectations(t)
	})
}
