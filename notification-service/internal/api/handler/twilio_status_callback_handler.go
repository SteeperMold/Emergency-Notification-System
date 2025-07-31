package handler

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type TwilioStatusCallbackHandler struct {
	service        domain.TwilioCallbackService
	logger         *zap.Logger
	contextTimeout time.Duration
}

func NewTwilioStatusCallbackHandler(s domain.TwilioCallbackService, logger *zap.Logger, timeout time.Duration) *TwilioStatusCallbackHandler {
	return &TwilioStatusCallbackHandler{
		service:        s,
		logger:         logger,
		contextTimeout: timeout,
	}
}

func (h *TwilioStatusCallbackHandler) ProcessCallback(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.contextTimeout)
	defer cancel()

	err := r.ParseForm()
	if err != nil {
		h.logger.Info("twilio callback: invalid form", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sid := r.PostFormValue("MessageSid")
	status := r.PostFormValue("MessageStatus")
	if sid == "" || status == "" {
		h.logger.Info("twilio callback: missing field", zap.String("sid", sid), zap.String("status", status))
		http.Error(w, "missing parameters", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("notification_id")

	err = h.service.ProcessCallback(ctx, idStr, status)
	if err != nil {
		h.logger.Error("twilio callback: failed to update status", zap.String("sid", sid), zap.String("status", status), zap.Error(err))
		// we still return 200 so Twilio does not retry
		w.WriteHeader(http.StatusOK)
		return
	}

	h.logger.Info("twilio callback: status updated", zap.String("sid", sid), zap.String("status", status))
	w.WriteHeader(http.StatusOK)
}
