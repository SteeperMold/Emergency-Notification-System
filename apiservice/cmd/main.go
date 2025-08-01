package main

import (
	"log"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/route"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	defer app.LoggerSync()
	defer app.CloseDBConnection()

	log.Printf("listening on port %v", app.Config.App.Port)

	route.Serve(app)
}
