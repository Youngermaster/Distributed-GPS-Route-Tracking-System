defmodule DataIngestionMicroservice.Application do
  @moduledoc false

  use Application

  def start(_type, _args) do
    children = [
      # Start the Redis connection (assumes "redis" is reachable on port 6379)
      {Redix, name: :redix, host: "redis", port: 6379},
      # Start the Tortoise MQTT client
      {Tortoise.Connection,
       [
         client_id: "data_ingestion_client",
         server: {Tortoise.Transport.Tcp, host: 'emqx', port: 1883},
         handler: {DataIngestionMicroservice.MQTTHandler, []},
         subscriptions: [{"drivers_location/#", 0}]
       ]}
    ]

    opts = [strategy: :one_for_one, name: DataIngestionMicroservice.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
