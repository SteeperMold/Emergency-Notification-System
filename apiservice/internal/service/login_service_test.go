package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		mockUser *models.User
		mockErr  error
		wantErr  error
	}{
		{
			name:     "user found",
			email:    "a@example.com",
			mockUser: &models.User{ID: 1, Email: "a@example.com"},
			mockErr:  nil,
			wantErr:  nil,
		},
		{
			name:     "user not found",
			email:    "unknown@example.com",
			mockUser: nil,
			mockErr:  errors.New("not found"),
			wantErr:  errors.New("not found"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			svc := service.NewLoginService(mockRepo)

			mockRepo.On("GetUserByEmail", mock.Anything, tc.email).
				Return(tc.mockUser, tc.mockErr).
				Once()

			user, err := svc.GetUserByEmail(context.Background(), tc.email)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockUser, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLoginService_CompareCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &models.User{PasswordHash: string(hash)}
	svc := service.NewLoginService(nil)

	t.Run("correct password", func(t *testing.T) {
		request := &domain.LoginRequest{Password: "secret"}
		assert.True(t, svc.CompareCredentials(user, request))
	})

	t.Run("incorrect password", func(t *testing.T) {
		request := &domain.LoginRequest{Password: "wrong"}
		assert.False(t, svc.CompareCredentials(user, request))
	})
}
