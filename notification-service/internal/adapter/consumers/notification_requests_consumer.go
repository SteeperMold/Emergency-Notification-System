package consumers

import (
	"context"
	"encoding/json"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"go.uber.org/zap"
	"time"
)

type NotificationRequestsConsumer struct {
	service        domain.NotificationRequestsService
	kafkaReader    domain.KafkaReader
	logger         *zap.Logger
	contextTimeout time.Duration
}

func NewNotificationRequestsConsumer(s domain.NotificationRequestsService, kafkaReader domain.KafkaReader, logger *zap.Logger, timeout time.Duration) *NotificationRequestsConsumer {
	return &NotificationRequestsConsumer{
		service:        s,
		kafkaReader:    kafkaReader,
		logger:         logger,
		contextTimeout: timeout,
	}
}

func (nrc *NotificationRequestsConsumer) StartConsumer(ctx context.Context) error {
	for {
		msg, err := nrc.kafkaReader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var nr domain.NotificationRequest
		err = json.Unmarshal(msg.Value, &nr)
		if err != nil {
			nrc.logger.Error("invalid notification request",
				zap.String("raw_notification", string(msg.Value)),
				zap.Error(err),
			)
			err := nrc.kafkaReader.CommitMessages(ctx, msg)
			if err != nil {
				return err
			}
			continue
		}

		msgCtx, cancel := context.WithTimeout(ctx, nrc.contextTimeout)
		start := time.Now()
		nrc.logger.Info("read notification request", zap.Int("user_id", nr.UserID))

		err = nrc.service.SaveNotification(msgCtx, &nr)
		if err != nil {
			nrc.logger.Error("failed to save notification",
				zap.Int("user_id", nr.UserID),
				zap.Error(err),
			)
		}
		cancel()

		duration := time.Since(start)
		nrc.logger.Info("finished saving notification request",
			zap.Int("user_id", nr.UserID),
			zap.Duration("duration", duration),
		)

		if err := nrc.kafkaReader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
