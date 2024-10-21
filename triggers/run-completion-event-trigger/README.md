# Run Completion Event Trigger gRPC service

## About:
This adds a grpc service intended to pass run completion events to a nats service. This is generally used only by the
events webhook built into the controller manager. 

## Configuration:
Configuration files `config.yaml` must be located in the same location the user is calling the binary from, i.e. `$pwd/.`.
Alternatively a `config.yaml` can be located at `/etc/run-completion-event-trigger/config.yaml` to be accessed when the binary is called
from anywhere in the system.
