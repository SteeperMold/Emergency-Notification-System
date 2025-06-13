package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewLoginRoute sets up the login route on the given mux.Router and
// registers the POST /login endpoint with the handler.
func NewLoginRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) {
	ur := repository.NewUserRepository(db)
	ls := service.NewLoginService(ur)
	lh := handler.NewLoginHandler(ls, logger, timeout, jwtConfig)

	mux.HandleFunc("/login", lh.Login).Methods(http.MethodPost, http.MethodOptions)
}
