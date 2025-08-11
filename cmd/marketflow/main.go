package main

import (
	"marketflow/internal/app"
	"marketflow/pkg/logger"
)

func main() {
	srv, cleanup := app.SetupApp()
	defer cleanup()

	app.StartServer(srv)

	app.WaitForShutdown()

	app.ShutdownServer(srv)

	logger.Info("App is closed...")
}
