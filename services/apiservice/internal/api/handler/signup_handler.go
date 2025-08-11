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

// SignupHandler handles user signup HTTP requests.
type SignupHandler struct {
	service            domain.SignupService
	logger             *zap.Logger
	contextTimeout     time.Duration
	accessTokenSecret  string
	accessTokenExpiry  time.Duration
	refreshTokenSecret string
	refreshTokenExpiry time.Duration
}

// NewSignupHandler creates a new SignupHandler with dependencies injected.
func NewSignupHandler(s domain.SignupService, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) *SignupHandler {
	return &SignupHandler{
		service:            s,
		logger:             logger,
		contextTimeout:     timeout,
		accessTokenSecret:  jwtConfig.AccessSecret,
		accessTokenExpiry:  jwtConfig.AccessExpiry,
		refreshTokenSecret: jwtConfig.RefreshSecret,
		refreshTokenExpiry: jwtConfig.RefreshExpiry,
	}
}

func (sh *SignupHandler) logError(msg string, r *http.Request, email string, err error) {
	cid := r.Header.Get("X-Correlation-ID")
	sh.logger.Error(msg,
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("user_email", email),
		zap.Error(err),
	)
}

// Signup handles the HTTP POST request for user signup.
// It reads the JSON body, validates and creates the user via the service layer,
// generates JWT access and refresh tokens, and responds with user info and tokens.
func (sh *SignupHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), sh.contextTimeout)
	defer cancel()

	var request domain.SignupRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	newUser, err := sh.service.CreateUser(ctx, &request)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			http.Error(w, "invalid email", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrInvalidPassword):
			http.Error(w, "invalid password", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			http.Error(w, "email already exists", http.StatusConflict)
		default:
			sh.logError("internal server error", r, request.Email, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	accessToken, err := tokenutils.CreateAccessToken(newUser, sh.accessTokenSecret, sh.accessTokenExpiry)
	if err != nil {
		sh.logError("failed to create access token", r, newUser.Email, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := tokenutils.CreateRefreshToken(newUser, sh.refreshTokenSecret, sh.refreshTokenExpiry)
	if err != nil {
		sh.logError("failed to create refresh token", r, newUser.Email, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := &domain.SignupResponse{
		User:         newUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sh.logError("failed to write json to client", r, newUser.Email, err)
	}
}
