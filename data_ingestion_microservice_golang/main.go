package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"data-ingestion-microservice/config"
	"data-ingestion-microservice/service"
)

func main() {
	// Initialize logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 Starting Distributed GPS Route Tracking System - Data Ingestion Microservice (Go)")

	// Create context for the application
	ctx := context.Background()

	// Load configuration from environment variables
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded:")
	log.Printf("  MQTT: %s:%d (topic: %s)", cfg.MQTT.Broker, cfg.MQTT.Port, cfg.MQTT.Topic)
	log.Printf("  Redis: %s", cfg.Redis.Address)
	log.Printf("  MongoDB: %s (database: %s)", cfg.MongoDB.URI, cfg.MongoDB.Database)
	log.Printf("  Route tolerance: %f", cfg.RouteSimplification.Tolerance)

	// Initialize the data ingestion service
	dataService, err := service.NewDataIngestionService(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize data ingestion service: %v", err)
	}
	defer dataService.Close()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✅ Data ingestion microservice started successfully!")
	log.Println("📊 Health status:", dataService.GetHealthStatus())
	log.Println("🔄 Processing MQTT messages... Press Ctrl+C to exit.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutdown signal received, cleaning up...")

	// Graceful shutdown
	if err := dataService.Close(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("✅ Data ingestion microservice shut down gracefully")
} 