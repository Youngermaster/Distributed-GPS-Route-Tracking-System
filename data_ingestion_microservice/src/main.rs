mod config;
mod route_simplification;
mod types;

use crate::config::Config;
use crate::route_simplification::RouteSimplifier;
use crate::types::{BusMessage, BusStatus, Location};

use log::{error, info};
use mongodb::{bson::doc, Client as MongoClient};
use redis::AsyncCommands;
use rumqttc::{AsyncClient, Event, MqttOptions, Packet, QoS};
use std::time::Duration;

#[tokio::main(flavor = "multi_thread")]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize logging
    init_logging();

    info!("🚀 Starting Distributed GPS Route Tracking System - Data Ingestion Microservice");

    // Load configuration from environment variables
    let config = Config::from_env();

    // Log configuration (without sensitive data)
    info!("Configuration loaded:");
    info!(
        "  MQTT: {}:{} (topic: {})",
        config.mqtt.broker, config.mqtt.port, config.mqtt.topic
    );
    info!("  Redis: {}", config.redis.url);
    info!(
        "  MongoDB: {} (db: {})",
        config.mongodb.uri.split('@').last().unwrap_or("***"),
        config.mongodb.database
    );
    info!(
        "  Route tolerance: {}",
        config.route_simplification.tolerance
    );

    // Validate configuration
    if let Err(e) = config.validate() {
        error!("Configuration validation failed: {}", e);
        std::process::exit(1);
    }

    // Setup MQTT Client
    let mut mqtt_options = MqttOptions::new(
        config.mqtt.client_id.clone(),
        config.mqtt.broker.clone(),
        config.mqtt.port,
    );
    mqtt_options.set_keep_alive(Duration::from_secs(config.mqtt.keep_alive_secs));
    let (mqtt_client, mut eventloop) = AsyncClient::new(mqtt_options, 10);
    mqtt_client
        .subscribe(&config.mqtt.topic, QoS::AtLeastOnce)
        .await?;

    // Setup Redis connection
    let redis_client = redis::Client::open(config.redis.url.as_str())?;

    // Setup MongoDB connection
    let mongo_client = MongoClient::with_uri_str(&config.mongodb.uri).await?;
    let db = mongo_client.database(&config.mongodb.database);
    let trips_collection = db.collection(&config.mongodb.collection);

    // Setup route simplifier
    let route_simplifier = RouteSimplifier::new(config.route_simplification.tolerance)?;

    info!("Data ingestion microservice started.");

    // Process incoming MQTT events
    loop {
        let event = eventloop.poll().await?;
        match event {
            Event::Incoming(Packet::Publish(publish)) => {
                let payload = publish.payload;
                // Spawn a task to process each message concurrently
                let mut redis_conn = redis_client.get_async_connection().await?;
                let trips_collection = trips_collection.clone();
                let route_simplifier = route_simplifier.clone();
                tokio::spawn(async move {
                    if let Err(e) = process_message(
                        &payload,
                        &mut redis_conn,
                        &trips_collection,
                        &route_simplifier,
                    )
                    .await
                    {
                        error!("Error processing message: {e}");
                    }
                });
            }
            other => {
                info!("MQTT event: {:?}", other);
            }
        }
    }
}

/// Initialize logging with environment variable support
fn init_logging() {
    // Check if RUST_LOG is set, otherwise default to info level
    if std::env::var("RUST_LOG").is_err() {
        std::env::set_var("RUST_LOG", "info");
    }

    pretty_env_logger::init();
}

/// Process an incoming MQTT message payload.
/// For "in_route": store the JSON in Redis list keyed by driverId:currentRouteId.
/// For "finished": retrieve the list, simplify it, and store it in MongoDB.
async fn process_message(
    payload: &[u8],
    redis_conn: &mut redis::aio::Connection,
    trips_collection: &mongodb::Collection<mongodb::bson::Document>,
    route_simplifier: &RouteSimplifier,
) -> Result<(), Box<dyn std::error::Error>> {
    let msg: BusMessage = serde_json::from_slice(payload)?;
    let key = format!("{}:{}", msg.driver_id, msg.current_route_id);

    match msg.status {
        BusStatus::InRoute => {
            // Store the raw JSON of the location in Redis
            let loc_json = serde_json::to_string(&msg.driver_location)?;
            let _: () = redis_conn.rpush(&key, loc_json).await?;
            info!("Stored location for key {} in Redis.", key);
        }
        BusStatus::Finished => {
            // Retrieve all stored points from Redis
            let points_json: Vec<String> = redis_conn.lrange(&key, 0, -1).await?;
            if points_json.is_empty() {
                info!("No stored points for key {}.", key);
                return Ok(());
            }

            // Parse the JSON strings into Location structs
            let mut locations: Vec<Location> = Vec::new();
            for p in points_json {
                let loc: Location = serde_json::from_str(&p)?;
                locations.push(loc);
            }

            // Simplify the route using the Ramer-Douglas-Peucker algorithm
            let simplified_locations = route_simplifier.simplify_route(&locations)?;

            info!(
                "Route {} finished. Original: {} points, Simplified: {} points",
                key,
                locations.len(),
                simplified_locations.len()
            );

            // Insert the simplified route into the MongoDB trips collection.
            let trip_doc = doc! {
                "driverId": msg.driver_id,
                "currentRouteId": msg.current_route_id,
                "simplifiedRoute": simplified_locations.iter().map(|loc| {
                    doc! { "latitude": loc.latitude, "longitude": loc.longitude }
                }).collect::<Vec<_>>(),
                "timestamp": msg.timestamp as i64,
                "originalPointsCount": locations.len() as i32,
                "simplifiedPointsCount": simplified_locations.len() as i32,
            };
            trips_collection.insert_one(trip_doc, None).await?;
            info!("Stored trip for key {} in MongoDB.", key);

            // Delete the Redis key
            let _: () = redis_conn.del(&key).await?;
            info!("Cleared route data for key {} from Redis.", key);
        }
    }

    Ok(())
}
