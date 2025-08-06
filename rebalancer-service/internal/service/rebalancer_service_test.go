package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FetchAndUpdatePending(ctx context.Context, batchSize int) ([]*models.Notification, error) {
	args := m.Called(ctx, batchSize)
	return args.Get(0).([]*models.Notification), args.Error(1)
}

type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

func TestRebalancerService_rebalance(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name             string
		fetchResult      []*models.Notification
		fetchErr         error
		writeErr         error
		expectWriteCalls bool
	}{
		{
			name:             "no notifications",
			fetchResult:      []*models.Notification{},
			fetchErr:         nil,
			expectWriteCalls: false,
		},
		{
			name:             "fetch error",
			fetchErr:         assert.AnError,
			expectWriteCalls: false,
		},
		{
			name: "successful fetch and write",
			fetchResult: []*models.Notification{
				{
					ID:             id,
					Text:           "Hello",
					RecipientPhone: "1234567890",
					Attempts:       1,
				},
			},
			fetchErr:         nil,
			writeErr:         nil,
			expectWriteCalls: true,
		},
		{
			name: "write error",
			fetchResult: []*models.Notification{
				{
					ID:             id,
					Text:           "Fail",
					RecipientPhone: "999",
					Attempts:       2,
				},
			},
			fetchErr:         nil,
			writeErr:         assert.AnError,
			expectWriteCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockWriter := new(MockKafkaWriter)
			logger := zaptest.NewLogger(t)

			ctx := context.Background()
			rs := NewRebalancerService(mockRepo, mockWriter, logger, 10, time.Second, time.Second)

			mockRepo.
				On("FetchAndUpdatePending", mock.Anything, 10).
				Return(tt.fetchResult, tt.fetchErr).
				Once()

			if tt.expectWriteCalls {
				var expectedMsgs []kafka.Message
				for _, n := range tt.fetchResult {
					b, _ := json.Marshal(&domain.SendNotificationTask{
						ID:             n.ID,
						Text:           n.Text,
						RecipientPhone: n.RecipientPhone,
						Attempts:       n.Attempts,
					})
					expectedMsgs = append(expectedMsgs, kafka.Message{Value: b})
				}
				mockWriter.
					On("WriteMessages", mock.Anything, expectedMsgs).
					Return(tt.writeErr).
					Once()
			}

			rs.rebalance(ctx)

			mockRepo.AssertExpectations(t)
			mockWriter.AssertExpectations(t)
		})
	}
}
