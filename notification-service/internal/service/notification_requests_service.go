package service

import (
	"context"
	"encoding/json"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type NotificationRequestsService struct {
	repository  domain.NotificationRepository
	kafkaWriter domain.KafkaWriter
}

func NewNotificationRequestsService(r domain.NotificationRepository, kw domain.KafkaWriter) *NotificationRequestsService {
	return &NotificationRequestsService{
		repository:  r,
		kafkaWriter: kw,
	}
}

func (nrs *NotificationRequestsService) SaveNotification(ctx context.Context, nr *domain.NotificationRequest) error {
	notifications := make([]*models.Notification, len(nr.Contacts))

	for i, c := range nr.Contacts {
		notifications[i] = &models.Notification{
			ID:             uuid.New(),
			UserID:         nr.UserID,
			Text:           nr.Template,
			RecipientPhone: c.Phone,
		}
	}

	err := nrs.repository.CreateMultipleNotifications(ctx, notifications)
	if err != nil {
		return err
	}

	msgs := make([]kafka.Message, len(notifications))
	for i, n := range notifications {
		taskBytes, err := json.Marshal(&domain.SendNotificationTask{
			ID:             n.ID,
			Text:           n.Text,
			RecipientPhone: n.RecipientPhone,
		})
		if err != nil {
			return err
		}
		msgs[i] = kafka.Message{Value: taskBytes}
	}

	const batchSize = 1000
	for start := 0; start < len(msgs); start += batchSize {
		end := start + batchSize
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
