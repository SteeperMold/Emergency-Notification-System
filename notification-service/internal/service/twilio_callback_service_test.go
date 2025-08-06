package service_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTwilioCallbackService_ProcessCallback(t *testing.T) {
	validID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name          string
		idStr         string
		status        string
		setupMocks    func(r *MockNotificationRepository)
		expectedError bool
	}{
		{
			name:   "status delivered",
			idStr:  validID.String(),
			status: "delivered",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("ChangeNotificationStatus", mock.Anything, validID, models.StatusSent).
					Return(nil).
					Once()
			},
		},
		{
			name:   "status sent",
			idStr:  validID.String(),
			status: "sent",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("ChangeNotificationStatus", mock.Anything, validID, models.StatusSent).
					Return(nil).
					Once()
			},
		},
		{
			name:   "failed with attempts < max => retry",
			idStr:  validID.String(),
			status: "failed",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("GetNotificationByID", mock.Anything, validID).
					Return(&models.Notification{
						ID:       validID,
						Attempts: 1,
					}, nil).
					Once()
				r.
					On("ChangeNotificationStatus", mock.Anything, validID, models.StatusPending).
					Return(nil).
					Once()
			},
		},
		{
			name:   "undelivered with attempts >= max => failed",
			idStr:  validID.String(),
			status: "undelivered",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("GetNotificationByID", mock.Anything, validID).
					Return(&models.Notification{
						ID:       validID,
						Attempts: 3,
					}, nil).
					Once()
				r.
					On("ChangeNotificationStatus", mock.Anything, validID, models.StatusFailed).
					Return(nil).
					Once()
			},
		},
		{
			name:          "invalid UUID",
			idStr:         "invalid-uuid",
			status:        "delivered",
			setupMocks:    func(r *MockNotificationRepository) {},
			expectedError: true,
		},
		{
			name:   "repo GetNotificationByID fails",
			idStr:  validID.String(),
			status: "failed",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("GetNotificationByID", mock.Anything, validID).
					Return((*models.Notification)(nil), assert.AnError).
					Once()
			},
			expectedError: true,
		},
		{
			name:   "repo ChangeNotificationStatus fails",
			idStr:  validID.String(),
			status: "sent",
			setupMocks: func(r *MockNotificationRepository) {
				r.
					On("ChangeNotificationStatus", mock.Anything, validID, models.StatusSent).
					Return(assert.AnError).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockNotificationRepository)
			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}
			svc := service.NewTwilioCallbackService(repo, 3)

			err := svc.ProcessCallback(context.Background(), tt.idStr, tt.status)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
