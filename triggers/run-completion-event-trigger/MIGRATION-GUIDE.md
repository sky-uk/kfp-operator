# NATS to JetStream Migration Guide

## Overview

This guide helps users migrate from NATS Core to JetStream for run completion events.

## Migration Phases

### Phase 1: Dual Publishing Enabled (Operator Team)

**What happens:**
- Events published to both NATS Core and JetStream
- Users continue using existing NATS EventSources
- JetStream messages accumulate (but with limits)

**Stream Configuration (prevents unbounded accumulation):**
```yaml
jetstream:
  enabled: true
  maxAge: 1h          # Messages older than 1 hour are discarded
  maxMsgs: 1000       # Maximum 1000 messages in stream
  maxBytes: 10485760  # Maximum 10MB of data
```

### Phase 2: User Migration (User Teams)

**For each user application:**

1. **Create new JetStream EventSource alongside existing one:**

```yaml
# Keep existing NATS EventSource (for safety)
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: run-completion-nats
spec:
  nats:
    run-completion:
      url: nats://eventbus-kfp-operator-events-stan-svc:4222
      subject: events
---
# Add new JetStream EventSource
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: run-completion-jetstream
spec:
  nats:
    run-completion:
      url: nats://eventbus-kfp-operator-events-stan-svc:4222
      subject: events
      # JetStream specific
      stream: run-completion-events
      consumer: my-app-consumer-v1  # Unique consumer name
      durableName: my-app-consumer-v1
```

2. **Test JetStream EventSource:**
   - Verify events are received
   - Test acknowledgment behavior
   - Validate downstream processing

3. **Update Sensors to use JetStream EventSource:**
```yaml
spec:
  dependencies:
    - name: run-completion
      eventSourceName: run-completion-jetstream  # Changed from run-completion-nats
      eventName: run-completion
```

4. **Delete old NATS EventSource** when confident

## Benefits of JetStream

- **Reliability**: At-least-once delivery with acknowledgments
- **Replay**: Can replay missed messages
- **Durability**: Messages persist across restarts
- **Monitoring**: Better observability with stream metrics

## Monitoring During Migration

Monitor JetStream stream status using NATS CLI:
```bash
# Check stream info
nats stream info run-completion-events

# List consumers
nats consumer ls run-completion-events
```

## Troubleshooting

**Q: Stream is accumulating too many messages**
A: Check if users have created JetStream consumers. Increase `maxAge` or `maxMsgs` temporarily.

**Q: Messages are being discarded**
A: Normal during migration. Old messages are discarded when limits reached.

**Q: JetStream EventSource not receiving events**
A: Verify consumer name is unique and stream exists.
