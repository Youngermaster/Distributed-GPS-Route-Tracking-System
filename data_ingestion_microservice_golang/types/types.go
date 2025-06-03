package types

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

// Config holds all configuration values for the application
type Config struct {
	MQTT                MQTTConfig
	Redis               RedisConfig
	MongoDB             MongoDBConfig
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