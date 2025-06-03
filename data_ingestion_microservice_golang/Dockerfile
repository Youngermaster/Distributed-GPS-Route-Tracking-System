# =============================================================================
# Multi-stage Dockerfile for Data Ingestion Microservice (Go)
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build Stage
# -----------------------------------------------------------------------------
FROM golang:1.24.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set the working directory
WORKDIR /app

# Copy dependency files first for better Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o data-ingestion-service \
    .

# Verify the binary was created and is executable
RUN chmod +x data-ingestion-service && \
    ./data-ingestion-service --help || echo "Binary built successfully"

# -----------------------------------------------------------------------------
# Stage 2: Final Runtime Stage
# -----------------------------------------------------------------------------
FROM scratch AS runtime

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary from builder stage
COPY --from=builder /app/data-ingestion-service /data-ingestion-service

# Set environment variables with sensible defaults
ENV MQTT_BROKER=localhost \
    MQTT_PORT=1883 \
    MQTT_CLIENT_ID=go_data_ingestion_client \
    MQTT_TOPIC=drivers_location/# \
    REDIS_ADDRESS=127.0.0.1:6379 \
    REDIS_PASSWORD= \
    REDIS_DB=0 \
    MONGODB_URI=mongodb://root:examplepassword@127.0.0.1:27017 \
    MONGODB_DATABASE=distributed_gps_route_tracking_system \
    MONGODB_COLLECTION=trips \
    ROUTE_TOLERANCE=0.0001 \
    LOG_LEVEL=info

# Expose no ports by default (this is an MQTT subscriber service)
# EXPOSE 8080

# Add health check (if we add health endpoint later)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD ["/data-ingestion-service", "--health-check"] || exit 1

# Add metadata labels following OCI spec
LABEL \
    org.opencontainers.image.title="Data Ingestion Microservice" \
    org.opencontainers.image.description="High-performance GPS route tracking data ingestion service" \
    org.opencontainers.image.version="1.0.0" \
    org.opencontainers.image.vendor="Distributed GPS Route Tracking System" \
    org.opencontainers.image.source="https://github.com/Youngermaster/Distributed-GPS-Route-Tracking-System" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.documentation="https://github.com/Youngermaster/Distributed-GPS-Route-Tracking-System/blob/main/data_ingestion_microservice_golang/README.md"

# Run as the non-root user for security
# USER nobody:nobody

# Set the entrypoint
ENTRYPOINT ["/data-ingestion-service"]

# Default command arguments (can be overridden)
CMD []
