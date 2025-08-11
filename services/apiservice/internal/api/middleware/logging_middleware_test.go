package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		initialCID string
		nextStatus int
	}{
		{
			name:       "no initial CID",
			initialCID: "",
			nextStatus: http.StatusOK,
		},
		{
			name:       "with initial CID",
			initialCID: "my-cid-123",
			nextStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// set up observer logger
			obsCore, logs := observer.New(zapcore.InfoLevel)
			logger := zap.New(obsCore)

			// next handler writes header status
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.nextStatus)
			})

			// wrap middleware
			h := middleware.LoggingMiddleware(logger)(next)

			// create request and recorder
			req := httptest.NewRequest("GET", "/foo", nil)
			if tc.initialCID != "" {
				req.Header.Set("X-Correlation-ID", tc.initialCID)
			}
			rw := httptest.NewRecorder()

			// capture start time
			h.ServeHTTP(rw, req)
			res := rw.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			// verify response header
			respCID := res.Header.Get("X-Correlation-ID")
			if tc.initialCID != "" {
				assert.Equal(t, tc.initialCID, respCID)
			} else {
				// should be valid UUID
				assert.Regexp(t, regexp.MustCompile(`^[0-9a-fA-F\-]{36}$`), respCID)
			}

			// expect two log entries
			all := logs.All()
			assert.Len(t, all, 2)

			// start entry validations
			start := all[0]
			assert.Equal(t, "start request", start.Message)
			cm := start.ContextMap()
			assert.Equal(t, respCID, cm["correlation_id"])
			assert.Equal(t, "GET", cm["method"])
			assert.Equal(t, "/foo", cm["uri"])

			// end entry validations
			end := all[1]
			assert.Equal(t, "end request", end.Message)
			cm2 := end.ContextMap()
			assert.Equal(t, respCID, cm2["correlation_id"])
			assert.Equal(t, "GET", cm2["method"])
			assert.Equal(t, "/foo", cm2["uri"])
			assert.Equal(t, int64(tc.nextStatus), cm2["status"])

			// duration should be a time.Duration > 0
			dur, ok := cm2["duration"].(time.Duration)
			assert.True(t, ok)
			assert.True(t, dur >= 0)
		})
	}
}
