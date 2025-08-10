package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/api/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHealthCheckHandler_HealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(m *MockHealthCheckService)
		wantStatus int
	}{
		{
			name: "healthy",
			mockSetup: func(m *MockHealthCheckService) {
				m.
					On("HealthCheck", mock.Anything).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "unhealthy",
			mockSetup: func(m *MockHealthCheckService) {
				m.
					On("HealthCheck", mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := new(MockHealthCheckService)
			tc.mockSetup(mockSvc)

			logger := zap.NewNop()
			h := handler.NewHealthHandler(mockSvc, logger, 2*time.Second)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			h.HealthCheck(rr, req)

			resp := rr.Result()
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			mockSvc.AssertExpectations(t)
		})
	}
}
