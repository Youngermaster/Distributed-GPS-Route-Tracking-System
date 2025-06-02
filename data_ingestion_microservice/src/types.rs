use serde::{Deserialize, Serialize};
use std::fmt;

/// Represents an incoming MQTT message from a bus/driver
#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct BusMessage {
    pub driver_id: String,
    pub driver_location: Location,
    pub timestamp: u64,
    pub current_route_id: String,
    pub status: BusStatus,
}

/// Represents a GPS location
#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
pub struct Location {
    pub latitude: f64,
    pub longitude: f64,
}

/// Status of a bus in its route
#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum BusStatus {
    InRoute,
    Finished,
}

impl fmt::Display for BusStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            BusStatus::InRoute => write!(f, "in_route"),
            BusStatus::Finished => write!(f, "finished"),
        }
    }
}

impl std::str::FromStr for BusStatus {
    type Err = ServiceError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "in_route" => Ok(BusStatus::InRoute),
            "finished" => Ok(BusStatus::Finished),
            _ => Err(ServiceError::InvalidStatus(s.to_string())),
        }
    }
}

/// Trip document structure for MongoDB storage
#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct TripDocument {
    pub driver_id: String,
    pub current_route_id: String,
    pub simplified_route: Vec<Location>,
    pub timestamp: i64,
    pub original_points_count: usize,
    pub simplified_points_count: usize,
    pub compression_ratio: f64,
}

impl TripDocument {
    pub fn new(
        driver_id: String,
        current_route_id: String,
        simplified_route: Vec<Location>,
        timestamp: i64,
        original_count: usize,
    ) -> Self {
        let simplified_count = simplified_route.len();
        let compression_ratio = if original_count > 0 {
            (simplified_count as f64) / (original_count as f64)
        } else {
            0.0
        };

        Self {
            driver_id,
            current_route_id,
            simplified_route,
            timestamp,
            original_points_count: original_count,
            simplified_points_count: simplified_count,
            compression_ratio,
        }
    }
}

/// Custom error types for the service
#[derive(Debug, thiserror::Error)]
pub enum ServiceError {
    #[error("MQTT error: {0}")]
    Mqtt(#[from] rumqttc::ClientError),

    #[error("Redis error: {0}")]
    Redis(#[from] redis::RedisError),

    #[error("MongoDB error: {0}")]
    MongoDB(#[from] mongodb::error::Error),

    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),

    #[error("Configuration error: {0}")]
    Config(String),

    #[error("Invalid status: {0}")]
    InvalidStatus(String),

    #[error("Route processing error: {0}")]
    RouteProcessing(String),

    #[error("Connection error: {0}")]
    Connection(String),

    #[error("Validation error: {0}")]
    Validation(String),
}

/// Type alias for Results using our custom error type
pub type ServiceResult<T> = Result<T, ServiceError>;

/// Metrics structure for monitoring
#[derive(Debug, Clone, Default)]
pub struct ServiceMetrics {
    pub messages_processed: u64,
    pub routes_in_progress: u64,
    pub routes_completed: u64,
    pub errors_count: u64,
    pub total_points_processed: u64,
    pub total_points_simplified: u64,
}

impl ServiceMetrics {
    pub fn increment_messages_processed(&mut self) {
        self.messages_processed += 1;
    }

    pub fn increment_routes_in_progress(&mut self) {
        self.routes_in_progress += 1;
    }

    pub fn decrement_routes_in_progress(&mut self) {
        if self.routes_in_progress > 0 {
            self.routes_in_progress -= 1;
        }
    }

    pub fn increment_routes_completed(&mut self) {
        self.routes_completed += 1;
    }

    pub fn increment_errors(&mut self) {
        self.errors_count += 1;
    }

    pub fn add_points_processed(&mut self, count: u64) {
        self.total_points_processed += count;
    }

    pub fn add_points_simplified(&mut self, count: u64) {
        self.total_points_simplified += count;
    }

    pub fn compression_ratio(&self) -> f64 {
        if self.total_points_processed > 0 {
            self.total_points_simplified as f64 / self.total_points_processed as f64
        } else {
            0.0
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bus_status_parsing() {
        assert_eq!("in_route".parse::<BusStatus>().unwrap(), BusStatus::InRoute);
        assert_eq!(
            "finished".parse::<BusStatus>().unwrap(),
            BusStatus::Finished
        );
        assert!("invalid".parse::<BusStatus>().is_err());
    }

    #[test]
    fn test_trip_document_creation() {
        let route = vec![
            Location {
                latitude: 1.0,
                longitude: 2.0,
            },
            Location {
                latitude: 3.0,
                longitude: 4.0,
            },
        ];

        let trip = TripDocument::new(
            "driver1".to_string(),
            "route1".to_string(),
            route,
            1234567890,
            10,
        );

        assert_eq!(trip.original_points_count, 10);
        assert_eq!(trip.simplified_points_count, 2);
        assert_eq!(trip.compression_ratio, 0.2);
    }

    #[test]
    fn test_metrics() {
        let mut metrics = ServiceMetrics::default();

        metrics.increment_messages_processed();
        metrics.add_points_processed(100);
        metrics.add_points_simplified(20);

        assert_eq!(metrics.messages_processed, 1);
        assert_eq!(metrics.compression_ratio(), 0.2);
    }
}
