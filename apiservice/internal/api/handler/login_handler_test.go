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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestLoginHandler_Login(t *testing.T) {
	jwtCfg := &bootstrap.JWTConfig{
		AccessSecret:  "access-secret",
		AccessExpiry:  time.Minute,
		RefreshSecret: "refresh-secret",
		RefreshExpiry: time.Hour,
	}

	tests := []struct {
		name             string
		body             string
		setupMock        func(s *MockLoginService)
		wantStatus       int
		expectTokenPairs bool
	}{
		{
			name:       "malformed JSON",
			body:       `{not-a-json}`,
			setupMock:  func(m *MockLoginService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			body: `{"email":"x@y.com","password":"p"}`,
			setupMock: func(m *MockLoginService) {
				m.
					On("GetUserByEmail", mock.Anything, "x@y.com").
					Return((*models.User)(nil), domain.ErrUserNotExists).
					Once()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "service error",
			body: `{"email":"err@y.com","password":"p"}`,
			setupMock: func(m *MockLoginService) {
				m.
					On("GetUserByEmail", mock.Anything, "err@y.com").
					Return((*models.User)(nil), assert.AnError).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "bad credentials",
			body: `{"email":"u@b.com","password":"wrong"}`,
			setupMock: func(m *MockLoginService) {
				user := &models.User{Email: "u@b.com", PasswordHash: "hash"}
				m.
					On("GetUserByEmail", mock.Anything, "u@b.com").
					Return(user, nil).
					Once()
				m.
					On("CompareCredentials", user, &domain.LoginRequest{
						Email:    "u@b.com",
						Password: "wrong",
					}).
					Return(false).
					Once()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "success",
			body: `{"email":"good@b.com","password":"right"}`,
			setupMock: func(m *MockLoginService) {
				user := &models.User{ID: 42, Email: "good@b.com", PasswordHash: ""}
				m.
					On("GetUserByEmail", mock.Anything, "good@b.com").
					Return(user, nil).
					Once()
				m.
					On("CompareCredentials", user, &domain.LoginRequest{
						Email:    "good@b.com",
						Password: "right",
					}).
					Return(true).
					Once()
			},
			wantStatus:       http.StatusOK,
			expectTokenPairs: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockLoginService)
			tc.setupMock(m)

			logger := zap.NewNop()
			h := handler.NewLoginHandler(m, logger, time.Second, jwtCfg)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tc.body))
			rr := httptest.NewRecorder()

			h.Login(rr, req)
			resp := rr.Result()
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.expectTokenPairs {
				var got domain.LoginResponse
				err := json.NewDecoder(resp.Body).Decode(&got)
				assert.NoError(t, err, "decoding response JSON")
				assert.Equal(t, "good@b.com", got.User.Email)
				assert.NotEmpty(t, got.AccessToken, "access token should be set")
				assert.NotEmpty(t, got.RefreshToken, "refresh token should be set")
			}

			m.AssertExpectations(t)
		})
	}
}
