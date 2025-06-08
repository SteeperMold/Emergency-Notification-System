package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		correlationID   string
		handlerStatus   int
		expectGenerated bool
	}{
		{
			name:            "Uses existing correlation ID",
			correlationID:   "existing-cid-123",
			handlerStatus:   http.StatusTeapot,
			expectGenerated: false,
		},
		{
			name:            "Generates new correlation ID",
			correlationID:   "",
			handlerStatus:   http.StatusOK,
			expectGenerated: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obsCore, obsLogs := observer.New(zap.InfoLevel)
			logger := zap.New(obsCore)

			// Handler that just writes status code from test case
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.handlerStatus)
				_, _ = w.Write([]byte("response"))
			})

			mw := middleware.LoggingMiddleware(logger)(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/testuri", nil)
			if tc.correlationID != "" {
				req.Header.Set("X-Correlation-ID", tc.correlationID)
			}

			rr := httptest.NewRecorder()

			start := time.Now()
			mw.ServeHTTP(rr, req)
			duration := time.Since(start)

			gotCorrelationID := rr.Header().Get("X-Correlation-ID")
			if tc.expectGenerated {
				assert.NotEmpty(t, gotCorrelationID, "should generate correlation ID")
				assert.NotEqual(t, tc.correlationID, gotCorrelationID)
				_, err := uuid.Parse(gotCorrelationID)
				assert.NoError(t, err, "generated correlation ID should be valid UUID")
			} else {
				assert.Equal(t, tc.correlationID, gotCorrelationID, "should keep existing correlation ID")
			}

			assert.Equal(t, tc.handlerStatus, rr.Code)

			logs := obsLogs.All()

			// Expect exactly two logs: "start request" and "end request"
			assert.Len(t, logs, 2)

			startLog := logs[0]
			assert.Equal(t, "start request", startLog.Message)
			if !tc.expectGenerated {
				assert.Equal(t, tc.correlationID, startLog.ContextMap()["correlation_id"], "start log correlation_id")
			}

			endLog := logs[1]
			assert.Equal(t, "end request", endLog.Message)
			assert.Equal(t, gotCorrelationID, endLog.ContextMap()["correlation_id"], "end log correlation_id")
			assert.Equal(t, "/testuri", endLog.ContextMap()["uri"])
			assert.Equal(t, "GET", endLog.ContextMap()["method"])
			assert.Equal(t, int64(tc.handlerStatus), endLog.ContextMap()["status"])

			// duration should be positive and close to actual duration
			loggedDuration, ok := endLog.ContextMap()["duration"].(time.Duration)
			if !ok {
				raw := endLog.ContextMap()["duration"]
				switch v := raw.(type) {
				case int64:
					loggedDuration = time.Duration(v)
				case float64:
					loggedDuration = time.Duration(v)
				default:
					t.Fatal("duration not found or invalid type in log")
				}
			}
			assert.InDelta(t, duration.Nanoseconds(), loggedDuration.Nanoseconds(), float64(time.Millisecond.Nanoseconds()), "duration close to real time")
		})
	}
}
