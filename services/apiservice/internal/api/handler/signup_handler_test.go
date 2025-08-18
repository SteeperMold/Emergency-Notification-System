package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestSignupHandler_Signup(t *testing.T) {
	validReq := domain.SignupRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}
	validUser := &models.User{
		ID:    1,
		Email: validReq.Email,
	}

	tests := []struct {
		name           string
		body           any
		setupMock      func(m *MockSignupService)
		expectedStatus int
	}{
		{
			name: "success",
			body: validReq,
			setupMock: func(m *MockSignupService) {
				m.
					On("CreateUser", mock.Anything, &validReq).
					Return(validUser, nil).
					Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			body: validReq,
			setupMock: func(m *MockSignupService) {
				m.
					On("CreateUser", mock.Anything, &validReq).
					Return((*models.User)(nil), domain.ErrInvalidEmail).
					Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid password",
			body: validReq,
			setupMock: func(m *MockSignupService) {
				m.
					On("CreateUser", mock.Anything, &validReq).
					Return((*models.User)(nil), domain.ErrInvalidPassword).
					Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "email already exists",
			body: validReq,
			setupMock: func(m *MockSignupService) {
				m.
					On("CreateUser", mock.Anything, &validReq).
					Return((*models.User)(nil), domain.ErrEmailAlreadyExists).
					Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal error on create",
			body: validReq,
			setupMock: func(m *MockSignupService) {
				m.
					On("CreateUser", mock.Anything, &validReq).
					Return((*models.User)(nil), assert.AnError).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid json body",
			body:           "not-json",
			setupMock:      func(m *MockSignupService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(MockSignupService)
			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			jwtCfg := &bootstrap.JWTConfig{
				AccessSecret:  "accesssecret",
				RefreshSecret: "refreshsecret",
				AccessExpiry:  time.Minute,
				RefreshExpiry: time.Hour,
			}

			h := handler.NewSignupHandler(m, zaptest.NewLogger(t), time.Second*2, jwtCfg)

			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.Signup(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			m.AssertExpectations(t)
		})
	}
}
