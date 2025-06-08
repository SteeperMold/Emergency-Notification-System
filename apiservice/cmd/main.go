package main

import (
	"log"

	"github.com/SteeperMold/Emergency-Notification-System/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()
	defer app.CloseDBConnection()

	route.Serve(app)

	log.Println("started up successfully")
}
