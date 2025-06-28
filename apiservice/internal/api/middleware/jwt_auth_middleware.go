package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/tokenutils"
)

// JwtAuthMiddleware returns an HTTP middleware that validates the JWT access token from the Authorization header.
// If the token is valid, it extracts the user ID from the token and stores it in the request context for downstream handlers.
// If the token is missing, malformed, invalid, or expired, it responds with HTTP 401 Unauthorized.
func JwtAuthMiddleware(accessSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHandler := r.Header.Get("Authorization")

			parts := strings.SplitN(authHandler, " ", 2)

			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			isAuthorized, err := tokenutils.IsAuthorized(token, accessSecret)
			if !isAuthorized || err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := tokenutils.ExtractIDFromToken(token, accessSecret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextkeys.UserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
