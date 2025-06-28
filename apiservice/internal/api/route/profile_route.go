package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewProfileRoute sets up the profile route on the given mux.Router and
// registers the GET /profile endpoint with the handler.
func NewProfileRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration) {
	ur := repository.NewUserRepository(db)
	ps := service.NewProfileService(ur)
	ph := handler.NewProfileHandler(ps, logger, timeout)

	mux.HandleFunc("/profile", ph.GetProfile).Methods(http.MethodGet, http.MethodOptions)
}
