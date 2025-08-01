package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
)

// SendNotificationService defines the behavior for sending notifications.
type SendNotificationService interface {
	SendNotification(ctx context.Context, userID int, templateID int) error
}

// OutgoingNotification represents the payload sent to the notification topic.
// UserID identifies the sender user, Template is the message body,
// and Contacts lists the phone-number targets for this batch.
type OutgoingNotification struct {
	UserID   int                   `json:"userID"`
	Template string                `json:"template"`
	Contacts []*models.SlimContact `json:"contacts"`
}
