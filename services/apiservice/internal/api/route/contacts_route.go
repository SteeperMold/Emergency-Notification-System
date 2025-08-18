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

// NewContactsRoute registers CRUD endpoints for managing contacts.
func NewContactsRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, paginationDefaultLimit, paginationMaxLimit int) {
	cr := repository.NewContactsRepository(db)
	cs := service.NewContactsService(cr, paginationDefaultLimit, paginationMaxLimit)
	ch := handler.NewContactsHandler(cs, logger, timeout)

	mux.HandleFunc("/contacts", ch.Get).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/contacts/{id}", ch.GetByID).Methods(http.MethodGet, http.MethodOptions)
	mux.HandleFunc("/contacts", ch.Post).Methods(http.MethodPost, http.MethodOptions)
	mux.HandleFunc("/contacts/{id}", ch.Put).Methods(http.MethodPut, http.MethodOptions)
	mux.HandleFunc("/contacts/{id}", ch.Delete).Methods(http.MethodDelete, http.MethodOptions)
}
