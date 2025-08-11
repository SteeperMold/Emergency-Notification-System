package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/rebalancer-service/internal/models"
	"github.com/google/uuid"
)

// NotificationRepository defines the data access methods for notifications
// that are due for delivery or retry.
type NotificationRepository interface {
	FetchAndUpdatePending(ctx context.Context, limit int) ([]*models.Notification, error)
}

// SendNotificationTask describes the payload sent to worker services
// for delivering a single notification via SMS or other channels.
type SendNotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
	Attempts       int       `json:"attempts"`
}
