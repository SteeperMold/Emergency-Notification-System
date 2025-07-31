package service

import (
	"context"
	"fmt"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
	"time"
)

type NotificationTasksService struct {
	repository  domain.NotificationTasksRepository
	smsSender   domain.SmsSender
	maxAttempts int
}

func NewNotificationTasksService(r domain.NotificationTasksRepository, ss domain.SmsSender, maxAttempts int) *NotificationTasksService {
	return &NotificationTasksService{
		repository:  r,
		smsSender:   ss,
		maxAttempts: maxAttempts,
	}
}

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
