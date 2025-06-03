# Data Ingestion Microservice (Go)

A high-performance, modular data ingestion microservice for GPS route tracking built with Go. This service subscribes to MQTT messages containing GPS location data, performs real-time route simplification using the Ramer-Douglas-Peucker algorithm, and stores optimized routes in MongoDB.

## 🏗️ Architecture

The microservice follows a clean, modular architecture with separated concerns:

```bash
data_ingestion_microservice_golang/
├── main.go                              # Application entry point
├── config/                              # Configuration management
│   └── config.go                        # Environment variable loading
├── types/                               # Data structures and types
│   └── types.go                         # Common types (Location, BusMessage, Config)
├── algorithm/                           # Route simplification algorithms
│   ├── simplification.go                # Douglas-Peucker implementation
│   └── simplification_test.go           # Algorithm tests and benchmarks
├── database/                            # Database connection management
│   └── connections.go                   # Redis, MongoDB, MQTT managers
├── service/                             # Business logic
│   └── ingestion_service.go             # Main service implementation
├── go.mod                               # Go module definition
├── go.sum                               # Dependency checksums
├── Dockerfile                           # Multi-stage Docker build
├── .dockerignore                        # Docker build context optimization
├── Makefile                             # Build and development commands
└── README.md                            # This file
```

## ✨ Features

- **Modular Design**: Clean separation of concerns with dedicated packages
- **Real-time Processing**: Concurrent MQTT message processing
- **Route Optimization**: Advanced Douglas-Peucker algorithm for GPS route simplification
- **Database Integration**: Redis for temporary storage, MongoDB for persistent data
- **Health Monitoring**: Built-in health checks for all components
- **Graceful Shutdown**: Proper cleanup of all connections
- **Comprehensive Testing**: Unit tests and benchmarks for algorithms
- **Configuration Management**: Environment variable-based configuration
- **Performance Metrics**: Route compression statistics and monitoring
- **Docker Ready**: Multi-stage Docker build for minimal production images

## 🐳 Docker Deployment

### Multi-stage Dockerfile

The service uses a sophisticated multi-stage Docker build:

```dockerfile
# Stage 1: Builder - Go compilation environment
FROM golang:1.24.1-alpine AS builder
# ... compile binary with optimizations

# Stage 2: Runtime - Minimal scratch-based image
FROM scratch AS runtime
# ... only binary + essential files
```

### Key Docker Features

- **Minimal Size**: Final image is only **~10.7MB**
- **Security**: Scratch-based runtime with no shell or package manager
- **Performance**: Statically linked binary with no dependencies
- **Optimization**: Multi-layer caching for faster rebuilds

### Docker Commands

```bash
# Build production image
make docker-build

# Build development image
make docker-build-dev

# Run with external services (requires Redis/MongoDB/MQTT)
make docker-run

# Run with Docker Compose services
make docker-run-with-services

# Inspect image details and layers
make docker-inspect

# Test the Docker image
make docker-test

# Clean Docker artifacts
make docker-clean
```

### Environment Variables

The Docker image supports all configuration via environment variables:

```bash
docker run --rm -it \
  -e MQTT_BROKER=localhost \
  -e MQTT_PORT=1883 \
  -e REDIS_ADDRESS=127.0.0.1:6379 \
  -e MONGODB_URI=mongodb://root:password@localhost:27017 \
  -e ROUTE_TOLERANCE=0.0001 \
  data-ingestion-service:latest
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Redis server
- MongoDB server
- MQTT broker (EMQX recommended)
- Docker (optional)

### Local Development

```bash
# Clone the repository
git clone <repository-url>
cd data_ingestion_microservice_golang

# Install dependencies
go mod tidy

# Start infrastructure services
make dev-start

# Run the service
make run
```

### Docker Deployment

```bash
# Build the Docker image
make docker-build

# Start infrastructure and run containerized service
make docker-run-with-services
```

### Using Docker Compose

The service integrates with the existing Docker Compose setup:

```bash
# Start all infrastructure services
make dev-start

# In another terminal, run the Go service
make run
# or
make docker-run
```

## ⚙️ Configuration

Configure the service using environment variables:

```bash
# MQTT Configuration
export MQTT_BROKER="localhost"
export MQTT_PORT="1883"
export MQTT_CLIENT_ID="go_data_ingestion_client"
export MQTT_TOPIC="drivers_location/#"

# Redis Configuration
export REDIS_ADDRESS="127.0.0.1:6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"

# MongoDB Configuration
export MONGODB_URI="mongodb://root:examplepassword@127.0.0.1:27017"
export MONGODB_DATABASE="distributed_gps_route_tracking_system"
export MONGODB_COLLECTION="trips"

# Route Simplification
export ROUTE_TOLERANCE="0.0001"
```

## 📡 Message Processing

The service processes MQTT messages with the following structure:

### Input Message Format

```json
{
  "driverId": "driver_001",
  "driverLocation": {
    "latitude": 40.7128,
    "longitude": -74.006
  },
  "timestamp": 1640995200000,
  "currentRouteId": "route_123",
  "status": "in_route" // or "finished"
}
```

### Processing Flow

1. **In Route**: GPS points are stored in Redis using the key pattern `{driverId}:{currentRouteId}`
2. **Route Finished**: All stored points are retrieved, simplified using Douglas-Peucker algorithm, and saved to MongoDB
3. **Cleanup**: Temporary data is removed from Redis

### Output Data (MongoDB)

```json
{
  "driverId": "driver_001",
  "currentRouteId": "route_123",
  "simplifiedRoute": [
    { "latitude": 40.7128, "longitude": -74.006 },
    { "latitude": 40.758, "longitude": -73.9855 }
  ],
  "timestamp": 1640995200000,
  "originalPointsCount": 150,
  "simplifiedPointsCount": 12,
  "compressionRatio": 0.08,
  "reductionPercent": 92.0
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./algorithm/

# Test specific package
go test ./algorithm/
```

### Benchmark Results

The Douglas-Peucker implementation is optimized for performance:

```bash
BenchmarkSimplifyRoute_100Points-8       10000    120000 ns/op
BenchmarkSimplifyRoute_1000Points-8       1000   1200000 ns/op
```

## 🛠️ Development

### Available Make Commands

```bash
make help           # Show available commands
make build          # Build the application
make run            # Run the application
make test           # Run tests
make clean          # Clean build artifacts
make dev-start      # Start development services
make dev-stop       # Stop development services
make lint           # Run linter (if available)
```

### Code Organization

- **`config/`**: Environment variable parsing and validation
- **`types/`**: Shared data structures and types
- **`algorithm/`**: Route simplification algorithms with comprehensive tests
- **`database/`**: Database connection management and health checks
- **`service/`**: Main business logic and message processing
- **`main.go`**: Application bootstrap and graceful shutdown

## 📊 Monitoring and Health Checks

The service provides health check endpoints and metrics:

```go
// Get health status of all components
status := service.GetHealthStatus()

// Returns:
{
  "service": "running",
  "databases": {
    "redis": true,
    "mongodb": true,
    "mqtt": true
  },
  "config": {
    "tolerance": 0.0001,
    "mqtt_topic": "drivers_location/#"
  }
}
```

## 🎯 Algorithm Details

### Douglas-Peucker Route Simplification

The service uses a custom implementation of the Ramer-Douglas-Peucker algorithm:

- **Purpose**: Reduces GPS route complexity while preserving shape
- **Method**: Recursively removes points below distance threshold
- **Performance**: O(n log n) average case, optimized for GPS data
- **Quality**: Configurable tolerance for different use cases

### Compression Statistics

Track route optimization effectiveness:

- **Original Points**: Number of GPS points received
- **Simplified Points**: Number of points after simplification
- **Compression Ratio**: Simplified/Original ratio
- **Reduction Percent**: Percentage of points removed

## 🔄 Comparison with Rust Implementation

This Go implementation offers several improvements over the original Rust version:

### Advantages

- **Modular Architecture**: Clean separation of concerns
- **Better Error Handling**: Comprehensive error types and handling
- **Enhanced Monitoring**: Built-in health checks and metrics
- **Flexible Configuration**: Environment-based configuration
- **Comprehensive Testing**: Unit tests and benchmarks
- **Documentation**: Extensive code documentation
- **Docker Optimization**: Multi-stage builds for production deployment

### Performance

Both implementations offer similar performance for the core algorithm, with the Go version providing better observability and maintainability.

## 📝 Contributing

1. Follow Go conventions and use `gofmt`
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass before submitting

## 📄 License

This project is licensed under the Apache License 2.0.
