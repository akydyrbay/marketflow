package main

import (
	"marketflow/internal/app"
)

func main() {
	srv, cleanup := app.SetupApp()
	defer cleanup()

	app.StartServer(srv)

	app.WaitForShutdown()

	app.ShutdownServer(srv)
}
