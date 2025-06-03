package main

import (
	"testing"

	"data-ingestion-microservice/algorithm"
	"data-ingestion-microservice/config"
	"data-ingestion-microservice/types"
)

func TestConfigLoading(t *testing.T) {
	// Test default configuration loading
	cfg := config.LoadConfig()
	
	// Test MQTT defaults
	if cfg.MQTT.Broker != "localhost" {
		t.Errorf("Expected MQTT broker 'localhost', got '%s'", cfg.MQTT.Broker)
	}
	if cfg.MQTT.Port != 1883 {
		t.Errorf("Expected MQTT port 1883, got %d", cfg.MQTT.Port)
	}
	if cfg.MQTT.ClientID != "go_data_ingestion_client" {
		t.Errorf("Expected MQTT client ID 'go_data_ingestion_client', got '%s'", cfg.MQTT.ClientID)
	}
	if cfg.MQTT.Topic != "drivers_location/#" {
		t.Errorf("Expected MQTT topic 'drivers_location/#', got '%s'", cfg.MQTT.Topic)
	}

	// Test Redis defaults
	if cfg.Redis.Address != "127.0.0.1:6379" {
		t.Errorf("Expected Redis address '127.0.0.1:6379', got '%s'", cfg.Redis.Address)
	}
	if cfg.Redis.Password != "" {
		t.Errorf("Expected empty Redis password, got '%s'", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Expected Redis DB 0, got %d", cfg.Redis.DB)
	}

	// Test MongoDB defaults
	expectedMongoURI := "mongodb://root:examplepassword@127.0.0.1:27017"
	if cfg.MongoDB.URI != expectedMongoURI {
		t.Errorf("Expected MongoDB URI '%s', got '%s'", expectedMongoURI, cfg.MongoDB.URI)
	}
	if cfg.MongoDB.Database != "distributed_gps_route_tracking_system" {
		t.Errorf("Expected MongoDB database 'distributed_gps_route_tracking_system', got '%s'", cfg.MongoDB.Database)
	}
	if cfg.MongoDB.Collection != "trips" {
		t.Errorf("Expected MongoDB collection 'trips', got '%s'", cfg.MongoDB.Collection)
	}

	// Test Route Simplification defaults
	if cfg.RouteSimplification.Tolerance != 0.0001 {
		t.Errorf("Expected route tolerance 0.0001, got %f", cfg.RouteSimplification.Tolerance)
	}
}

func TestAlgorithmIntegration(t *testing.T) {
	simplifier := algorithm.NewRouteSimplifier(0.001)
	
	// Test with a simple route
	locations := []types.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 1.0, Longitude: 1.0},
		{Latitude: 2.0, Longitude: 2.0},
		{Latitude: 3.0, Longitude: 3.0},
		{Latitude: 4.0, Longitude: 4.0},
	}
	
	simplified, err := simplifier.SimplifyRoute(locations)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// A straight line should be simplified to start and end points
	if len(simplified) < 2 {
		t.Errorf("Expected at least 2 points, got %d", len(simplified))
	}
	
	// Test compression stats
	stats := simplifier.GetCompressionStats(locations, simplified)
	if stats.OriginalPoints != 5 {
		t.Errorf("Expected 5 original points, got %d", stats.OriginalPoints)
	}
	if stats.SimplifiedPoints != len(simplified) {
		t.Errorf("Expected %d simplified points, got %d", len(simplified), stats.SimplifiedPoints)
	}
}

func TestTypeStructures(t *testing.T) {
	// Test BusMessage structure
	busMsg := types.BusMessage{
		DriverID: "driver_001",
		DriverLocation: types.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
		Timestamp:      1640995200000,
		CurrentRouteID: "route_123",
		Status:         "in_route",
	}
	
	if busMsg.DriverID != "driver_001" {
		t.Errorf("Expected driver ID 'driver_001', got '%s'", busMsg.DriverID)
	}
	if busMsg.DriverLocation.Latitude != 40.7128 {
		t.Errorf("Expected latitude 40.7128, got %f", busMsg.DriverLocation.Latitude)
	}
	if busMsg.Status != "in_route" {
		t.Errorf("Expected status 'in_route', got '%s'", busMsg.Status)
	}
}

func TestAlgorithmTolerance(t *testing.T) {
	simplifier := algorithm.NewRouteSimplifier(0.001)
	
	// Test initial tolerance
	if simplifier.GetTolerance() != 0.001 {
		t.Errorf("Expected tolerance 0.001, got %f", simplifier.GetTolerance())
	}
	
	// Test tolerance update
	simplifier.SetTolerance(0.005)
	if simplifier.GetTolerance() != 0.005 {
		t.Errorf("Expected tolerance 0.005, got %f", simplifier.GetTolerance())
	}
}

// Benchmark test for the new algorithm structure
func BenchmarkRouteSimplification(b *testing.B) {
	simplifier := algorithm.NewRouteSimplifier(0.001)
	
	// Generate 100 points for testing
	locations := make([]types.Location, 100)
	for i := 0; i < 100; i++ {
		locations[i] = types.Location{
			Latitude:  float64(i) * 0.001,
			Longitude: float64(i) * 0.001,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := simplifier.SimplifyRoute(locations)
		if err != nil {
			b.Fatalf("Error in simplification: %v", err)
		}
	}
}

func BenchmarkCompressionStats(b *testing.B) {
	simplifier := algorithm.NewRouteSimplifier(0.001)
	
	original := make([]types.Location, 100)
	simplified := make([]types.Location, 10)
	
	for i := 0; i < 100; i++ {
		original[i] = types.Location{Latitude: float64(i), Longitude: float64(i)}
	}
	for i := 0; i < 10; i++ {
		simplified[i] = types.Location{Latitude: float64(i * 10), Longitude: float64(i * 10)}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = simplifier.GetCompressionStats(original, simplified)
	}
} 