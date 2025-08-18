package domain

import (
	"context"
	"fmt"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrNotificationNotExists is returned when an operation references
	// a notification that does not exist in the database
	ErrNotificationNotExists = fmt.Errorf("notification not exists")
)

// NotificationRequestsService defines the behavior for handling incoming notification batches
type NotificationRequestsService interface {
	SaveNotifications(ctx context.Context, notifications *[]*models.Notification) error
}

// NotificationRepository encapsulates database operations
// for notifications, including bulk creation and status updates
type NotificationRepository interface {
	CreateMultipleNotifications(ctx context.Context, notifications []*models.Notification) error
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	ChangeNotificationStatus(ctx context.Context, id uuid.UUID, newStatus models.NotificationStatus) error
}

// NotificationRequest represents the payload received from the API
// containing a template and a list of contacts to notify.
type NotificationRequest struct {
	UserID   int                   `json:"userID"`
	Template string                `json:"template"`
	Contacts []*models.SlimContact `json:"contacts"`
}

// SendNotificationTask describes the individual unit of work
// sent to a worker for sending a single SMS.
type SendNotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
	Attempts       int       `json:"attempts"`
}
