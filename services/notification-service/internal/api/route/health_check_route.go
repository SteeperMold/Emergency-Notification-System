package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewHealthCheckRoute registers the /health endpoint on the provided router.
func NewHealthCheckRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, kf domain.KafkaFactory) {
	hs := service.NewHealthCheckService(db, kf)
	hh := handler.NewHealthHandler(hs, logger, timeout)

	mux.HandleFunc("/health", hh.HealthCheck).Methods(http.MethodGet)
}
