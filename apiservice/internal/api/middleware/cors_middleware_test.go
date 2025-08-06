package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCorsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigin  string
		method         string
		expectStatus   int
		expectNextCall bool
	}{
		{
			name:           "OPTIONS preflight with origin",
			allowedOrigin:  "https://example.com",
			method:         http.MethodOptions,
			expectStatus:   http.StatusNoContent,
			expectNextCall: false,
		},
		{
			name:           "GET with origin",
			allowedOrigin:  "https://example.com",
			method:         http.MethodGet,
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "POST without origin",
			allowedOrigin:  "",
			method:         http.MethodPost,
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			h := middleware.CorsMiddleware(tc.allowedOrigin)(next)

			req := httptest.NewRequest(tc.method, "/test", nil)
			rw := httptest.NewRecorder()

			h.ServeHTTP(rw, req)
			res := rw.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.expectStatus, res.StatusCode)

			assert.Equal(t, tc.expectNextCall, nextCalled)

			origin := res.Header.Get("Access-Control-Allow-Origin")
			if tc.allowedOrigin != "" {
				assert.Equal(t, tc.allowedOrigin, origin)
			} else {
				assert.Empty(t, origin)
			}
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", res.Header.Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "Content-Type, Authorization", res.Header.Get("Access-Control-Allow-Headers"))
			assert.Equal(t, "true", res.Header.Get("Access-Control-Allow-Credentials"))
		})
	}
}
