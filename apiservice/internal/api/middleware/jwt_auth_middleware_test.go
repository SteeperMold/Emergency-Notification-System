package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/middleware"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/tokenutils"
	"github.com/stretchr/testify/assert"
)

func TestJwtAuthMiddleware_Integration(t *testing.T) {
	secret := "mysecret"
	validToken, err := tokenutils.CreateAccessToken(&models.User{Email: "t@t.com", ID: 123}, secret, 10*time.Second)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectStatus   int
		expectNextCall bool
		expectUserID   any
	}{
		{
			name:           "no header",
			authHeader:     "",
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
			expectUserID:   nil,
		},
		{
			name:           "malformed",
			authHeader:     "BearerOnly",
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
			expectUserID:   nil,
		},
		{
			name:           "bad token",
			authHeader:     "Bearer invalid",
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
			expectUserID:   nil,
		},
		{
			name:           "good token",
			authHeader:     "Bearer " + validToken,
			expectStatus:   http.StatusOK,
			expectNextCall: true,
			expectUserID:   123,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			var gotUserID interface{}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID = r.Context().Value(contextkeys.UserID)
				w.WriteHeader(http.StatusOK)
			})

			h := middleware.JwtAuthMiddleware(secret)(next)
			req := httptest.NewRequest("GET", "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.expectStatus, res.StatusCode)
			assert.Equal(t, tc.expectNextCall, nextCalled)
			assert.Equal(t, tc.expectUserID, gotUserID)
		})
	}
}
