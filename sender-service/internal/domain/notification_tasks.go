package domain

import (
	"context"
	"fmt"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/models"
	"github.com/google/uuid"
	"time"
)

var (
	ErrNotificationNotExists = fmt.Errorf("notification not exists")
)

type NotificationTasksService interface {
	SendNotification(ctx context.Context, notification *NotificationTask) error
}

type NotificationTasksRepository interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	Reschedule(ctx context.Context, id uuid.UUID, nextRunAt time.Time) (*models.Notification, error)
	MarkFailed(ctx context.Context, id uuid.UUID) (*models.Notification, error)
}

type NotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
	Attempts       int       `json:"attempts"`
}
