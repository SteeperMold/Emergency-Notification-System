package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewTwilioCallbackRoute registers the HTTP endpoint for handling Twilio status callbacks.
// It composes the repository, service, and handler layers and attaches the POST /callback route.
func NewTwilioCallbackRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, maxAttempts int, timeout time.Duration) {
	nr := repository.NewNotificationRepository(db)
	cs := service.NewTwilioCallbackService(nr, maxAttempts)
	ch := handler.NewTwilioStatusCallbackHandler(cs, logger, timeout)

	mux.HandleFunc("/callback", ch.ProcessCallback).Methods(http.MethodPost)
}
