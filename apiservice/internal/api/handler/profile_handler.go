package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"go.uber.org/zap"
)

// ProfileHandler handles HTTP requests related to user profile retrieval.
type ProfileHandler struct {
	service        domain.ProfileService
	logger         *zap.Logger
	contextTimeout time.Duration
}

// NewProfileHandler creates and returns a new ProfileHandler instance.
func NewProfileHandler(s domain.ProfileService, logger *zap.Logger, contextTimeout time.Duration) *ProfileHandler {
	return &ProfileHandler{
		service:        s,
		logger:         logger,
		contextTimeout: contextTimeout,
	}
}

func (ph *ProfileHandler) logError(msg string, r *http.Request, fields ...zap.Field) {
	cid := r.Header.Get("X-Correlation-ID")

	allFields := []zap.Field{
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}
	allFields = append(allFields, fields...)

	ph.logger.Error(msg, allFields...)
}

// GetProfile handles HTTP GET requests to fetch the authenticated user's profile.
// It reads the user ID from the request context, fetches the profile via the service,
// and writes the profile as JSON in the response.
func (ph *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ph.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)

	userID, ok := rawUserID.(int)
	if !ok {
		ph.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	profile, err := ph.service.GetUserByID(ctx, userID)
	if err != nil {
		ph.logError("internal server error", r, zap.Int("user_id", userID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&profile)
	if err != nil {
		ph.logError("failed to write json to client", r, zap.Int("user_id", userID), zap.Error(err))
	}
}
