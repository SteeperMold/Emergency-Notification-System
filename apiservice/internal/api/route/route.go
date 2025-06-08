package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/middleware"
	"github.com/SteeperMold/Emergency-Notification-System/internal/bootstrap"
	"github.com/gorilla/mux"
)

// Serve configures and starts the HTTP server with routing and middleware.
func Serve(app *bootstrap.Application) {
	db := app.DB
	logger := app.Logger
	timeout := app.Config.App.ContextTimeout

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))

	public := r.NewRoute().Subrouter()

	NewSignupRouter(public, db, logger, timeout, &app.Config.App.Jwt)
	NewLoginRoute(public, db, logger, timeout, &app.Config.App.Jwt)
	NewRefreshTokenRoute(public, db, logger, timeout, &app.Config.App.Jwt)

	private := r.NewRoute().Subrouter()
	private.Use(middleware.JwtAuthMiddleware(app.Config.App.Jwt.AccessSecret))

	NewProfileRoute(private, db, logger, timeout)

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
