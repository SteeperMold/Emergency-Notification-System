package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRefreshTokenRoute sets up the refresh token route on the given mux.Router and
// registers the POST /refresh-token endpoint with the handler.
func NewRefreshTokenRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, timeout time.Duration, jwtConfig *bootstrap.JWTConfig) {
	ur := repository.NewUserRepository(db)
	rts := service.NewRefreshTokenService(ur)
	rth := handler.NewRefreshTokenHandler(rts, logger, timeout, jwtConfig)

	mux.HandleFunc("/refresh-token", rth.RefreshToken).Methods(http.MethodPost, http.MethodOptions)
}
