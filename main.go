package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/zainokta/item-sync/internal/infrastructure/server"

	_ "github.com/joho/godotenv/autoload"
)

// @title Item Sync Service API
// @version 1.0.0
// @description A robust Go service for synchronizing external API data with automatic background jobs, retry logic, and idempotent operations.
//
// @contact.name API Support
// @contact.email support@itemsync.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @schemes http https
//
// @tag.name health
// @tag.description Health check endpoints
//
// @tag.name items
// @tag.description Item management endpoints
//
// @tag.name sync
// @tag.description Data synchronization endpoints
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
