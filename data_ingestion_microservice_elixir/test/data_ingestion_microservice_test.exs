defmodule DataIngestionMicroserviceTest do
  use ExUnit.Case
  doctest DataIngestionMicroservice

  test "greets the world" do
    assert DataIngestionMicroservice.hello() == :world
  end
end
