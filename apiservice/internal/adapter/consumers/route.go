package consumers

import (
	"context"
	"log"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
)

// ServeConsumers configures and starts all background Kafka consumers for the application.
// It launches each consumer loop in its own goroutine and returns a cancellable Context
// that can be used to shut them down gracefully.
func ServeConsumers(app *bootstrap.Application) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	topics := app.Config.Kafka.Topics
	cgs := app.Config.Kafka.ConsumerGroups

	contactsReader := app.KafkaFactory.NewReader(topics["contacts.loading.results"], cgs["contacts.loading.results"])
	cr := repository.NewContactsRepository(app.DB)
	svc := service.NewContactsService(cr)

	consumer := NewContactsKafkaConsumer(svc, contactsReader, app.Logger)

	go func() {
		log.Fatal(consumer.StartConsumer(ctx))
	}()

	return ctx, cancel
}
