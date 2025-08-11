package route

import (
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewLoadContactsRoute registers the /load-contacts endpoint.
// It constructs necessary service and handler components and attaches
// the handler function to the provided mux.Router.
func NewLoadContactsRoute(mux *mux.Router, logger *zap.Logger, s3Client *s3.S3, bucket string, kafkaFactory *bootstrap.KafkaFactory, topic string, timeout time.Duration) {
	writer := kafkaFactory.NewWriter(topic)

	lcs := service.NewLoadContactsService(s3Client, bucket, writer)
	lch := handler.NewLoadContactsHandler(lcs, logger, timeout)

	mux.HandleFunc("/load-contacts", lch.LoadContactsFile).Methods(http.MethodPost, http.MethodOptions)
}
