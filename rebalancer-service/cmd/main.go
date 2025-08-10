package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/bootstrap"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/repository"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/service"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()
	defer app.CloseDBConnection()

	notificationTasksWriter := app.KafkaFactory.NewWriter(app.Config.Kafka.Topics["notification.tasks"])

	appCfg := app.Config.App

	nr := repository.NewNotificationRepository(app.DB)
	ns := service.NewRebalancerService(nr, notificationTasksWriter, app.Logger, appCfg.BatchSize, appCfg.Interval, appCfg.ContextTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("shutting down rebalancer service")
		cancel()
	}()

	go func() {
		ns.Start(ctx)
	}()

	log.Printf("listening on port %v", app.Config.App.Port)

	route.Serve(app)
}
