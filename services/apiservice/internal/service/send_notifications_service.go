package service

import (
	"context"
	"encoding/json"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/segmentio/kafka-go"
)

// SendNotificationService orchestrates reading a template and contacts,
// splitting them into chunks, and emitting one Kafka message per chunk.
type SendNotificationService struct {
	contactsRepository domain.ContactsRepository
	templateRepository domain.TemplateRepository
	kafkaWriter        domain.KafkaWriter
	contactsPerMessage int
}

// NewSendNotificationService constructs a SendNotificationService.
func NewSendNotificationService(cr domain.ContactsRepository, tr domain.TemplateRepository, kw domain.KafkaWriter, cpm int) *SendNotificationService {
	return &SendNotificationService{
		contactsRepository: cr,
		templateRepository: tr,
		kafkaWriter:        kw,
		contactsPerMessage: cpm,
	}
}

// SendNotification loads the template and contacts for userId/templateID,
// splits contacts into batches of size contactsPerMessage, and writes one
// Kafka message per batch. Returns an error if any repository or Kafka call fails.
func (sns *SendNotificationService) SendNotification(ctx context.Context, userID int, templateID int) error {
	tmpl, err := sns.templateRepository.GetTemplateByID(ctx, userID, templateID)
	if err != nil {
		return err
	}

	contacts, err := sns.contactsRepository.GetContactsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(contacts) == 0 {
		return domain.ErrContactNotExists
	}

	slimContacts := models.ToSlim(contacts)

	for start := 0; start < len(slimContacts); start += sns.contactsPerMessage {
		end := start + sns.contactsPerMessage
		if end > len(slimContacts) {
			end = len(slimContacts)
		}
		chunk := slimContacts[start:end]

		notification := &domain.OutgoingNotification{
			UserID:   userID,
			Template: tmpl.Body,
			Contacts: chunk,
		}

		msgBytes, err := json.Marshal(notification)
		if err != nil {
			return err
		}

		err = sns.kafkaWriter.WriteMessages(ctx, kafka.Message{
			Value: msgBytes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
