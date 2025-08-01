package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/service"
)

func main() {
	app := bootstrap.NewApp()
	defer app.CloseDBConnection()
	defer app.LoggerSync()

	kafkaCfg := app.Config.Kafka
	contactsReader := app.KafkaFactory.NewReader(kafkaCfg.Topics["contacts.loading.tasks"], kafkaCfg.ConsumerGroup)

	cr := repository.NewContactsRepository(app.DB)
	cs := service.NewContactsService(cr, app.S3Client, app.Config.S3.Bucket, app.Config.App.ContextTimeout, app.Config.App.BatchSize)
	cc := consumers.NewContactsConsumer(cs, contactsReader, app.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("shutting down consumer")
		cancel()
	}()

	log.Println("contacts worker started")

	log.Fatal(cc.StartConsumer(ctx))
}
