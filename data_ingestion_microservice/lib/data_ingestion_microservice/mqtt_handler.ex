defmodule DataIngestionMicroservice.MQTTHandler do
  @moduledoc """
  A Tortoise MQTT handler that processes incoming bus location messages.
  """
  @behaviour Tortoise.Handler
  require Logger

  @impl true
  def init(args), do: args

  @impl true
  def connection(:up, state) do
    Logger.info("Connected to MQTT broker.")
    {:ok, state}
  end

  @impl true
  def connection(:down, state) do
    Logger.warn("Disconnected from MQTT broker!")
    {:ok, state}
  end

  @impl true
  def handle_message(topic, payload, state) do
    Logger.info("Received message on #{topic}")

    with {:ok, message} <- Jason.decode(payload) do
      process_message(message)
    else
      error ->
        Logger.error("Failed to decode message: #{inspect(error)}")
    end

    {:ok, state}
  end

  @impl true
  def terminate(reason, state) do
    Logger.info("Terminating MQTT Handler: #{inspect(reason)}")
    :ok
  end

  @impl true
  def code_change(_old_vsn, state, _extra), do: {:ok, state}

  defp process_message(%{
         "driverId" => driver_id,
         "status" => status,
         "currentRouteId" => route_id,
         "driverLocation" => location
       } = _message) do
    key = "#{driver_id}:#{route_id}"

    case status do
      "in_route" ->
        # Store each location in a Redis list
        {:ok, _} = Redix.command(:redix, ["RPUSH", key, Jason.encode!(location)])
        Logger.info("Stored location for #{key} in Redis.")

      "finished" ->
        # Retrieve all stored locations from Redis for this route
        case Redix.command(:redix, ["LRANGE", key, "0", "-1"]) do
          {:ok, points_json} ->
            points =
              points_json
              |> Enum.map(fn p -> Jason.decode!(p) end)
            # Apply the Ramer-Douglas-Peucker algorithm via the simplify library
            simplified_points = Simplify.simplify(points, 0.0001)
            Logger.info("Route #{key} finished. Simplified points: #{inspect(simplified_points)}")
            # Here you would typically store `simplified_points` into your MongoDB trips collection.
            # After processing, remove the key from Redis:
            {:ok, _} = Redix.command(:redix, ["DEL", key])
          {:error, reason} ->
            Logger.error("Failed to retrieve points from Redis: #{inspect(reason)}")
        end

      _ ->
        Logger.warn("Unknown status received: #{status}")
    end
  end

  defp process_message(_message) do
    Logger.warn("Invalid message format received.")
  end
end
