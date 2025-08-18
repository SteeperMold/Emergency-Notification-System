package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// NotificationRequestsConsumer reads notification requests from Kafka,
// buffers them into in-memory batches by configured size or interval,
// and persists them via the NotificationRequestsService.
type NotificationRequestsConsumer struct {
	service        domain.NotificationRequestsService
	kafkaReader    domain.KafkaReader
	logger         *zap.Logger
	contextTimeout time.Duration
	batchSize      int
	flushInterval  time.Duration
}

// NewNotificationRequestsConsumer constructs the consumer with required dependencies and settings.
func NewNotificationRequestsConsumer(s domain.NotificationRequestsService, kr domain.KafkaReader, logger *zap.Logger, timeout time.Duration, batchSize int, flushInterval time.Duration) *NotificationRequestsConsumer {
	return &NotificationRequestsConsumer{
		service:        s,
		kafkaReader:    kr,
		logger:         logger,
		contextTimeout: timeout,
		batchSize:      batchSize,
		flushInterval:  flushInterval,
	}
}

type message struct {
	raw kafka.Message
	err error
}

// StartConsumer begins polling Kafka for NotificationRequest messages.
// It runs until the provided context is cancelled, automatically flushing
// any buffered notifications on shutdown or when batchSize or flushInterval triggers.
func (nrc *NotificationRequestsConsumer) StartConsumer(ctx context.Context) error {
	msgCh := nrc.startFetchLoop(ctx)
	buffered := make([]*models.Notification, 0, nrc.batchSize)
	ticker := time.NewTicker(nrc.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nrc.flush(ctx, &buffered)

		case <-ticker.C:
			err := nrc.flush(ctx, &buffered)
			if err != nil {
				nrc.logger.Info("failed to flush notifications buffer on interval",
					zap.Int("buffered", len(buffered)),
					zap.Error(err),
				)
			}

		case msg := <-msgCh:
			if msg.err != nil {
				return msg.err
			}
			raw := msg.raw

			var nr domain.NotificationRequest
			err := json.Unmarshal(raw.Value, &nr)
			if err != nil {
				nrc.logger.Error("invalid notification request",
					zap.String("raw_notification", string(raw.Value)),
					zap.Error(err),
				)
				err := nrc.kafkaReader.CommitMessages(ctx, raw)
				if err != nil {
					return err
				}
				continue
			}

			for _, c := range nr.Contacts {
				buffered = append(buffered, &models.Notification{
					ID:             uuid.New(),
					UserID:         nr.UserID,
					Text:           nr.Template,
					RecipientPhone: c.Phone,
				})
			}

			if len(buffered) >= nrc.batchSize {
				err := nrc.flush(ctx, &buffered)
				if err != nil {
					nrc.logger.Info("failed to flush notifications buffer on batch size",
						zap.Int("buffered", len(buffered)),
						zap.Error(err),
					)
				}
			}

			err = nrc.kafkaReader.CommitMessages(ctx, raw)
			if err != nil {
				return err
			}
		}
	}
}

func (nrc *NotificationRequestsConsumer) startFetchLoop(ctx context.Context) chan message {
	ch := make(chan message)
	go func() {
		for {
			msg, err := nrc.kafkaReader.FetchMessage(ctx)
			select {
			case ch <- message{msg, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}

func (nrc *NotificationRequestsConsumer) flush(ctx context.Context, ntfs *[]*models.Notification) error {
	if len(*ntfs) == 0 {
		return nil
	}

	saveCtx, cancel := context.WithTimeout(ctx, nrc.contextTimeout)
	defer cancel()

	start := time.Now()

	err := nrc.service.SaveNotifications(saveCtx, ntfs)
	if err != nil {
		return err
	}

	*ntfs = (*ntfs)[:0]

	duration := time.Since(start)
	nrc.logger.Info("saved notifications batch", zap.Duration("duration", duration))

	return nil
}
