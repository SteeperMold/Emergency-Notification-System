package service

import (
	"context"
	"encoding/json"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/segmentio/kafka-go"
)

type SendNotificationService struct {
	contactsRepository domain.ContactsRepository
	templateRepository domain.TemplateRepository
	kafkaWriter        domain.KafkaWriter
	contactsPerMessage int
}

func NewSendNotificationService(cr domain.ContactsRepository, tr domain.TemplateRepository, kw domain.KafkaWriter, cpm int) *SendNotificationService {
	return &SendNotificationService{
		contactsRepository: cr,
		templateRepository: tr,
		kafkaWriter:        kw,
		contactsPerMessage: cpm,
	}
}

func (sns *SendNotificationService) SendNotification(ctx context.Context, userId int, templateID int) error {
	tmpl, err := sns.templateRepository.GetTemplateByID(ctx, userId, templateID)
	if err != nil {
		return err
	}

	contacts, err := sns.contactsRepository.GetContactsByUserID(ctx, userId)
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
			UserID:   userId,
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
