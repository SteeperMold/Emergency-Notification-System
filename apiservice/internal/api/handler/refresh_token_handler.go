package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/tokenutils"
	"go.uber.org/zap"
)

// RefreshTokenHandler handles refresh token HTTP requests.
type RefreshTokenHandler struct {
	service            domain.RefreshTokenService
	logger             *zap.Logger
	contextTimeout     time.Duration
	accessTokenSecret  string
	accessTokenExpiry  time.Duration
	refreshTokenSecret string
	refreshTokenExpiry time.Duration
}

// NewRefreshTokenHandler initializes a new RefreshTokenHandler with dependencies.
func NewRefreshTokenHandler(s domain.RefreshTokenService, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		service:            s,
		logger:             logger,
		contextTimeout:     timeout,
		accessTokenSecret:  jwtConfig.AccessSecret,
		accessTokenExpiry:  jwtConfig.AccessExpiry,
		refreshTokenSecret: jwtConfig.RefreshSecret,
		refreshTokenExpiry: jwtConfig.RefreshExpiry,
	}
}

func (rth *RefreshTokenHandler) logError(msg string, r *http.Request, userID int, err error) {
	cid := r.Header.Get("X-Correlation-ID")
	rth.logger.Error(msg,
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Int("user_id", userID),
		zap.Error(err),
	)
}

// RefreshToken processes refresh token requests.
// It validates the refresh token, fetches the user, and issues new JWT tokens.
func (rth *RefreshTokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), rth.contextTimeout)
	defer cancel()

	var request domain.RefreshTokenRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	id, err := tokenutils.ExtractIDFromToken(request.RefreshToken, rth.refreshTokenSecret)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	user, err := rth.service.GetUserByID(ctx, id)
	if err != nil {
		rth.logError("internal server error", r, id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := tokenutils.CreateAccessToken(user, rth.accessTokenSecret, rth.accessTokenExpiry)
	if err != nil {
		rth.logError("failed to create access token", r, id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := tokenutils.CreateRefreshToken(user, rth.refreshTokenSecret, rth.refreshTokenExpiry)
	if err != nil {
		rth.logError("failed to create refresh token", r, id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshTokenResponse := &domain.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(refreshTokenResponse)
	if err != nil {
		rth.logError("failed to write json to client", r, id, err)
	}
}
