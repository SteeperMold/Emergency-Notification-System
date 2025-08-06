package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestSignupService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		req       *domain.SignupRequest
		setupMock func(m *MockUserRepository)
		wantErr   error
	}{
		{
			name:    "invalid email",
			req:     &domain.SignupRequest{Email: "not-an-email", Password: "validPass"},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "password too short",
			req:     &domain.SignupRequest{Email: "a@b.com", Password: "1234"},
			wantErr: domain.ErrInvalidPassword,
		},
		{
			name:    "password too long",
			req:     &domain.SignupRequest{Email: "a@b.com", Password: strings.Repeat("x", 101)},
			wantErr: domain.ErrInvalidPassword,
		},
		{
			name: "email already exists",
			req:  &domain.SignupRequest{Email: "user@example.com", Password: "goodPass"},
			setupMock: func(m *MockUserRepository) {
				m.
					On("GetUserByEmail", mock.Anything, "user@example.com").
					Return(&models.User{Email: "user@example.com"}, nil).
					Once()
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name: "repo.CreateUser failure",
			req:  &domain.SignupRequest{Email: "new@user.com", Password: "goodPass"},
			setupMock: func(m *MockUserRepository) {
				// No existing user
				m.
					On("GetUserByEmail", mock.Anything, "new@user.com").
					Return((*models.User)(nil), domain.ErrUserNotExists).
					Once()
				// CreateUser returns DB error
				m.
					On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
						return u.Email == "new@user.com" && len(u.PasswordHash) > 0
					})).
					Return((*models.User)(nil), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			req:  &domain.SignupRequest{Email: "a@b.com", Password: "goodPass"},
			setupMock: func(m *MockUserRepository) {
				// Not found on lookup
				m.
					On("GetUserByEmail", mock.Anything, "a@b.com").
					Return((*models.User)(nil), domain.ErrUserNotExists).
					Once()
				// CreateUser returns created user
				m.
					On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
						// ensure email set and password hashed
						return u.Email == "a@b.com" && bcrypt.CompareHashAndPassword(
							[]byte(u.PasswordHash), []byte("goodPass")) == nil
					})).
					Return(&models.User{ID: 123, Email: "a@b.com", PasswordHash: "ignored"}, nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			svc := service.NewSignupService(mockRepo)
			usr, err := svc.CreateUser(context.Background(), tc.req)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, usr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "a@b.com", usr.Email)
				assert.NotEmpty(t, usr.PasswordHash)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
