package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestSendNotificationHandler(t *testing.T) {
	userID := 1
	validIDStr := "123"
	validID := 123

	tests := []struct {
		name           string
		templateID     string
		userInContext  any
		mockSetup      func(m *MockSendNotificationService)
		expectedStatus int
	}{
		{
			name:          "success",
			templateID:    validIDStr,
			userInContext: userID,
			mockSetup: func(m *MockSendNotificationService) {
				m.
					On("SendNotification", mock.Anything, userID, validID).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid user ID type",
			templateID:     validIDStr,
			userInContext:  "not-an-int",
			mockSetup:      func(m *MockSendNotificationService) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid template ID",
			templateID:     "abc",
			userInContext:  userID,
			mockSetup:      func(m *MockSendNotificationService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "template not exists",
			templateID:    validIDStr,
			userInContext: userID,
			mockSetup: func(m *MockSendNotificationService) {
				m.
					On("SendNotification", mock.Anything, userID, validID).
					Return(domain.ErrTemplateNotExists).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "no contacts",
			templateID:    validIDStr,
			userInContext: userID,
			mockSetup: func(m *MockSendNotificationService) {
				m.
					On("SendNotification", mock.Anything, userID, validID).
					Return(domain.ErrContactNotExists).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "internal error",
			templateID:    validIDStr,
			userInContext: userID,
			mockSetup: func(m *MockSendNotificationService) {
				m.
					On("SendNotification", mock.Anything, userID, validID).
					Return(assert.AnError).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(MockSendNotificationService)
			tt.mockSetup(m)

			h := handler.NewSendNotificationHandler(m, zaptest.NewLogger(t), time.Second)

			r := httptest.NewRequest(http.MethodPost, "/send-notification/"+tt.templateID, nil)
			r = mux.SetURLVars(r, map[string]string{"id": tt.templateID})
			ctx := context.WithValue(r.Context(), contextkeys.UserID, tt.userInContext)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h.SendNotification(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			m.AssertExpectations(t)
		})
	}
}
