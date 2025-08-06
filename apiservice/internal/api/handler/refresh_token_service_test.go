package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/tokenutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestRefreshTokenHandler(t *testing.T) {
	user := &models.User{ID: 42, Email: "john@example.com"}
	refreshSecret := "refreshsecret"
	accessSecret := "accesssecret"

	validRefresh, _ := tokenutils.CreateRefreshToken(user, refreshSecret, time.Minute)
	invalidRefresh := "invalid.token.value"

	tests := []struct {
		name           string
		refreshToken   string
		mockSetup      func(m *MockRefreshTokenService)
		expectedStatus int
	}{
		{
			name:         "valid token",
			refreshToken: validRefresh,
			mockSetup: func(m *MockRefreshTokenService) {
				m.
					On("GetUserByID", mock.Anything, 42).
					Return(user, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "malformed token",
			refreshToken:   invalidRefresh,
			mockSetup:      func(m *MockRefreshTokenService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:         "user not found",
			refreshToken: validRefresh,
			mockSetup: func(m *MockRefreshTokenService) {
				m.
					On("GetUserByID", mock.Anything, 42).
					Return(&models.User{}, assert.AnError).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(domain.RefreshTokenRequest{RefreshToken: tt.refreshToken})
			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			m := new(MockRefreshTokenService)
			tt.mockSetup(m)

			h := handler.NewRefreshTokenHandler(m, zaptest.NewLogger(t), time.Second, &bootstrap.JWTConfig{
				AccessSecret:  accessSecret,
				AccessExpiry:  time.Minute,
				RefreshSecret: refreshSecret,
				RefreshExpiry: time.Hour,
			})

			h.RefreshToken(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			m.AssertExpectations(t)
		})
	}
}
