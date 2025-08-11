package consumers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, task *domain.NotificationTask) error {
	return m.Called(ctx, task).Error(0)
}

type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msg ...kafka.Message) error {
	return m.Called(ctx, msg).Error(0)
}

// retryableErr implements domain.SendError and returns Retryable=true
type retryableErr struct {
	msg string
}

func (e retryableErr) Error() string {
	return e.msg
}
func (e retryableErr) Retryable() bool {
	return true
}

// permanentErr implements domain.SendError and returns Retryable=false
type permanentErr struct {
	msg string
}

func (e permanentErr) Error() string {
	return e.msg
}
func (e permanentErr) Retryable() bool {
	return false
}

func buildMsg(id uuid.UUID) kafka.Message {
	task := domain.NotificationTask{ID: id, RecipientPhone: "+123", Text: "hello"}
	b, _ := json.Marshal(task)
	return kafka.Message{Value: b}
}

func TestStartConsumer(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	cases := []struct {
		name      string
		setup     func(s *MockNotificationService, r *MockKafkaReader)
		expectErr bool
	}{
		{
			name: "fetch error",
			setup: func(_ *MockNotificationService, r *MockKafkaReader) {
				r.On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, errors.New("kafka down"))
			},
			expectErr: true,
		},
		{
			name: "invalid JSON",
			setup: func(_ *MockNotificationService, r *MockKafkaReader) {
				bad := kafka.Message{Value: []byte("not json")}
				r.
					On("FetchMessage", mock.Anything).
					Return(bad, nil).
					Once()
				r.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				r.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, context.Canceled).
					Once()
			},
			expectErr: true,
		},
		{
			name: "retryable send error",
			setup: func(s *MockNotificationService, r *MockKafkaReader) {
				msg := buildMsg(id)
				r.
					On("FetchMessage", mock.Anything).
					Return(msg, nil).
					Once()
				s.
					On("SendNotification", mock.Anything, &domain.NotificationTask{
						ID:             id,
						RecipientPhone: "+123",
						Text:           "hello",
					}).
					Return(retryableErr{"temp fail"}).
					Once()
				r.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, errors.New("stop")).
					Once()
			},
			expectErr: true,
		},
		{
			name: "permanent send error",
			setup: func(s *MockNotificationService, r *MockKafkaReader) {
				msg := buildMsg(id)
				r.
					On("FetchMessage", mock.Anything).
					Return(msg, nil).
					Once()
				s.
					On("SendNotification", mock.Anything, &domain.NotificationTask{
						ID:             id,
						RecipientPhone: "+123",
						Text:           "hello",
					}).
					Return(permanentErr{"perm fail"}).
					Once()
				r.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				r.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, context.Canceled).
					Once()
			},
			expectErr: true,
		},
		{
			name: "successful send",
			setup: func(s *MockNotificationService, r *MockKafkaReader) {
				msg := buildMsg(id)
				r.
					On("FetchMessage", mock.Anything).
					Return(msg, nil).
					Once()
				s.
					On("SendNotification", mock.Anything, &domain.NotificationTask{
						ID:             id,
						RecipientPhone: "+123",
						Text:           "hello",
					}).
					Return(nil).
					Once()
				r.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				r.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, context.Canceled).
					Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := new(MockNotificationService)
			rdr := new(MockKafkaReader)
			logger := zap.NewNop()

			tc.setup(svc, rdr)

			consumer := consumers.NewNotificationTasksConsumer(svc, rdr, logger, 100*time.Millisecond)
			err := consumer.StartConsumer(ctx)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			svc.AssertExpectations(t)
			rdr.AssertExpectations(t)
		})
	}
}
