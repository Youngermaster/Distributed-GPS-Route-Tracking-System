package main

import (
	"os"
	"strconv"
)

// Config holds all configuration values for the application
type Config struct {
	MQTT     MQTTConfig
	Redis    RedisConfig
	MongoDB  MongoDBConfig
	RouteSimplification RouteSimplificationConfig
}

// MQTTConfig holds MQTT broker configuration
type MQTTConfig struct {
	Broker   string
	Port     int
	ClientID string
	Topic    string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	URI        string
	Database   string
	Collection string
}

// RouteSimplificationConfig holds route simplification parameters
type RouteSimplificationConfig struct {
	Tolerance float64
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() Config {
	return Config{
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnvAsInt("MQTT_PORT", 1883),
			ClientID: getEnv("MQTT_CLIENT_ID", "go_data_ingestion_client"),
			Topic:    getEnv("MQTT_TOPIC", "drivers_location/#"),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		MongoDB: MongoDBConfig{
			URI:        getEnv("MONGODB_URI", "mongodb://root:examplepassword@127.0.0.1:27017"),
			Database:   getEnv("MONGODB_DATABASE", "distributed_gps_route_tracking_system"),
			Collection: getEnv("MONGODB_COLLECTION", "trips"),
		},
		RouteSimplification: RouteSimplificationConfig{
			Tolerance: getEnvAsFloat("ROUTE_TOLERANCE", 0.0001),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsFloat gets an environment variable as float64 with a default value
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
} 