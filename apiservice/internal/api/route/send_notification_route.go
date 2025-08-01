package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewSendNotificationRoute registers the HTTP route for sending notifications.
// It sets up the necessary repository, service, and handler layers, wiring them together.
func NewSendNotificationRoute(mux *mux.Router, db domain.DBConn, logger *zap.Logger, kafkaFactory *bootstrap.KafkaFactory, topic string, timeout time.Duration, contactsPerMessage int, writerBatchTimeout time.Duration) {
	cr := repository.NewContactsRepository(db)
	tr := repository.NewTemplateRepository(db)
	kw := kafkaFactory.NewWriter(topic, bootstrap.WithBatchTimeout(writerBatchTimeout))

	sns := service.NewSendNotificationService(cr, tr, kw, contactsPerMessage)
	snh := handler.NewSendNotificationHandler(sns, logger, timeout)

	mux.HandleFunc("/send-notification/{id}", snh.SendNotification).Methods(http.MethodPost, http.MethodOptions)
}
