package domain

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
)

type SendNotificationService interface {
	SendNotification(ctx context.Context, userId int, templateID int) error
}

type OutgoingNotification struct {
	UserID   int                   `json:"userID"`
	Template string                `json:"template"`
	Contacts []*models.SlimContact `json:"contacts"`
}
