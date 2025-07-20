package handler

import (
	"context"
	"errors"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type SendNotificationHandler struct {
	service        domain.SendNotificationService
	logger         *zap.Logger
	contextTimeout time.Duration
}

func NewSendNotificationHandler(s domain.SendNotificationService, logger *zap.Logger, timeout time.Duration) *SendNotificationHandler {
	return &SendNotificationHandler{
		service:        s,
		logger:         logger,
		contextTimeout: timeout,
	}
}

func (snh *SendNotificationHandler) logError(msg string, r *http.Request, fields ...zap.Field) {
	cid := r.Header.Get("X-Correlation-ID")

	allFields := []zap.Field{
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}
	allFields = append(allFields, fields...)

	snh.logger.Error(msg, allFields...)
}

func (snh *SendNotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), snh.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		snh.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	templateIDStr := vars["id"]
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = snh.service.SendNotification(ctx, userID, templateID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTemplateNotExists):
			http.Error(w, "template not exists", http.StatusNotFound)
		case errors.Is(err, domain.ErrContactNotExists):
			http.Error(w, "no contacts", http.StatusNotFound)
		default:
			snh.logError("failed to send notification", r, zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
