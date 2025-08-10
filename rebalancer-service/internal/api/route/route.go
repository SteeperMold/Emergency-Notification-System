package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/bootstrap"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve configures and starts the HTTP server with routing and middleware.
func Serve(app *bootstrap.Application) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, mux))
}
