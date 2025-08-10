package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/api/middleware"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/bootstrap"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twilio/twilio-go/client"
)

// Serve configures and starts the HTTP server for handling Twilio callbacks.
// It applies the Twilio signature validation middleware in production,
// registers the callback route, and listens on the configured port.
func Serve(app *bootstrap.Application) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	NewHealthCheckRoute(r, app.DB, app.Logger, app.Config.App.ContextTimeout, app.KafkaFactory)

	if app.Config.App.AppEnv == "production" {
		validator := client.NewRequestValidator(app.Config.Twilio.AuthToken)
		r.Use(middleware.RequireValidTwilioSignatureMiddleware(app.Config.Twilio.StatusCallbackEndpoint, &validator))
	}

	NewTwilioCallbackRoute(r, app.DB, app.Logger, app.Config.App.MaxAttempts, app.Config.App.ContextTimeout)

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
