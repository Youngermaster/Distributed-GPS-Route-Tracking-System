defmodule DataIngestionMicroservice.Application do
  @moduledoc false

  use Application

  def start(_type, _args) do
    children = [
      # Start the Redis connection (assumes "redis" is reachable on port 6379)
      {Redix, name: :redix, host: "localhost", port: 6379},
      # Start the MQTT handler (using EMQTT)
      DataIngestionMicroservice.MQTTHandler
    ]

    opts = [strategy: :one_for_one, name: DataIngestionMicroservice.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
