defmodule DataIngestionMicroservice.MixProject do
  use Mix.Project

  def project do
    [
      app: :data_ingestion_microservice,
      version: "0.1.0",
      elixir: "~> 1.18",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger, :emqtt],
      mod: {DataIngestionMicroservice.Application, []}
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      {:simplify, "~> 2.0"},
      {:redix, "~> 1.5"},
      {:jason, "~> 1.4"},
      {:emqtt, github: "emqx/emqtt", tag: "1.4.4", system_env: [{"BUILD_WITHOUT_QUIC", "1"}]}
    ]
  end
end
