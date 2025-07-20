package main

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()

	kafkaCfg := app.Config.Kafka
	notificationTasksReader := app.KafkaFactory.NewReader(kafkaCfg.Topics["notification.tasks"], kafkaCfg.ConsumerGroup)

	ntr := repository.NewNotificationTasksRepository(app.DB)
	nts := service.NewNotificationTasksService(ntr, app.SmsSender, app.Config.App.MaxAttempts)
	ntc := consumers.NewNotificationTasksConsumer(nts, notificationTasksReader, app.Logger, app.Config.App.ContextTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Printf("shutting down sender service")
		cancel()
	}()

	log.Fatal(ntc.StartConsumer(ctx))
}
