package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"data-ingestion-microservice/algorithm"
	"data-ingestion-microservice/database"
	"data-ingestion-microservice/types"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
)

// DataIngestionService handles the main business logic
type DataIngestionService struct {
	config      types.Config
	dbManager   *database.DatabaseManager
	simplifier  *algorithm.RouteSimplifier
	ctx         context.Context
}

// NewDataIngestionService creates a new data ingestion service
func NewDataIngestionService(ctx context.Context, config types.Config) (*DataIngestionService, error) {
	// Initialize database manager
	dbManager, err := database.NewDatabaseManager(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database manager: %w", err)
	}

	// Initialize route simplifier
	simplifier := algorithm.NewRouteSimplifier(config.RouteSimplification.Tolerance)

	service := &DataIngestionService{
		config:     config,
		dbManager:  dbManager,
		simplifier: simplifier,
		ctx:        ctx,
	}

	// Subscribe to MQTT topic
	err = dbManager.SubscribeToTopic(config.MQTT.Topic, service.messageHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to MQTT topic: %w", err)
	}

	log.Printf("Successfully initialized data ingestion service")
	log.Printf("Subscribed to MQTT topic: %s", config.MQTT.Topic)

	return service, nil
}

// messageHandler processes incoming MQTT messages
func (s *DataIngestionService) messageHandler(client mqtt.Client, msg mqtt.Message) {
	go func() {
		if err := s.processMessage(msg.Payload()); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}()
}

// processMessage processes an incoming MQTT message payload
func (s *DataIngestionService) processMessage(payload []byte) error {
	var busMsg types.BusMessage
	if err := json.Unmarshal(payload, &busMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	key := fmt.Sprintf("%s:%s", busMsg.DriverID, busMsg.CurrentRouteID)

	switch busMsg.Status {
	case "in_route":
		return s.handleInRoute(key, busMsg.DriverLocation)
	case "finished":
		return s.handleFinished(key, busMsg)
	default:
		log.Printf("Unknown status received: %s", busMsg.Status)
		return nil
	}
}

// handleInRoute stores location data in Redis
func (s *DataIngestionService) handleInRoute(key string, location types.Location) error {
	locationJSON, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	err = s.dbManager.RedisClient.RPush(s.ctx, key, string(locationJSON)).Err()
	if err != nil {
		return fmt.Errorf("failed to store location in Redis: %w", err)
	}

	log.Printf("Stored location for key %s in Redis", key)
	return nil
}

// handleFinished retrieves route data, simplifies it, and stores in MongoDB
func (s *DataIngestionService) handleFinished(key string, busMsg types.BusMessage) error {
	// Retrieve all stored points from Redis
	pointsJSON, err := s.dbManager.RedisClient.LRange(s.ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to retrieve points from Redis: %w", err)
	}

	if len(pointsJSON) == 0 {
		log.Printf("No stored points for key %s", key)
		return nil
	}

	// Parse JSON strings into Location structs
	var locations []types.Location
	for _, pointJSON := range pointsJSON {
		var location types.Location
		if err := json.Unmarshal([]byte(pointJSON), &location); err != nil {
			log.Printf("Failed to unmarshal location: %v", err)
			continue
		}
		locations = append(locations, location)
	}

	if len(locations) == 0 {
		log.Printf("No valid locations found for key %s", key)
		return nil
	}

	// Simplify the route using the algorithm
	simplifiedLocations, err := s.simplifier.SimplifyRoute(locations)
	if err != nil {
		return fmt.Errorf("failed to simplify route: %w", err)
	}

	// Get compression statistics
	stats := s.simplifier.GetCompressionStats(locations, simplifiedLocations)

	log.Printf("Route %s finished. Original: %d points, Simplified: %d points (%.2f%% reduction)",
		key, stats.OriginalPoints, stats.SimplifiedPoints, stats.ReductionPercent)

	// Convert simplified points to MongoDB format
	var simplifiedRoute []bson.M
	for _, location := range simplifiedLocations {
		simplifiedRoute = append(simplifiedRoute, bson.M{
			"latitude":  location.Latitude,
			"longitude": location.Longitude,
		})
	}

	// Insert the simplified route into MongoDB
	tripDoc := bson.M{
		"driverId":              busMsg.DriverID,
		"currentRouteId":        busMsg.CurrentRouteID,
		"simplifiedRoute":       simplifiedRoute,
		"timestamp":             int64(busMsg.Timestamp),
		"originalPointsCount":   stats.OriginalPoints,
		"simplifiedPointsCount": stats.SimplifiedPoints,
		"compressionRatio":      stats.CompressionRatio,
		"reductionPercent":      stats.ReductionPercent,
	}

	_, err = s.dbManager.MongoCollection.InsertOne(s.ctx, tripDoc)
	if err != nil {
		return fmt.Errorf("failed to store trip in MongoDB: %w", err)
	}

	log.Printf("Stored trip for key %s in MongoDB", key)

	// Delete the Redis key
	err = s.dbManager.RedisClient.Del(s.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	log.Printf("Cleared route data for key %s from Redis", key)
	return nil
}

// GetHealthStatus returns the health status of all components
func (s *DataIngestionService) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"service":   "running",
		"databases": s.dbManager.IsHealthy(),
		"config": map[string]interface{}{
			"tolerance": s.simplifier.GetTolerance(),
			"mqtt_topic": s.config.MQTT.Topic,
		},
	}
}

// UpdateTolerance allows updating the route simplification tolerance
func (s *DataIngestionService) UpdateTolerance(newTolerance float64) {
	s.simplifier.SetTolerance(newTolerance)
	log.Printf("Updated route simplification tolerance to: %f", newTolerance)
}

// Close gracefully closes the service
func (s *DataIngestionService) Close() error {
	log.Println("Shutting down data ingestion service...")
	return s.dbManager.Close()
} 