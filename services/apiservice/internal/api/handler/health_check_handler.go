package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"go.uber.org/zap"
)

// HealthCheckHandler handles HTTP requests for system health checks.
type HealthCheckHandler struct {
	service        domain.HealthCheckService
	logger         *zap.Logger
	contextTimeout time.Duration
}

// NewHealthHandler creates and returns a new HealthCheckHandler.
func NewHealthHandler(s domain.HealthCheckService, logger *zap.Logger, timeout time.Duration) *HealthCheckHandler {
	return &HealthCheckHandler{
		service:        s,
		logger:         logger,
		contextTimeout: timeout,
	}
}

// HealthCheck processes an HTTP GET request to verify the application's health.
func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.contextTimeout)
	defer cancel()

	err := h.service.HealthCheck(ctx)
	if err != nil {
		http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
