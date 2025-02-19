defmodule DataIngestionMicroservice.MQTTHandler do
  use GenServer
  require Logger

  @topic "drivers_location/#"

  def start_link(_args) do
    GenServer.start_link(__MODULE__, %{}, name: __MODULE__)
  end

  def init(state) do
    emqtt_opts = [
      # string, not charlist
      host: "localhost",
      port: 1883,
      clientid: "data_ingestion_client",
      clean_start: true,
      name: :emqtt
    ]

    case :emqtt.start_link(emqtt_opts) do
      {:ok, pid} ->
        :ok = :emqtt.connect(pid)
        :ok = :emqtt.subscribe(pid, {@topic, 0})
        Logger.info("Subscribed to #{@topic} with EMQTT.")

        {:ok, Map.put(state, :emqtt_pid, pid)}

      {:error, reason} ->
        Logger.error("Failed to start emqtt: #{inspect(reason)}")
        {:stop, reason}
    end
  end

  def handle_info({:publish, publish}, state = %{emqtt_pid: pid}) do
    Logger.info("Received message on #{publish.topic}")

    case Jason.decode(publish.payload) do
      {:ok, message} ->
        process_message(message)

      error ->
        Logger.error("Failed to decode payload: #{inspect(error)}")
    end

    {:noreply, state}
  end

  def handle_info(msg, state) do
    Logger.debug("Unhandled message: #{inspect(msg)}")
    {:noreply, state}
  end

  # Process the bus message
  defp process_message(
         %{
           "driverId" => driver_id,
           "status" => status,
           "currentRouteId" => route_id,
           "driverLocation" => location
         } = _message
       ) do
    key = "#{driver_id}:#{route_id}"

    case status do
      "in_route" ->
        # Append the location to a Redis list
        case Redix.command(:redix, ["RPUSH", key, Jason.encode!(location)]) do
          {:ok, _} ->
            Logger.info("Stored location for #{key} in Redis.")

          {:error, err} ->
            Logger.error("Error storing location in Redis: #{inspect(err)}")
        end

      "finished" ->
        # Retrieve all stored points from Redis
        case Redix.command(:redix, ["LRANGE", key, "0", "-1"]) do
          {:ok, points_json} ->
            points =
              points_json
              |> Enum.map(fn p -> Jason.decode!(p) end)

            # Simplify the route using the Ramer-Douglas-Peucker algorithm (tolerance value is configurable)
            simplified_points = Simplify.simplify(points, 0.0001)
            Logger.info("Route #{key} finished. Simplified points: #{inspect(simplified_points)}")

            # Here you would store the simplified_points into your MongoDB trips collection.
            # After processing, remove the key from Redis.
            case Redix.command(:redix, ["DEL", key]) do
              {:ok, _} -> Logger.info("Cleared route data for #{key} from Redis.")
              {:error, err} -> Logger.error("Error deleting Redis key: #{inspect(err)}")
            end

          {:error, reason} ->
            Logger.error("Failed to retrieve points from Redis: #{inspect(reason)}")
        end

      _ ->
        Logger.warning("Unknown status received: #{status}")
    end
  end

  defp process_message(_message) do
    Logger.warning("Invalid message format received.")
  end
end
