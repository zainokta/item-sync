package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/zainokta/item-sync/internal/infrastructure/server"
)

func main() {
	app, err := server.NewApplication()
	if err != nil {
		log.Fatal("failed to create application: ", err)
	}

	// Start the server in a goroutine
	go func() {
		if err := app.Start(); err != nil {
			app.GetLogger().Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown gracefully
	app.Shutdown()
}
