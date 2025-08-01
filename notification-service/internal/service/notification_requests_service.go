package service

import (
	"context"
	"encoding/json"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/segmentio/kafka-go"
)

// NotificationRequestsService coordinates persistence of new notifications
// and dispatching tasks to Kafka for downstream processing.
type NotificationRequestsService struct {
	repository  domain.NotificationRepository
	kafkaWriter domain.KafkaWriter
	batchSize   int
}

// NewNotificationRequestsService constructs a NotificationRequestsService
func NewNotificationRequestsService(r domain.NotificationRepository, kw domain.KafkaWriter, batchSize int) *NotificationRequestsService {
	return &NotificationRequestsService{
		repository:  r,
		kafkaWriter: kw,
		batchSize:   batchSize,
	}
}

// SaveNotifications persists a slice of notifications to the database and then
// publishes SendNotificationTask messages to Kafka in batches.
func (nrs *NotificationRequestsService) SaveNotifications(ctx context.Context, ntfs *[]*models.Notification) error {
	err := nrs.repository.CreateMultipleNotifications(ctx, *ntfs)
	if err != nil {
		return err
	}

	msgs := make([]kafka.Message, len(*ntfs))
	for i, n := range *ntfs {
		taskBytes, err := json.Marshal(&domain.SendNotificationTask{
			ID:             n.ID,
			Text:           n.Text,
			RecipientPhone: n.RecipientPhone,
			Attempts:       1,
		})
		if err != nil {
			return err
		}
		msgs[i] = kafka.Message{Value: taskBytes}
	}

	for start := 0; start < len(msgs); start += nrs.batchSize {
		end := start + nrs.batchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		batch := msgs[start:end]

		err := nrs.kafkaWriter.WriteMessages(ctx, batch...)
		if err != nil {
			return err
		}
	}

	return nil
}
