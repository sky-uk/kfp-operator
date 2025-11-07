# Testing Dual Publishing Implementation

This document describes how to test the dual publishing functionality.

## Quick Test with Docker Compose

1. **Start the services:**
   ```bash
   cd triggers/run-completion-event-trigger
   docker-compose up --build
   ```

2. **Check NATS JetStream is enabled:**
   ```bash
   # Connect to NATS container
   docker exec -it run-completion-event-trigger-nats-1 nats stream ls
   ```

3. **Send a test event:**
   ```bash
   # Use grpcurl to send a test event
   grpcurl -plaintext -d '{
     "pipeline_name": "test-pipeline",
     "provider": "test-provider", 
     "run_id": "test-run-123",
     "status": "SUCCEEDED"
   }' localhost:50051 run_completion_event_trigger.RunCompletionEventTrigger/ProcessEventFeed
   ```

4. **Verify JetStream stream was created:**
   ```bash
   docker exec -it run-completion-event-trigger-nats-1 nats stream info run-completion-events
   ```

## Configuration Options

### Enable JetStream (in config.yaml):
```yaml
natsConfig:
  jetstream:
    enabled: true
    streamName: run-completion-events
    subject: events
    maxAge: 24h
    maxMsgs: 10000
```

### Disable JetStream (NATS only):
```yaml
natsConfig:
  jetstream:
    enabled: false
```

## Expected Behavior

- **JetStream Enabled**: Events published to both NATS Core and JetStream
- **JetStream Disabled**: Events published only to NATS Core
- **Health Checks**: Consider both NATS and JetStream status when enabled
