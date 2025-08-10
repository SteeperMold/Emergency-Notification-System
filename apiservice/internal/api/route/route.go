package route

import (
	"log"
	"net/http"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/middleware"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve configures and starts the HTTP server with routing and middleware.
func Serve(app *bootstrap.Application) {
	db := app.DB
	logger := app.Logger
	timeout := app.Config.App.ContextTimeout

	r := mux.NewRouter()
	r.Use(middleware.CorsMiddleware(app.Config.App.FrontendOrigin))
	r.Use(middleware.LoggingMiddleware(logger))

	r.Handle("/metrics", promhttp.Handler())
	NewHealthCheckRoute(r, db, logger, timeout, app.KafkaFactory)

	NewSignupRouter(r, db, logger, timeout, app.Config.App.Jwt)
	NewLoginRoute(r, db, logger, timeout, app.Config.App.Jwt)
	NewRefreshTokenRoute(r, db, logger, timeout, app.Config.App.Jwt)

	private := r.NewRoute().Subrouter()
	private.Use(middleware.JwtAuthMiddleware(app.Config.App.Jwt.AccessSecret))

	NewProfileRoute(private, db, logger, timeout)
	NewTemplateRoute(private, db, logger, timeout)
	NewContactsRoute(private, db, logger, timeout)

	contactsBucket := app.Config.S3.Buckets["contacts"]
	contactsTopic := app.Config.Kafka.Topics["contacts.loading.tasks"]
	NewLoadContactsRoute(private, logger, app.S3Client, contactsBucket, app.KafkaFactory, contactsTopic, timeout)

	notificationTopic := app.Config.Kafka.Topics["notification.requests"]
	contactsPerMessage := app.Config.App.ContactsPerKafkaMessage
	writerBatchTimeout := app.Config.Kafka.NotificationRequestsBatchTimeout
	NewSendNotificationRoute(private, db, logger, app.KafkaFactory, notificationTopic, timeout, contactsPerMessage, writerBatchTimeout)

	log.Fatal(http.ListenAndServe(":"+app.Config.App.Port, r))
}
