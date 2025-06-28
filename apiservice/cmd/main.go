package main

import (
	"log"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()
	defer app.CloseDBConnection()

	_, cancelConsumers := consumers.ServeConsumers(app)
	defer cancelConsumers()

	log.Println("started up successfully")

	route.Serve(app)
}
