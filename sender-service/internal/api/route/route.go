package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/bootstrap"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve configures and starts the HTTP server with routing and middleware.
func Serve(app *bootstrap.Application) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
