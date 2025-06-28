package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"go.uber.org/zap"
)

// ContactsConsumer consumes contact‐creation messages from Kafka and
// invokes the ContactsService to persist each contact.
type ContactsConsumer struct {
	service     domain.ContactsService
	kafkaReader domain.KafkaReader
	logger      *zap.Logger
}

// NewContactsKafkaConsumer constructs a ContactsConsumer.
func NewContactsKafkaConsumer(s domain.ContactsService, kafkaReader domain.KafkaReader, logger *zap.Logger) *ContactsConsumer {
	return &ContactsConsumer{
		service:     s,
		kafkaReader: kafkaReader,
		logger:      logger,
	}
}

// StartConsumer begins an infinite loop to process incoming Kafka messages.
func (cc *ContactsConsumer) StartConsumer(ctx context.Context) error {
	for {
		msg, err := cc.kafkaReader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var ct models.Contact
		err = json.Unmarshal(msg.Value, &ct)
		if err != nil {
			cc.logger.Error("invalid contact payload", zap.String("raw_task", string(msg.Value)), zap.Error(err))
			err := cc.kafkaReader.CommitMessages(ctx, msg)
			if err != nil {
				return err
			}
			continue
		}

		start := time.Now()
		cc.logger.Info("read contact save task", zap.Int("user_id", ct.UserID))

		_, err = cc.service.CreateContact(ctx, &ct)
		if err != nil && !errors.Is(err, domain.ErrContactAlreadyExists) {
			cc.logger.Error("failed to save contact", zap.Int("user_id", ct.UserID), zap.Error(err))
		}

		duration := time.Since(start)
		cc.logger.Info("finished saving task",
			zap.Int("user_id", ct.UserID),
			zap.Duration("duration", duration),
		)

		if err := cc.kafkaReader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
