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

// NewSignupRouter sets up the signup route on the given mux.Router and
// registers the POST /signup endpoint with the handler.
func NewSignupRouter(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) {
	ur := repository.NewUserRepository(db)
	ss := service.NewSignupService(ur)
	sh := handler.NewSignupHandler(ss, logger, timeout, jwtConfig)

	mux.HandleFunc("/signup", sh.Signup).Methods(http.MethodPost, http.MethodOptions)
}
