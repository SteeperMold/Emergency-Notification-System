package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestSignupService_CreateUser(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := service.NewSignupService(mockRepo)

	var (
		ErrNotFound  = fmt.Errorf("not found")
		ErrDBFailure = fmt.Errorf("db failure")
	)

	tests := []struct {
		name        string
		req         *domain.SignupRequest
		setupMocks  func()
		wantErr     error
		wantCreated *models.User
	}{
		{
			name:       "invalid email format",
			req:        &domain.SignupRequest{Email: "not-an-email", Password: "validPass"},
			setupMocks: func() {},
			wantErr:    domain.ErrInvalidEmail,
		},
		{
			name:       "password too short",
			req:        &domain.SignupRequest{Email: "a@b.com", Password: "1234"},
			setupMocks: func() {},
			wantErr:    domain.ErrInvalidPassword,
		},
		{
			name: "email already exists",
			req:  &domain.SignupRequest{Email: "alice@example.com", Password: "longenough"},
			setupMocks: func() {
				mockRepo.
					On("GetUserByEmail", mock.Anything, "alice@example.com").
					Return(&models.User{Email: "alice@example.com"}, nil).
					Once()
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name: "successful create",
			req:  &domain.SignupRequest{Email: "bob@example.com", Password: "securePassword"},
			setupMocks: func() {
				mockRepo.
					On("GetUserByEmail", mock.Anything, "bob@example.com").
					Return(nil, ErrNotFound).
					Once()

				created := &models.User{
					ID:           42,
					Email:        "bob@example.com",
					CreationTime: time.Now(),
				}
				mockRepo.
					On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
						err := bcrypt.CompareHashAndPassword(
							[]byte(u.PasswordHash),
							[]byte("securePassword"),
						)
						return u.Email == "bob@example.com" && err == nil
					})).
					Return(created, nil).
					Once()
			},
			wantErr:     nil,
			wantCreated: &models.User{ID: 42, Email: "bob@example.com"},
		},
		{
			name: "repository CreateUser error",
			req:  &domain.SignupRequest{Email: "z@z.com", Password: "validPass"},
			setupMocks: func() {
				mockRepo.
					On("GetUserByEmail", mock.Anything, "z@z.com").
					Return(nil, ErrNotFound).
					Once()

				mockRepo.
					On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(nil, ErrDBFailure).
					Once()
			},
			wantErr: ErrDBFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tc.setupMocks()

			got, err := svc.CreateUser(context.Background(), tc.req)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantCreated.ID, got.ID)
				assert.Equal(t, tc.wantCreated.Email, got.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
