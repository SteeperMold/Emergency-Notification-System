package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/service"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()
	defer app.CloseDBConnection()

	kafkaCfg := app.Config.Kafka
	notificationRequestsReader := app.KafkaFactory.NewReader(kafkaCfg.Topics["notification.requests"], kafkaCfg.ConsumerGroup)
	sendTasksWriter := app.KafkaFactory.NewWriter(kafkaCfg.Topics["notification.tasks"], bootstrap.WithBatchTimeout(kafkaCfg.NotificationTasksWriterBatchTimeout))

	nr := repository.NewNotificationRepository(app.DB)
	appCfg := app.Config.App
	nrs := service.NewNotificationRequestsService(nr, sendTasksWriter, appCfg.NotificationTasksWriterBatchSize)
	nrc := consumers.NewNotificationRequestsConsumer(nrs, notificationRequestsReader, app.Logger, appCfg.ContextTimeout, appCfg.NotificationConsumerBatchSize, appCfg.NotificationConsumerFlushInterval)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("shutting down consumer")
		cancel()
	}()

	go func() {
		log.Fatal(nrc.StartConsumer(ctx))
	}()

	log.Printf("listening on port %v", app.Config.App.Port)

	route.Serve(app)
}
