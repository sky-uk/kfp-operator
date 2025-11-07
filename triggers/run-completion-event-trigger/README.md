# Run Completion Event Trigger gRPC service

## About

This is a gRPC service that publishes [run completion events](https://sky-uk.github.io/kfp-operator/docs/reference/run-completion/)
to NATS. It supports dual publishing to both NATS Core and JetStream for seamless migration.
This is currently used only by the events webhook built into the controller manager.

## Configuration

The service can be configured by creating a file called `config.yaml`, which can be located either in the same directory 
the binary is executed from (i.e. `$pwd/.`) or `/etc/run-completion-event-trigger/config.yaml`.

You can see the default configuration in [config/config.yaml](https://github.com/sky-uk/kfp-operator/blob/master/triggers/run-completion-event-trigger/config/config.yaml).

## Dual Publishing

The service supports dual publishing to both NATS Core and JetStream:

- **NATS Core**: Traditional fire-and-forget messaging (backward compatible)
- **JetStream**: Persistent streams with acknowledgments and replay capabilities

### Configuration

```yaml
natsConfig:
  subject: events
  serverConfig:
    host: localhost
    port: 4222
  jetstream:
    enabled: true              # Enable JetStream publishing
    streamName: run-completion-events
    subject: events            # Can be different from NATS Core subject
    maxAge: 24h               # Message retention time
    maxMsgs: 10000            # Maximum messages in stream
```

### Migration Strategy

1. **Phase 1**: Deploy with `jetstream.enabled: false` (NATS only)
2. **Phase 2**: Enable JetStream (`jetstream.enabled: true`) - dual publishing
3. **Phase 3**: Users migrate EventSources to consume from JetStream
4. **Phase 4**: Disable NATS Core publishing (future release)


