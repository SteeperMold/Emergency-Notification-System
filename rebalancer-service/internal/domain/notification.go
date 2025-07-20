package domain

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/models"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	FetchPending(ctx context.Context, limit int) ([]*models.Notification, error)
}

type SendNotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
}
