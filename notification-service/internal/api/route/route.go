package route

import (
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/api/middleware"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/bootstrap"
	"github.com/gorilla/mux"
	"github.com/twilio/twilio-go/client"
	"log"
	"net/http"
)

func Serve(app *bootstrap.Application) {
	r := mux.NewRouter()

	if app.Config.App.AppEnv == "production" {
		validator := client.NewRequestValidator(app.Config.Twilio.AuthToken)
		r.Use(middleware.RequireValidTwilioSignatureMiddleware(app.Config.Twilio.StatusCallbackEndpoint, &validator))
	}

	NewTwilioCallbackRoute(r, app.DB, app.Logger, app.Config.App.MaxAttempts, app.Config.App.ContextTimeout)

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
