package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/api/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockTwilioCallbackService struct {
	mock.Mock
}

func (m *MockTwilioCallbackService) ProcessCallback(ctx context.Context, idStr, status string) error {
	return m.Called(ctx, idStr, status).Error(0)
}

func TestTwilioStatusCallbackHandler_ProcessCallback(t *testing.T) {
	type args struct {
		method         string
		query          string
		form           url.Values
		mockErr        error
		expectedStatus int
	}

	tests := []struct {
		name           string
		args           args
		expectCall     bool
		expectedID     string
		expectedStatus string
	}{
		{
			name: "successful callback",
			args: args{
				method:         http.MethodPost,
				query:          "?notification_id=1234",
				form:           url.Values{"MessageSid": {"abc123"}, "MessageStatus": {"delivered"}},
				mockErr:        nil,
				expectedStatus: http.StatusOK,
			},
			expectCall:     true,
			expectedID:     "1234",
			expectedStatus: "delivered",
		},
		{
			name: "missing parameters",
			args: args{
				method:         http.MethodPost,
				query:          "?notification_id=5678",
				form:           url.Values{"MessageSid": {""}, "MessageStatus": {""}},
				expectedStatus: http.StatusBadRequest,
			},
			expectCall: false,
		},
		{
			name: "malformed form",
			args: args{
				method:         http.MethodPost,
				query:          "?notification_id=5678",
				form:           nil,
				expectedStatus: http.StatusBadRequest,
			},
			expectCall: false,
		},
		{
			name: "service returns error",
			args: args{
				method:         http.MethodPost,
				query:          "?notification_id=5678",
				form:           url.Values{"MessageSid": {"xyz"}, "MessageStatus": {"failed"}},
				mockErr:        assert.AnError,
				expectedStatus: http.StatusOK, // Still 200 to prevent Twilio retry
			},
			expectCall:     true,
			expectedID:     "5678",
			expectedStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTwilioCallbackService)
			logger := zaptest.NewLogger(t)
			h := handler.NewTwilioStatusCallbackHandler(mockService, logger, 2*time.Second)

			var req *http.Request
			if tt.args.form != nil {
				formBody := bytes.NewBufferString(tt.args.form.Encode())
				req = httptest.NewRequest(tt.args.method, "/callback"+tt.args.query, formBody)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.args.method, "/callback"+tt.args.query, nil)
			}

			if tt.expectCall {
				mockService.
					On("ProcessCallback", mock.Anything, tt.expectedID, tt.expectedStatus).
					Return(tt.args.mockErr).
					Once()
			}

			rec := httptest.NewRecorder()
			h.ProcessCallback(rec, req)

			assert.Equal(t, tt.args.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}
