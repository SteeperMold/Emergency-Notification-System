package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/service"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveNotifications(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	baseNtf := &models.Notification{
		ID:             id,
		Text:           "Test message",
		RecipientPhone: "+1234567890",
	}

	tests := []struct {
		name          string
		notifications []*models.Notification
		batchSize     int
		setupMocks    func(r *MockNotificationRepository, w *MockKafkaWriter)
		expectErr     bool
	}{
		{
			name:          "success - single batch",
			notifications: []*models.Notification{baseNtf, baseNtf},
			batchSize:     5,
			setupMocks: func(r *MockNotificationRepository, w *MockKafkaWriter) {
				r.
					On("CreateMultipleNotifications", mock.Anything, mock.Anything).
					Return(nil).
					Once()

				var msgs []kafka.Message
				for _, n := range []*models.Notification{baseNtf, baseNtf} {
					taskBytes, _ := json.Marshal(&domain.SendNotificationTask{
						ID:             n.ID,
						Text:           n.Text,
						RecipientPhone: n.RecipientPhone,
						Attempts:       1,
					})
					msgs = append(msgs, kafka.Message{Value: taskBytes})
				}

				w.
					On("WriteMessages", mock.Anything, msgs).
					Return(nil).
					Once()
			},
			expectErr: false,
		},
		{
			name:          "repository failure",
			notifications: []*models.Notification{baseNtf},
			batchSize:     2,
			setupMocks: func(r *MockNotificationRepository, w *MockKafkaWriter) {
				r.
					On("CreateMultipleNotifications", mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			expectErr: true,
		},
		{
			name:          "kafka write error",
			notifications: []*models.Notification{baseNtf, baseNtf, baseNtf},
			batchSize:     2,
			setupMocks: func(r *MockNotificationRepository, w *MockKafkaWriter) {
				r.
					On("CreateMultipleNotifications", mock.Anything, mock.Anything).
					Return(nil).
					Once()

				var firstBatch []kafka.Message
				for _, n := range []*models.Notification{baseNtf, baseNtf} {
					taskBytes, _ := json.Marshal(&domain.SendNotificationTask{
						ID:             n.ID,
						Text:           n.Text,
						RecipientPhone: n.RecipientPhone,
						Attempts:       1,
					})
					firstBatch = append(firstBatch, kafka.Message{Value: taskBytes})
				}
				w.
					On("WriteMessages", mock.Anything, firstBatch).
					Return(nil).
					Once()

				third := baseNtf
				taskBytes, _ := json.Marshal(&domain.SendNotificationTask{
					ID:             third.ID,
					Text:           third.Text,
					RecipientPhone: third.RecipientPhone,
					Attempts:       1,
				})
				lastBatch := []kafka.Message{{Value: taskBytes}}

				w.
					On("WriteMessages", mock.Anything, lastBatch).
					Return(assert.AnError).
					Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockNotificationRepository)
			writer := new(MockKafkaWriter)
			tt.setupMocks(repo, writer)

			svc := service.NewNotificationRequestsService(repo, writer, tt.batchSize)
			err := svc.SaveNotifications(context.Background(), &tt.notifications)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
			writer.AssertExpectations(t)
		})
	}
}
