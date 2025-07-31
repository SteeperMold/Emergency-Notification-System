package domain

import (
	"context"
	"fmt"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
)

var (
	ErrNotificationNotExists = fmt.Errorf("notification not exists")
)

type NotificationRequestsService interface {
	SaveNotifications(ctx context.Context, notifications *[]*models.Notification) error
}

type NotificationRepository interface {
	CreateMultipleNotifications(ctx context.Context, notifications []*models.Notification) error
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	ChangeNotificationStatus(ctx context.Context, id uuid.UUID, newStatus models.NotificationStatus) error
}

type NotificationRequest struct {
	UserID   int                   `json:"userID"`
	Template string                `json:"template"`
	Contacts []*models.SlimContact `json:"contacts"`
}

type SendNotificationTask struct {
	ID             uuid.UUID `json:"id"`
	Text           string    `json:"text"`
	RecipientPhone string    `json:"recipientPhone"`
	Attempts       int       `json:"attempts"`
}
