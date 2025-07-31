package service

import (
	"context"
	"encoding/json"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/segmentio/kafka-go"
)

type NotificationRequestsService struct {
	repository  domain.NotificationRepository
	kafkaWriter domain.KafkaWriter
	batchSize   int
}

func NewNotificationRequestsService(r domain.NotificationRepository, kw domain.KafkaWriter, batchSize int) *NotificationRequestsService {
	return &NotificationRequestsService{
		repository:  r,
		kafkaWriter: kw,
		batchSize:   batchSize,
	}
}

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
