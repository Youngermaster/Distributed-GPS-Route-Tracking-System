package database

import (
	"context"
	"fmt"
	"time"

	"data-ingestion-microservice/types"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseManager handles all database connections
type DatabaseManager struct {
	RedisClient     *redis.Client
	MongoClient     *mongo.Client
	MongoCollection *mongo.Collection
	MQTTClient      mqtt.Client
	ctx             context.Context
}

// NewDatabaseManager creates and initializes all database connections
func NewDatabaseManager(ctx context.Context, config types.Config) (*DatabaseManager, error) {
	manager := &DatabaseManager{
		ctx: ctx,
	}

	// Setup Redis connection
	if err := manager.setupRedis(config.Redis); err != nil {
		return nil, fmt.Errorf("failed to setup Redis: %w", err)
	}

	// Setup MongoDB connection
	if err := manager.setupMongoDB(ctx, config.MongoDB); err != nil {
		return nil, fmt.Errorf("failed to setup MongoDB: %w", err)
	}

	// Setup MQTT connection
	if err := manager.setupMQTT(config.MQTT); err != nil {
		return nil, fmt.Errorf("failed to setup MQTT: %w", err)
	}

	return manager, nil
}

// setupRedis initializes Redis connection
func (dm *DatabaseManager) setupRedis(config types.RedisConfig) error {
	dm.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	_, err := dm.RedisClient.Ping(dm.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// setupMongoDB initializes MongoDB connection
func (dm *DatabaseManager) setupMongoDB(ctx context.Context, config types.MongoDBConfig) error {
	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	dm.MongoClient = client
	db := client.Database(config.Database)
	dm.MongoCollection = db.Collection(config.Collection)

	return nil
}

// setupMQTT initializes MQTT connection
func (dm *DatabaseManager) setupMQTT(config types.MQTTConfig) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.Broker, config.Port))
	opts.SetClientID(config.ClientID)
	opts.SetKeepAlive(5 * time.Second)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		// Connection lost handler can be set externally if needed
	})

	dm.MQTTClient = mqtt.NewClient(opts)
	token := dm.MQTTClient.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return nil
}

// SubscribeToTopic subscribes to an MQTT topic with a message handler
func (dm *DatabaseManager) SubscribeToTopic(topic string, handler mqtt.MessageHandler) error {
	token := dm.MQTTClient.Subscribe(topic, 1, handler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to MQTT topic %s: %w", topic, token.Error())
	}
	return nil
}

// Close gracefully closes all database connections
func (dm *DatabaseManager) Close() error {
	var errs []error

	// Close MQTT connection
	if dm.MQTTClient != nil && dm.MQTTClient.IsConnected() {
		dm.MQTTClient.Disconnect(250)
	}

	// Close Redis connection
	if dm.RedisClient != nil {
		if err := dm.RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	// Close MongoDB connection
	if dm.MongoClient != nil {
		if err := dm.MongoClient.Disconnect(dm.ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to close MongoDB: %w", err))
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// IsHealthy checks if all database connections are healthy
func (dm *DatabaseManager) IsHealthy() map[string]bool {
	health := make(map[string]bool)

	// Check Redis
	_, err := dm.RedisClient.Ping(dm.ctx).Result()
	health["redis"] = err == nil

	// Check MongoDB
	err = dm.MongoClient.Ping(dm.ctx, nil)
	health["mongodb"] = err == nil

	// Check MQTT
	health["mqtt"] = dm.MQTTClient != nil && dm.MQTTClient.IsConnected()

	return health
} 