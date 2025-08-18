package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSmsSender struct {
	mock.Mock
}

func (m *MockSmsSender) SendSMS(phone, text, id string) error {
	return m.Called(phone, text, id).Error(0)
}

type MockNotificationTasksRepository struct {
	mock.Mock
}

func (m *MockNotificationTasksRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationTasksRepository) Reschedule(ctx context.Context, id uuid.UUID, nextRunAt time.Time) (*models.Notification, error) {
	args := m.Called(ctx, id, nextRunAt)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationTasksRepository) MarkFailed(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func TestSendNotification(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	tasks := map[string]struct {
		task        domain.NotificationTask
		maxAttempts int
		senderErr   error
		repoSetup   func(r *MockNotificationTasksRepository, task domain.NotificationTask)
		expectErr   bool
	}{
		"success": {
			task:        domain.NotificationTask{ID: id, RecipientPhone: "+100", Text: "hello", Attempts: 1},
			maxAttempts: 3,
			senderErr:   nil,
			repoSetup:   func(r *MockNotificationTasksRepository, task domain.NotificationTask) {},
			expectErr:   false,
		},
		"reschedule on first failure": {
			task:        domain.NotificationTask{ID: id, RecipientPhone: "+100", Text: "hello", Attempts: 1},
			maxAttempts: 3,
			senderErr:   errors.New("sms down"),
			repoSetup: func(r *MockNotificationTasksRepository, task domain.NotificationTask) {
				nextRunMin := time.Now().Add(1 * time.Second)
				nextRunMax := time.Now().Add(1*time.Second + 50*time.Millisecond)
				r.
					On("Reschedule", mock.Anything, task.ID, mock.MatchedBy(func(t time.Time) bool {
						// allow small clock skew
						return t.After(nextRunMin) && t.Before(nextRunMax)
					})).
					Return((*models.Notification)(nil), nil).
					Once()
			},
			expectErr: false,
		},
		"reschedule repo error": {
			task:        domain.NotificationTask{ID: id, RecipientPhone: "+100", Text: "hello", Attempts: 2},
			maxAttempts: 3,
			senderErr:   errors.New("sms down"),
			repoSetup: func(r *MockNotificationTasksRepository, task domain.NotificationTask) {
				r.
					On("Reschedule", mock.Anything, task.ID, mock.Anything).
					Return((*models.Notification)(nil), assert.AnError).
					Once()
			},
			expectErr: true,
		},
		"mark failed on max attempts": {
			task:        domain.NotificationTask{ID: id, RecipientPhone: "+100", Text: "hello", Attempts: 3},
			maxAttempts: 3,
			senderErr:   errors.New("sms fail"),
			repoSetup: func(r *MockNotificationTasksRepository, task domain.NotificationTask) {
				r.
					On("MarkFailed", mock.Anything, task.ID).
					Return((*models.Notification)(nil), nil).
					Once()
			},
			expectErr: false,
		},
		"mark failed repo error": {
			task:        domain.NotificationTask{ID: id, RecipientPhone: "+100", Text: "hello", Attempts: 5},
			maxAttempts: 5,
			senderErr:   errors.New("sms fail"),
			repoSetup: func(r *MockNotificationTasksRepository, task domain.NotificationTask) {
				r.
					On("MarkFailed", mock.Anything, task.ID).
					Return((*models.Notification)(nil), assert.AnError).
					Once()
			},
			expectErr: true,
		},
	}

	for name, tc := range tasks {
		t.Run(name, func(t *testing.T) {
			sender := &MockSmsSender{}
			repo := &MockNotificationTasksRepository{}
			sender.
				On("SendSMS", tc.task.RecipientPhone, tc.task.Text, tc.task.ID.String()).
				Return(tc.senderErr)

			tc.repoSetup(repo, tc.task)

			svc := service.NewNotificationTasksService(repo, sender, tc.maxAttempts)

			err := svc.SendNotification(ctx, &tc.task)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			sender.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
