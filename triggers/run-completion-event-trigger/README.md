# Run Completion Event Trigger gRPC service

## About

This is a gRPC service that publishes [run completion events](https://sky-uk.github.io/kfp-operator/docs/reference/run-completion/) 
to a NATS service. This is currently used only by the events webhook built into the controller manager.

## Configuration

The service can be configured by creating a file called `config.yaml`, which can be located either in the same directory 
the binary is executed from (i.e. `$pwd/.`) or `/etc/run-completion-event-trigger/config.yaml`.

You can see the default configuration in [config/config.yaml](https://github.com/sky-uk/kfp-operator/blob/master/triggers/run-completion-event-trigger/config/config.yaml).
