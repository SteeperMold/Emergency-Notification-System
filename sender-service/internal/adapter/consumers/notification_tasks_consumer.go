package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
	"go.uber.org/zap"
	"time"
)

type NotificationTasksConsumer struct {
	service        domain.NotificationTasksService
	kafkaReader    domain.KafkaReader
	logger         *zap.Logger
	contextTimeout time.Duration
}

func NewNotificationTasksConsumer(s domain.NotificationTasksService, kr domain.KafkaReader, logger *zap.Logger, timeout time.Duration) *NotificationTasksConsumer {
	return &NotificationTasksConsumer{
		service:        s,
		kafkaReader:    kr,
		logger:         logger,
		contextTimeout: timeout,
	}
}

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
