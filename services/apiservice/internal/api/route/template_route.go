package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewTemplateRoute registers HTTP routes for managing message templates on the given mux.Router.
// Routes include GET, POST, PUT, and DELETE operations for /template and /template/{id}.
func NewTemplateRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration) {
	tr := repository.NewTemplateRepository(db)
	ts := service.NewTemplateService(tr)
	th := handler.NewTemplateHandler(ts, logger, timeout)

	mux.HandleFunc("/template", th.Get).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/template/{id}", th.GetByID).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/template", th.Post).Methods(http.MethodPost, http.MethodOptions)
	mux.HandleFunc("/template/{id}", th.Put).Methods(http.MethodPut, http.MethodOptions)
	mux.HandleFunc("/template/{id}", th.Delete).Methods(http.MethodDelete, http.MethodOptions)
}
