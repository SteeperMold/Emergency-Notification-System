package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
	"go.uber.org/zap"
)

// NotificationTasksConsumer is responsible for consuming notification tasks from a Kafka topic,
// sending them using the NotificationTasksService, and committing messages based on success or failure.
type NotificationTasksConsumer struct {
	service        domain.NotificationTasksService
	kafkaReader    domain.KafkaReader
	logger         *zap.Logger
	contextTimeout time.Duration
}

// NewNotificationTasksConsumer creates a new instance of NotificationTasksConsumer.
func NewNotificationTasksConsumer(s domain.NotificationTasksService, kr domain.KafkaReader, logger *zap.Logger, timeout time.Duration) *NotificationTasksConsumer {
	return &NotificationTasksConsumer{
		service:        s,
		kafkaReader:    kr,
		logger:         logger,
		contextTimeout: timeout,
	}
}

// StartConsumer continuously reads messages from the Kafka topic, decodes them into NotificationTasks,
// processes them using the NotificationTasksService, and commits messages to Kafka accordingly.
// Retryable errors are skipped to allow future retries; permanent failures are logged and committed.
func (ntc *NotificationTasksConsumer) StartConsumer(ctx context.Context) error {
	for {
		msg, err := ntc.kafkaReader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var nt domain.NotificationTask
		err = json.Unmarshal(msg.Value, &nt)
		if err != nil {
			ntc.logger.Error("invalid notification task", zap.String("raw_task", string(msg.Value)), zap.Error(err))
			err := ntc.kafkaReader.CommitMessages(ctx, msg)
			if err != nil {
				return err
			}
			continue
		}

		msgCtx, cancel := context.WithTimeout(ctx, ntc.contextTimeout)
		start := time.Now()

		ntc.logger.Info("read notification task", zap.String("notification_id", nt.ID.String()))
		err = ntc.service.SendNotification(msgCtx, &nt)
		cancel()

		if err != nil {
			var rerr domain.SendError
			if errors.As(err, &rerr) && rerr.Retryable() {
				ntc.logger.Info("retryable send failure, will retry", zap.String("notification_id", nt.ID.String()), zap.Error(err))
				continue
			}

			ntc.logger.Error("permanent send failure, commiting", zap.String("notification_id", nt.ID.String()), zap.Error(err))
		}

		duration := time.Since(start)
		ntc.logger.Info("finished notification task", zap.String("notification_id", nt.ID.String()), zap.Duration("duration", duration))

		if err := ntc.kafkaReader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
