package consumers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/domain"
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

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msg ...kafka.Message) error {
	return m.Called(ctx, msg).Error(0)
}

type MockContactsService struct {
	mock.Mock
}

func (m *MockContactsService) ProcessFile(ctx context.Context, t *domain.Task) (int, error) {
	args := m.Called(ctx, t)
	return args.Int(0), args.Error(1)
}

func TestContactsConsumer_StartConsumer(t *testing.T) {
	validTask := domain.Task{
		UserID: 42,
		S3Key:  "contacts.csv",
	}
	validJSON, _ := json.Marshal(validTask)
	validMsg := kafka.Message{Value: validJSON}
	invalidMsg := kafka.Message{Value: []byte("invalid-json")}

	tests := []struct {
		name                 string
		setupMocks           func(*MockKafkaReader, *MockContactsService)
		expectError          error
		cancelContextAfterMs int
	}{
		{
			name: "successful processing",
			setupMocks: func(mr *MockKafkaReader, ms *MockContactsService) {
				mr.
					On("FetchMessage", mock.Anything).
					Return(validMsg, nil).
					Once()
				mr.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				ms.
					On("ProcessFile", mock.Anything, &validTask).
					Return(5, nil).
					Once()
				mr.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, context.Canceled).
					Once()
			},
			expectError: context.Canceled,
		},
		{
			name: "invalid json task",
			setupMocks: func(mr *MockKafkaReader, ms *MockContactsService) {
				mr.
					On("FetchMessage", mock.Anything).
					Return(invalidMsg, nil).
					Once()
				mr.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				mr.
					On("FetchMessage", mock.Anything).
					Return(validMsg, context.Canceled).
					Once()
			},
			expectError: context.Canceled,
		},
		{
			name: "process file error",
			setupMocks: func(mr *MockKafkaReader, ms *MockContactsService) {
				mr.
					On("FetchMessage", mock.Anything).
					Return(validMsg, nil).
					Once()
				ms.
					On("ProcessFile", mock.Anything, &validTask).
					Return(0, assert.AnError).
					Once()
				mr.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(nil).
					Once()
				mr.
					On("FetchMessage", mock.Anything).
					Return(validMsg, context.Canceled).
					Once()
			},
			expectError: context.Canceled,
		},
		{
			name: "commit error",
			setupMocks: func(mr *MockKafkaReader, ms *MockContactsService) {
				mr.
					On("FetchMessage", mock.Anything).
					Return(validMsg, nil).
					Once()
				ms.
					On("ProcessFile", mock.Anything, &validTask).
					Return(5, nil).
					Once()
				mr.
					On("CommitMessages", mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			expectError: assert.AnError,
		},
		{
			name: "fetch message error",
			setupMocks: func(mr *MockKafkaReader, ms *MockContactsService) {
				mr.
					On("FetchMessage", mock.Anything).
					Return(kafka.Message{}, assert.AnError).
					Once()
			},
			expectError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaReader := &MockKafkaReader{}
			service := &MockContactsService{}
			logger := zaptest.NewLogger(t)

			tt.setupMocks(kafkaReader, service)
			consumer := consumers.NewContactsConsumer(service, kafkaReader, logger)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := consumer.StartConsumer(ctx)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
			}

			kafkaReader.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}
