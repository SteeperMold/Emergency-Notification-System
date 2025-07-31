package service

import (
	"context"
	"encoding/json"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

type RebalancerService struct {
	repository     domain.NotificationRepository
	kafkaWriter    domain.KafkaWriter
	logger         *zap.Logger
	batchSize      int
	interval       time.Duration
	contextTimeout time.Duration
}

func NewRebalancerService(r domain.NotificationRepository, kw domain.KafkaWriter, logger *zap.Logger, batchSize int, interval, timeout time.Duration) *RebalancerService {
	return &RebalancerService{
		repository:     r,
		kafkaWriter:    kw,
		logger:         logger,
		batchSize:      batchSize,
		interval:       interval,
		contextTimeout: timeout,
	}
}

func (rs *RebalancerService) Start(ctx context.Context) {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rs.rebalance(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (rs *RebalancerService) rebalance(ctx context.Context) {
	dbCtx, cancel := context.WithTimeout(ctx, rs.contextTimeout)
	defer cancel()

	notifications, err := rs.repository.FetchAndUpdatePending(dbCtx, rs.batchSize)
	if err != nil {
		rs.logger.Error("failed to fetch pending notifications", zap.Error(err))
		return
	}

	if len(notifications) == 0 {
		rs.logger.Info("no pending notifications, skipping...")
		return
	}

	msgs := make([]kafka.Message, len(notifications))
	for i, n := range notifications {
		taskBytes, err := json.Marshal(&domain.SendNotificationTask{
			ID:             n.ID,
			Text:           n.Text,
			RecipientPhone: n.RecipientPhone,
			Attempts:       n.Attempts,
		})
		if err != nil {
			rs.logger.Error("failed to marshal task", zap.Error(err))
			return
		}
		msgs[i] = kafka.Message{Value: taskBytes}
	}

	rs.logger.Info("wrote pending notifications to kafka", zap.Int("notifications_count", len(msgs)))

	err = rs.kafkaWriter.WriteMessages(ctx, msgs...)
	if err != nil {
		rs.logger.Error("failed to write to kafka", zap.Error(err))
	}
}
