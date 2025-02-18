import argparse
import json
import random
import threading
import time
import paho.mqtt.client as mqtt

BROKER_HOST = "localhost"  # adjust if needed (e.g. the docker network alias for EMQX)
BROKER_PORT = 1883
TOPIC = "drivers_location/"

def simulate_bus(driver_id, route_id, status, interval=1.0, iterations=20):
    """
    Simulate a bus sending location messages periodically.
    If status is "in_route", it will continuously send updated locations.
    When status is "finished", it sends the final message with the finished flag.
    """
    client = mqtt.Client(client_id=driver_id)
    client.connect(BROKER_HOST, BROKER_PORT, keepalive=60)
    client.loop_start()
    
    # Start with some base coordinates
    lat = 40.0 + random.random()
    lon = -74.0 + random.random()
    
    for i in range(iterations):
        # Simulate some movement
        lat += random.uniform(-0.0005, 0.0005)
        lon += random.uniform(-0.0005, 0.0005)
        message = {
            "driverId": driver_id,
            "driverLocation": {
                "latitude": lat,
                "longitude": lon,
            },
            "timestamp": int(time.time() * 1000),
            "currentRouteId": route_id,
            "status": "in_route" if i < iterations - 1 or status != "finished" else "finished"
        }
        client.publish(TOPIC + driver_id, json.dumps(message))
        print(f"Sent: {message}")
        time.sleep(interval)
    
    client.loop_stop()
    client.disconnect()

def main():
    parser = argparse.ArgumentParser(
        description="Simulate bus location messages to the EMQX broker."
    )
    parser.add_argument(
        "--buses",
        type=int,
        default=1,
        help="Number of simulated buses to run concurrently."
    )
    parser.add_argument(
        "--status",
        choices=["in_route", "finished"],
        default="in_route",
        help="Initial status for each bus simulation. If 'finished', the last message will be flagged as finished."
    )
    parser.add_argument(
        "--iterations",
        type=int,
        default=20,
        help="Number of messages per bus."
    )
    args = parser.parse_args()

    threads = []
    for i in range(args.buses):
        driver_id = f"driver-{100 + i}"
        route_id = "route-123"
        t = threading.Thread(target=simulate_bus, args=(driver_id, route_id, args.status, 1.0, args.iterations))
        t.start()
        threads.append(t)

    for t in threads:
        t.join()

if __name__ == "__main__":
    main()
