package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewHealthCheckRoute registers the /health endpoint on the provided router.
func NewHealthCheckRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, kf domain.KafkaFactory) {
	hs := service.NewHealthCheckService(db, kf)
	hh := handler.NewHealthHandler(hs, logger, timeout)

	mux.HandleFunc("/health", hh.HealthCheck).Methods(http.MethodGet)
}
