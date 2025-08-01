package service

import (
	"context"
	"fmt"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
)

// NotificationTasksService coordinates the delivery and retry logic for SMS notification tasks.
// It sends notifications using the provided SmsSender and handles rescheduling or marking as failed
// based on the result and number of attempts.
type NotificationTasksService struct {
	repository  domain.NotificationTasksRepository
	smsSender   domain.SmsSender
	maxAttempts int
}

// NewNotificationTasksService creates a new NotificationTasksService.
func NewNotificationTasksService(r domain.NotificationTasksRepository, ss domain.SmsSender, maxAttempts int) *NotificationTasksService {
	return &NotificationTasksService{
		repository:  r,
		smsSender:   ss,
		maxAttempts: maxAttempts,
	}
}

// SendNotification attempts to send a notification task via SMS.
// If sending fails and the attempt count is below the maximum, it reschedules the task using exponential backoff.
// If the maximum number of attempts is reached, it marks the task as permanently failed.
func (nts *NotificationTasksService) SendNotification(ctx context.Context, task *domain.NotificationTask) error {
	err := nts.smsSender.SendSMS(task.RecipientPhone, task.Text, task.ID.String())
	if err != nil {
		if task.Attempts < nts.maxAttempts {
			// exponential backoff: base * 2^(attempts-1)
			delay := time.Second * (1 << (task.Attempts - 1))
			nextRunAt := time.Now().Add(delay)

			_, repoErr := nts.repository.Reschedule(ctx, task.ID, nextRunAt)
			if repoErr != nil {
				return fmt.Errorf("send failed: %w; reschedule failed: %v", err, repoErr)
			}
		} else {
			_, repoErr := nts.repository.MarkFailed(ctx, task.ID)
			if repoErr != nil {
				return fmt.Errorf("send failed: %w; mark failed error: %v", err, repoErr)
			}
		}
		// we swallow the send error so the caller knows we handled it
		return nil
	}

	return nil
}
