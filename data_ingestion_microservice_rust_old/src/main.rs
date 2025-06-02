use futures::StreamExt;
use log::{error, info};
use mongodb::{bson::doc, Client as MongoClient};
use redis::AsyncCommands;
use rumqttc::{AsyncClient, Event, MqttOptions, Packet, QoS};
use serde::{Deserialize, Serialize};
use std::error::Error;
use std::time::Duration;
use tokio::time;

#[derive(Debug, Deserialize)]
struct BusMessage {
    driverId: String,
    driverLocation: Location,
    timestamp: u64,
    currentRouteId: String,
    status: String, // "in_route" or "finished"
}

#[derive(Debug, Deserialize, Serialize, Clone)]
struct Location {
    latitude: f64,
    longitude: f64,
}

#[tokio::main(flavor = "multi_thread")]
async fn main() -> Result<(), Box<dyn Error>> {
    // Initialize logging (using pretty_env_logger)
    pretty_env_logger::init();

    // --- Setup MQTT Client ---
    let mut mqttoptions = MqttOptions::new("rust_client", "localhost", 1883);
    mqttoptions.set_keep_alive(Duration::from_secs(5));
    let (mqtt_client, mut eventloop) = AsyncClient::new(mqttoptions, 10);
    mqtt_client
        .subscribe("drivers_location/#", QoS::AtLeastOnce)
        .await?;

    // --- Setup Redis connection ---
    let redis_client = redis::Client::open("redis://127.0.0.1/")?;
    let mut redis_conn = redis_client.get_async_connection().await?;

    // --- Setup MongoDB connection (optional) ---
    // For this example, we'll assume a local MongoDB instance; adjust the URI as needed.
    let mongo_client =
        MongoClient::with_uri_str("mongodb://root:examplepassword@127.0.0.1:27017").await?;
    let db = mongo_client.database("distributed_gps_route_tracking_system");
    let trips_collection = db.collection("trips");

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
                tokio::spawn(async move {
                    if let Err(e) =
                        process_message(&payload, &mut redis_conn, &trips_collection).await
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

/// Process an incoming MQTT message payload.
/// For "in_route": store the JSON in Redis list keyed by driverId:currentRouteId.
/// For "finished": retrieve the list, simplify it, and store it (here we log it and insert into MongoDB).
async fn process_message(
    payload: &[u8],
    redis_conn: &mut redis::aio::Connection,
    trips_collection: &mongodb::Collection<mongodb::bson::Document>,
) -> Result<(), Box<dyn Error>> {
    let msg: BusMessage = serde_json::from_slice(payload)?;
    let key = format!("{}:{}", msg.driverId, msg.currentRouteId);

    match msg.status.as_str() {
        "in_route" => {
            // Store the raw JSON of the location in Redis
            let loc_json = serde_json::to_string(&msg.driverLocation)?;
            let _: () = redis_conn.rpush(&key, loc_json).await?;
            info!("Stored location for key {} in Redis.", key);
        }
        "finished" => {
            // Retrieve all stored points from Redis
            let points_json: Vec<String> = redis_conn.lrange(&key, 0, -1).await?;
            if points_json.is_empty() {
                info!("No stored points for key {}.", key);
                return Ok(());
            }

            // Parse the JSON strings into Location structs
            let mut points: Vec<geo::Point<f64>> = Vec::new();
            for p in points_json {
                let loc: Location = serde_json::from_str(&p)?;
                points.push(geo::Point::new(loc.longitude, loc.latitude));
            }

            // Simplify the route using the Ramer-Douglas-Peucker algorithm
            // The simplify function requires a slice of points and a tolerance value.
            let tolerance = 0.0001; // adjust tolerance as needed
            let linestring = geo::LineString::from(points);
            let simplified = geo::algorithm::simplify::Simplify::simplify(&linestring, &tolerance);
            info!(
                "Route {} finished. Simplified points: {:?}",
                key, simplified
            );

            // Insert the simplified route into the MongoDB trips collection.
            let trip_doc = doc! {
                "driverId": msg.driverId,
                "currentRouteId": msg.currentRouteId,
                "simplifiedRoute": simplified.0.iter().map(|p| {
                    doc! { "latitude": p.y, "longitude": p.x }
                }).collect::<Vec<_>>(),
                "timestamp": msg.timestamp as i64,
            };
            trips_collection.insert_one(trip_doc).await?;
            info!("Stored trip for key {} in MongoDB.", key);

            // Delete the Redis key
            let _: () = redis_conn.del(&key).await?;
            info!("Cleared route data for key {} from Redis.", key);
        }
        other => {
            info!("Unknown status received: {}", other);
        }
    }

    Ok(())
}
