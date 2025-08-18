package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/domain"
	"go.uber.org/zap"
)

// ContactsConsumer reads contact-loading tasks from Kafka, invokes the ContactsService,
// and commits offsets after each message is handled.
type ContactsConsumer struct {
	service     domain.ContactsService
	kafkaReader domain.KafkaReader
	logger      *zap.Logger
}

// NewContactsConsumer constructs a ContactsConsumer.
func NewContactsConsumer(s domain.ContactsService, kafkaReader domain.KafkaReader, logger *zap.Logger) *ContactsConsumer {
	return &ContactsConsumer{
		service:     s,
		kafkaReader: kafkaReader,
		logger:      logger,
	}
}

// StartConsumer enters a loop fetching messages indefinitely until the context is canceled
// or an unrecoverable error occurs.
func (cc *ContactsConsumer) StartConsumer(ctx context.Context) error {
	for {
		msg, err := cc.kafkaReader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var t domain.Task
		err = json.Unmarshal(msg.Value, &t)
		if err != nil {
			cc.logger.Error("invalid task", zap.String("raw_task", string(msg.Value)), zap.Error(err))
			err := cc.kafkaReader.CommitMessages(ctx, msg)
			if err != nil {
				return err
			}
			continue
		}

		start := time.Now()
		cc.logger.Info("started task",
			zap.Int("user_id", t.UserID),
			zap.String("s3_key", t.S3Key),
		)

		processedContacts, err := cc.service.ProcessFile(ctx, &t)
		if err != nil {
			cc.logger.Error("failed to process task", zap.String("file_key", t.S3Key), zap.Error(err))
			err := cc.kafkaReader.CommitMessages(ctx, msg)
			if err != nil {
				return err
			}
			continue
		}

		duration := time.Since(start)
		cc.logger.Info("finished task",
			zap.Int("user_id", t.UserID),
			zap.String("s3_key", t.S3Key),
			zap.Duration("duration", duration),
			zap.Int("processed_contacts", processedContacts),
		)

		err = cc.kafkaReader.CommitMessages(ctx, msg)
		if err != nil {
			return err
		}
	}
}
