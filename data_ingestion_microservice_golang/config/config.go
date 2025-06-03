package config

import (
	"os"
	"strconv"

	"data-ingestion-microservice/types"
)

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() types.Config {
	return types.Config{
		MQTT: types.MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnvAsInt("MQTT_PORT", 1883),
			ClientID: getEnv("MQTT_CLIENT_ID", "go_data_ingestion_client"),
			Topic:    getEnv("MQTT_TOPIC", "drivers_location/#"),
		},
		Redis: types.RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		MongoDB: types.MongoDBConfig{
			URI:        getEnv("MONGODB_URI", "mongodb://root:examplepassword@127.0.0.1:27017"),
			Database:   getEnv("MONGODB_DATABASE", "distributed_gps_route_tracking_system"),
			Collection: getEnv("MONGODB_COLLECTION", "trips"),
		},
		RouteSimplification: types.RouteSimplificationConfig{
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