package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/tokenutils"
	"go.uber.org/zap"
)

// LoginHandler handles user login HTTP requests, token creation, and response.
type LoginHandler struct {
	service            domain.LoginService
	logger             *zap.Logger
	contextTimeout     time.Duration
	accessTokenSecret  string
	accessTokenExpiry  time.Duration
	refreshTokenSecret string
	refreshTokenExpiry time.Duration
}

// NewLoginHandler initializes and returns a new LoginHandler instance.
func NewLoginHandler(s domain.LoginService, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) *LoginHandler {
	return &LoginHandler{
		service:            s,
		logger:             logger,
		contextTimeout:     timeout,
		accessTokenSecret:  jwtConfig.AccessSecret,
		accessTokenExpiry:  jwtConfig.AccessExpiry,
		refreshTokenSecret: jwtConfig.RefreshSecret,
		refreshTokenExpiry: jwtConfig.RefreshExpiry,
	}
}

func (lh *LoginHandler) logError(msg string, r *http.Request, email string, err error) {
	cid := r.Header.Get("X-Correlation-ID")
	lh.logger.Error(msg,
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("user_email", email),
		zap.Error(err),
	)
}

// Login processes user login requests.
// It validates credentials, generates JWT tokens, and responds with user data and tokens.
func (lh *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), lh.contextTimeout)
	defer cancel()

	var request domain.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := lh.service.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotExists) {
			http.Error(w, "Invalid credentials or user doesn't exist", http.StatusUnauthorized)
		} else {
			lh.logError("failed to get user by email", r, request.Email, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if !lh.service.CompareCredentials(user, &request) {
		http.Error(w, "Invalid credentials or user does not exist", http.StatusUnauthorized)
		return
	}

	accessToken, err := tokenutils.CreateAccessToken(user, lh.accessTokenSecret, lh.accessTokenExpiry)
	if err != nil {
		lh.logError("failed to create access token", r, user.Email, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	refreshToken, err := tokenutils.CreateRefreshToken(user, lh.refreshTokenSecret, lh.refreshTokenExpiry)
	if err != nil {
		lh.logError("failed to create refresh token", r, user.Email, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := &domain.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		lh.logError("failed to write json to client", r, user.Email, err)
	}
}
