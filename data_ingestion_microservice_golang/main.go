package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BusMessage represents the incoming MQTT message structure
type BusMessage struct {
	DriverID        string   `json:"driverId"`
	DriverLocation  Location `json:"driverLocation"`
	Timestamp       uint64   `json:"timestamp"`
	CurrentRouteID  string   `json:"currentRouteId"`
	Status          string   `json:"status"` // "in_route" or "finished"
}

// Location represents GPS coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Point represents a simplified point for route simplification
type Point struct {
	X, Y float64
}

// DataIngestionService holds all service dependencies
type DataIngestionService struct {
	config           Config
	mqttClient       mqtt.Client
	redisClient      *redis.Client
	mongoClient      *mongo.Client
	tripsCollection  *mongo.Collection
	ctx              context.Context
}

func main() {
	// Initialize logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Data Ingestion Microservice")

	// Load configuration
	config := LoadConfig()
	log.Printf("Loaded configuration: MQTT=%s:%d, Redis=%s, MongoDB=%s", 
		config.MQTT.Broker, config.MQTT.Port, config.Redis.Address, config.MongoDB.Database)

	ctx := context.Background()

	// Initialize service
	service, err := NewDataIngestionService(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}
	defer service.Close()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Data ingestion microservice started. Press Ctrl+C to exit.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
}

// NewDataIngestionService creates and initializes a new service instance
func NewDataIngestionService(ctx context.Context, config Config) (*DataIngestionService, error) {
	service := &DataIngestionService{
		config: config,
		ctx:    ctx,
	}

	// Setup Redis connection
	if err := service.setupRedis(); err != nil {
		return nil, fmt.Errorf("failed to setup Redis: %w", err)
	}

	// Setup MongoDB connection
	if err := service.setupMongoDB(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup MongoDB: %w", err)
	}

	// Setup MQTT connection
	if err := service.setupMQTT(); err != nil {
		return nil, fmt.Errorf("failed to setup MQTT: %w", err)
	}

	return service, nil
}

// setupRedis initializes Redis connection
func (s *DataIngestionService) setupRedis() error {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.config.Redis.Address,
		Password: s.config.Redis.Password,
		DB:       s.config.Redis.DB,
	})

	// Test connection
	_, err := s.redisClient.Ping(s.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis")
	return nil
}

// setupMongoDB initializes MongoDB connection
func (s *DataIngestionService) setupMongoDB(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(s.config.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	s.mongoClient = client
	db := client.Database(s.config.MongoDB.Database)
	s.tripsCollection = db.Collection(s.config.MongoDB.Collection)

	log.Println("Connected to MongoDB")
	return nil
}

// setupMQTT initializes MQTT connection and subscription
func (s *DataIngestionService) setupMQTT() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", s.config.MQTT.Broker, s.config.MQTT.Port))
	opts.SetClientID(s.config.MQTT.ClientID)
	opts.SetKeepAlive(5 * time.Second)
	opts.SetDefaultPublishHandler(s.messageHandler)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	s.mqttClient = mqtt.NewClient(opts)
	token := s.mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	// Subscribe to drivers location topic
	token = s.mqttClient.Subscribe(s.config.MQTT.Topic, 1, s.messageHandler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to MQTT topic: %w", token.Error())
	}

	log.Printf("Connected to MQTT broker and subscribed to %s", s.config.MQTT.Topic)
	return nil
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
	var busMsg BusMessage
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
func (s *DataIngestionService) handleInRoute(key string, location Location) error {
	locationJSON, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	err = s.redisClient.RPush(s.ctx, key, string(locationJSON)).Err()
	if err != nil {
		return fmt.Errorf("failed to store location in Redis: %w", err)
	}

	log.Printf("Stored location for key %s in Redis", key)
	return nil
}

// handleFinished retrieves route data, simplifies it, and stores in MongoDB
func (s *DataIngestionService) handleFinished(key string, busMsg BusMessage) error {
	// Retrieve all stored points from Redis
	pointsJSON, err := s.redisClient.LRange(s.ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to retrieve points from Redis: %w", err)
	}

	if len(pointsJSON) == 0 {
		log.Printf("No stored points for key %s", key)
		return nil
	}

	// Parse JSON strings into Location structs
	var points []Point
	for _, pointJSON := range pointsJSON {
		var location Location
		if err := json.Unmarshal([]byte(pointJSON), &location); err != nil {
			log.Printf("Failed to unmarshal location: %v", err)
			continue
		}
		points = append(points, Point{X: location.Longitude, Y: location.Latitude})
	}

	// Simplify the route using Ramer-Douglas-Peucker algorithm
	simplifiedPoints := simplifyRoute(points, s.config.RouteSimplification.Tolerance)

	log.Printf("Route %s finished. Original points: %d, Simplified points: %d", key, len(points), len(simplifiedPoints))

	// Convert simplified points back to Location format
	var simplifiedRoute []bson.M
	for _, point := range simplifiedPoints {
		simplifiedRoute = append(simplifiedRoute, bson.M{
			"latitude":  point.Y,
			"longitude": point.X,
		})
	}

	// Insert the simplified route into MongoDB
	tripDoc := bson.M{
		"driverId":        busMsg.DriverID,
		"currentRouteId":  busMsg.CurrentRouteID,
		"simplifiedRoute": simplifiedRoute,
		"timestamp":       int64(busMsg.Timestamp),
	}

	_, err = s.tripsCollection.InsertOne(s.ctx, tripDoc)
	if err != nil {
		return fmt.Errorf("failed to store trip in MongoDB: %w", err)
	}

	log.Printf("Stored trip for key %s in MongoDB", key)

	// Delete the Redis key
	err = s.redisClient.Del(s.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	log.Printf("Cleared route data for key %s from Redis", key)
	return nil
}

// Close gracefully closes all connections
func (s *DataIngestionService) Close() {
	if s.mqttClient != nil && s.mqttClient.IsConnected() {
		s.mqttClient.Disconnect(250)
		log.Println("Disconnected from MQTT broker")
	}

	if s.redisClient != nil {
		s.redisClient.Close()
		log.Println("Disconnected from Redis")
	}

	if s.mongoClient != nil {
		if err := s.mongoClient.Disconnect(s.ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Disconnected from MongoDB")
		}
	}
}

// simplifyRoute implements the Ramer-Douglas-Peucker algorithm for route simplification
func simplifyRoute(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// Find the point with the maximum distance from the line segment
	// defined by the first and last points
	maxDistance := 0.0
	maxIndex := 0
	start := points[0]
	end := points[len(points)-1]

	for i := 1; i < len(points)-1; i++ {
		distance := perpendicularDistance(points[i], start, end)
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	// If the maximum distance is greater than tolerance, recursively simplify
	if maxDistance > tolerance {
		// Recursive call on the first part
		firstPart := simplifyRoute(points[:maxIndex+1], tolerance)
		// Recursive call on the second part
		secondPart := simplifyRoute(points[maxIndex:], tolerance)

		// Combine results (remove duplicate point at the junction)
		result := make([]Point, len(firstPart)+len(secondPart)-1)
		copy(result, firstPart)
		copy(result[len(firstPart):], secondPart[1:])
		return result
	}

	// If no point is farther than tolerance, return only start and end points
	return []Point{start, end}
}

// perpendicularDistance calculates the perpendicular distance from a point to a line segment
func perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	// Calculate the area of the triangle formed by the three points
	// using the cross product, then divide by the length of the base
	A := lineStart.X*(lineEnd.Y-point.Y) + lineEnd.X*(point.Y-lineStart.Y) + point.X*(lineStart.Y-lineEnd.Y)
	if A < 0 {
		A = -A
	}

	// Calculate the length of the base (line segment)
	B := distance(lineStart, lineEnd)

	if B == 0 {
		return distance(point, lineStart)
	}

	return A / B
}

// distance calculates the Euclidean distance between two points
func distance(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
} 