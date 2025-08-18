package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrNotificationNotExists is returned when a notification with the given ID does not exist in the repository.
	ErrNotificationNotExists = fmt.Errorf("notification not exists")
)

// NotificationTasksService defines the interface for processing and sending notification tasks.
type NotificationTasksService interface {
	SendNotification(ctx context.Context, notification *NotificationTask) error
}

// NotificationTasksRepository defines the interface for interacting with the notifications data store.
type NotificationTasksRepository interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	Reschedule(ctx context.Context, id uuid.UUID, nextRunAt time.Time) (*models.Notification, error)
	MarkFailed(ctx context.Context, id uuid.UUID) (*models.Notification, error)
}

// NotificationTask represents a task to send a single notification to a recipient.
type NotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
	Attempts       int       `json:"attempts"`
}
