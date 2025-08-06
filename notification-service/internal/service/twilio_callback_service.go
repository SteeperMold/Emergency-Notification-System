package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
)

// TwilioCallbackService processes status callbacks from Twilio and updates
// the corresponding notification record in the database.
type TwilioCallbackService struct {
	repository  domain.NotificationRepository
	maxAttempts int
}

// NewTwilioCallbackService constructs a TwilioCallbackService.
func NewTwilioCallbackService(r domain.NotificationRepository, maxAttempts int) *TwilioCallbackService {
	return &TwilioCallbackService{
		repository:  r,
		maxAttempts: maxAttempts,
	}
}

// ProcessCallback handles an incoming Twilio status callback.
func (s *TwilioCallbackService) ProcessCallback(ctx context.Context, idStr, status string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	var newStatus models.NotificationStatus

	switch status {
	case "delivered", "sent":
		newStatus = models.StatusSent
	case "failed", "undelivered":
		ntf, err := s.repository.GetNotificationByID(ctx, id)
		if err != nil {
			return err
		}
		if ntf.Attempts < s.maxAttempts {
			newStatus = models.StatusPending
		} else {
			newStatus = models.StatusFailed
		}
	}

	err = s.repository.ChangeNotificationStatus(ctx, id, newStatus)
	if err != nil {
		return err
	}

	return nil
}
