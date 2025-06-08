package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshTokenService_GetUserById(t *testing.T) {
	tests := []struct {
		name      string
		inputID   int
		mockUser  *models.User
		mockErr   error
		expect    *models.User
		expectErr error
	}{
		{
			name:      "valid id",
			inputID:   1,
			mockUser:  &models.User{ID: 1, Email: "a@b.com"},
			mockErr:   nil,
			expect:    &models.User{ID: 1, Email: "a@b.com"},
			expectErr: nil,
		},
		{
			name:      "user not found",
			inputID:   2,
			mockUser:  nil,
			mockErr:   domain.ErrUserNotExists,
			expect:    nil,
			expectErr: domain.ErrUserNotExists,
		},
		{
			name:      "db error",
			inputID:   3,
			mockUser:  nil,
			mockErr:   errors.New("db down"),
			expect:    nil,
			expectErr: errors.New("db down"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockRepo.
				On("GetUserByID", mock.Anything, tc.inputID).
				Return(tc.mockUser, tc.mockErr)

			s := service.NewRefreshTokenService(mockRepo)
			user, err := s.GetUserByID(context.Background(), tc.inputID)

			assert.Equal(t, tc.expect, user)
			if tc.expectErr != nil {
				assert.EqualError(t, err, tc.expectErr.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
