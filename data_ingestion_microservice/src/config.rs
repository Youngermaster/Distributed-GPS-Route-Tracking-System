use serde::Deserialize;
use std::env;

/// Configuration structure for the data ingestion microservice
#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub mqtt: MqttConfig,
    pub redis: RedisConfig,
    pub mongodb: MongoDbConfig,
    pub route_simplification: RouteSimplificationConfig,
    pub logging: LoggingConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct MqttConfig {
    pub broker: String,
    pub port: u16,
    pub client_id: String,
    pub topic: String,
    pub keep_alive_secs: u64,
    pub qos: u8,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RedisConfig {
    pub url: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct MongoDbConfig {
    pub uri: String,
    pub database: String,
    pub collection: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RouteSimplificationConfig {
    pub tolerance: f64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct LoggingConfig {
    pub level: String,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            mqtt: MqttConfig::default(),
            redis: RedisConfig::default(),
            mongodb: MongoDbConfig::default(),
            route_simplification: RouteSimplificationConfig::default(),
            logging: LoggingConfig::default(),
        }
    }
}

impl Default for MqttConfig {
    fn default() -> Self {
        Self {
            broker: "localhost".to_string(),
            port: 1883,
            client_id: "rust_data_ingestion_client".to_string(),
            topic: "drivers_location/#".to_string(),
            keep_alive_secs: 5,
            qos: 1,
        }
    }
}

impl Default for RedisConfig {
    fn default() -> Self {
        Self {
            url: "redis://127.0.0.1:6379".to_string(),
        }
    }
}

impl Default for MongoDbConfig {
    fn default() -> Self {
        Self {
            uri: "mongodb://root:examplepassword@127.0.0.1:27017".to_string(),
            database: "distributed_gps_route_tracking_system".to_string(),
            collection: "trips".to_string(),
        }
    }
}

impl Default for RouteSimplificationConfig {
    fn default() -> Self {
        Self { tolerance: 0.0001 }
    }
}

impl Default for LoggingConfig {
    fn default() -> Self {
        Self {
            level: "info".to_string(),
        }
    }
}

impl Config {
    /// Load configuration from environment variables with fallback to defaults
    pub fn from_env() -> Self {
        Self {
            mqtt: MqttConfig {
                broker: get_env("MQTT_BROKER", "localhost"),
                port: get_env_as::<u16>("MQTT_PORT", 1883),
                client_id: get_env("MQTT_CLIENT_ID", "rust_data_ingestion_client"),
                topic: get_env("MQTT_TOPIC", "drivers_location/#"),
                keep_alive_secs: get_env_as::<u64>("MQTT_KEEP_ALIVE_SECS", 5),
                qos: get_env_as::<u8>("MQTT_QOS", 1),
            },
            redis: RedisConfig {
                url: get_env("REDIS_URL", "redis://127.0.0.1:6379"),
            },
            mongodb: MongoDbConfig {
                uri: get_env(
                    "MONGODB_URI",
                    "mongodb://root:examplepassword@127.0.0.1:27017",
                ),
                database: get_env("MONGODB_DATABASE", "distributed_gps_route_tracking_system"),
                collection: get_env("MONGODB_COLLECTION", "trips"),
            },
            route_simplification: RouteSimplificationConfig {
                tolerance: get_env_as::<f64>("ROUTE_TOLERANCE", 0.0001),
            },
            logging: LoggingConfig {
                level: get_env("LOG_LEVEL", "info"),
            },
        }
    }

    /// Validate the configuration
    pub fn validate(&self) -> Result<(), String> {
        if self.mqtt.broker.is_empty() {
            return Err("MQTT broker cannot be empty".to_string());
        }
        if self.mqtt.port == 0 {
            return Err("MQTT port must be greater than 0".to_string());
        }
        if self.redis.url.is_empty() {
            return Err("Redis URL cannot be empty".to_string());
        }
        if self.mongodb.uri.is_empty() {
            return Err("MongoDB URI cannot be empty".to_string());
        }
        if self.route_simplification.tolerance <= 0.0 {
            return Err("Route tolerance must be greater than 0".to_string());
        }

        Ok(())
    }
}

/// Helper function to get environment variable with default value
fn get_env(key: &str, default: &str) -> String {
    env::var(key).unwrap_or_else(|_| default.to_string())
}

/// Helper function to get environment variable as specific type with default value
fn get_env_as<T>(key: &str, default: T) -> T
where
    T: std::str::FromStr + Clone,
{
    env::var(key)
        .ok()
        .and_then(|val| val.parse().ok())
        .unwrap_or(default)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config_defaults() {
        let config = Config::default();

        assert_eq!(config.mqtt.broker, "localhost");
        assert_eq!(config.mqtt.port, 1883);
        assert_eq!(config.redis.url, "redis://127.0.0.1:6379");
        assert_eq!(config.route_simplification.tolerance, 0.0001);
    }

    #[test]
    fn test_config_validation_success() {
        let config = Config::default();
        assert!(config.validate().is_ok());
    }

    #[test]
    fn test_config_validation_failure() {
        let mut config = Config::default();
        config.mqtt.broker = "".to_string();
        assert!(config.validate().is_err());

        config = Config::default();
        config.route_simplification.tolerance = -1.0;
        assert!(config.validate().is_err());
    }
}
