package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/bootstrap"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve configures and starts the HTTP server with routing and middleware.
func Serve(app *bootstrap.Application) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	NewHealthCheckRoute(r, app.DB, app.Logger, app.Config.App.ContextTimeout, app.KafkaFactory)

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
