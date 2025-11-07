package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/pkg/common"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
)

type PublisherHandler interface {
	Publish(runCompletionEvent common.RunCompletionEvent) error
}

type MarshallingError struct {
	Message string
}

func (e *MarshallingError) Error() string {
	return e.Message
}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

type natsConn interface {
	Publish(subject string, data []byte) error
	IsConnected() bool
}

type NatsPublisher struct {
	NatsConn natsConn
	Subject  string
}

type DataWrapper struct {
	Data common.RunCompletionEvent `json:"data"`
}

func NewNatsPublisher(ctx context.Context, nc *nats.Conn, subject string) *NatsPublisher {
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info("New nats publisher:", "Subject", subject, "Server", nc.ConnectedUrl())
	return &NatsPublisher{
		NatsConn: nc,
		Subject:  subject,
	}
}

func (nc *NatsPublisher) Publish(runCompletionEvent common.RunCompletionEvent) error {
	dataWrapper := DataWrapper{Data: runCompletionEvent}
	eventData, err := json.Marshal(dataWrapper)
	if err != nil {
		return &MarshallingError{err.Error()}
	}
	if err := nc.NatsConn.Publish(nc.Subject, eventData); err != nil {
		return &ConnectionError{err.Error()}
	}
	return nil
}

func (nc *NatsPublisher) Name() string {
	return "nats-publisher"
}

func (nc *NatsPublisher) IsHealthy() bool {
	return nc.NatsConn.IsConnected()
}

// JetStreamPublisher publishes events to NATS JetStream
type JetStreamPublisher struct {
	JetStream nats.JetStreamContext
	Subject   string
	logger    logr.Logger
}

// NewJetStreamPublisher creates a new JetStream publisher and ensures the stream exists
func NewJetStreamPublisher(ctx context.Context, nc *nats.Conn, config configLoader.JetStreamConfig) (*JetStreamPublisher, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	publisher := &JetStreamPublisher{
		JetStream: js,
		Subject:   config.Subject,
		logger:    logger,
	}

	// Ensure stream exists
	if err := publisher.ensureStream(config); err != nil {
		return nil, fmt.Errorf("failed to ensure stream exists: %w", err)
	}

	logger.Info("New JetStream publisher created", "Subject", config.Subject, "Stream", config.StreamName)
	return publisher, nil
}

func (jsp *JetStreamPublisher) ensureStream(config configLoader.JetStreamConfig) error {
	// Parse maxAge duration
	maxAge, err := time.ParseDuration(config.MaxAge)
	if err != nil {
		return fmt.Errorf("invalid maxAge duration %s: %w", config.MaxAge, err)
	}

	// Define stream configuration with limits to prevent unbounded accumulation
	streamConfig := &nats.StreamConfig{
		Name:      config.StreamName,
		Subjects:  []string{config.Subject},
		Retention: nats.LimitsPolicy,
		MaxAge:    maxAge,
		MaxMsgs:   config.MaxMsgs,
		MaxBytes:  config.MaxBytes,
		Storage:   nats.FileStorage,
		Discard:   nats.DiscardOld, // Discard old messages when limits reached
	}

	// Try to get existing stream info
	_, err = jsp.JetStream.StreamInfo(config.StreamName)
	if err != nil {
		// Stream doesn't exist, create it
		if err == nats.ErrStreamNotFound {
			jsp.logger.Info("Creating JetStream stream", "name", config.StreamName)
			_, err = jsp.JetStream.AddStream(streamConfig)
			if err != nil {
				return fmt.Errorf("failed to create stream %s: %w", config.StreamName, err)
			}
			jsp.logger.Info("JetStream stream created successfully", "name", config.StreamName)
		} else {
			return fmt.Errorf("failed to get stream info for %s: %w", config.StreamName, err)
		}
	} else {
		jsp.logger.Info("JetStream stream already exists", "name", config.StreamName)
	}

	return nil
}

func (jsp *JetStreamPublisher) Publish(runCompletionEvent common.RunCompletionEvent) error {
	dataWrapper := DataWrapper{Data: runCompletionEvent}
	eventData, err := json.Marshal(dataWrapper)
	if err != nil {
		return &MarshallingError{err.Error()}
	}

	// Publish to JetStream with acknowledgment
	ack, err := jsp.JetStream.Publish(jsp.Subject, eventData)
	if err != nil {
		return &ConnectionError{fmt.Sprintf("JetStream publish failed: %v", err)}
	}

	jsp.logger.V(1).Info("Event published to JetStream",
		"subject", jsp.Subject,
		"sequence", ack.Sequence,
		"stream", ack.Stream)

	return nil
}

func (jsp *JetStreamPublisher) Name() string {
	return "jetstream-publisher"
}

func (jsp *JetStreamPublisher) IsHealthy() bool {
	// Check if we can get stream info (indicates JetStream is working)
	_, err := jsp.JetStream.AccountInfo()
	return err == nil
}

// DualPublisher publishes to both NATS Core and JetStream
type DualPublisher struct {
	natsPublisher      *NatsPublisher
	jetstreamPublisher *JetStreamPublisher
	jetstreamEnabled   bool
	logger             logr.Logger
}

// NewDualPublisher creates a publisher that can publish to both NATS and JetStream
func NewDualPublisher(ctx context.Context, nc *nats.Conn, config *configLoader.NATSConfig) (*DualPublisher, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Always create NATS publisher (for backward compatibility)
	natsPublisher := NewNatsPublisher(ctx, nc, config.Subject)

	dp := &DualPublisher{
		natsPublisher:    natsPublisher,
		jetstreamEnabled: config.JetStream.Enabled,
		logger:           logger,
	}

	// Optionally create JetStream publisher
	if config.JetStream.Enabled {
		jetstreamPublisher, err := NewJetStreamPublisher(ctx, nc, config.JetStream)
		if err != nil {
			return nil, fmt.Errorf("failed to create JetStream publisher: %w", err)
		}
		dp.jetstreamPublisher = jetstreamPublisher
		logger.Info("Dual publisher created with JetStream enabled")
	} else {
		logger.Info("Dual publisher created with JetStream disabled")
	}

	return dp, nil
}

func (dp *DualPublisher) Publish(runCompletionEvent common.RunCompletionEvent) error {
	// Always publish to NATS Core (maintain existing behavior)
	natsErr := dp.natsPublisher.Publish(runCompletionEvent)
	if natsErr != nil {
		dp.logger.Error(natsErr, "NATS Core publish failed")
	}

	// Optionally publish to JetStream
	var jetstreamErr error
	if dp.jetstreamEnabled && dp.jetstreamPublisher != nil {
		jetstreamErr = dp.jetstreamPublisher.Publish(runCompletionEvent)
		if jetstreamErr != nil {
			dp.logger.Error(jetstreamErr, "JetStream publish failed")
		}
	}

	// Fail only if NATS Core fails (maintain backward compatibility)
	// JetStream failures are logged but don't cause overall failure during transition
	if natsErr != nil {
		return natsErr
	}

	return nil
}

func (dp *DualPublisher) Name() string {
	if dp.jetstreamEnabled {
		return "dual-publisher"
	}
	return "nats-only-publisher"
}

func (dp *DualPublisher) IsHealthy() bool {
	natsHealthy := dp.natsPublisher.IsHealthy()

	if dp.jetstreamEnabled && dp.jetstreamPublisher != nil {
		jetstreamHealthy := dp.jetstreamPublisher.IsHealthy()
		// Both must be healthy if JetStream is enabled
		return natsHealthy && jetstreamHealthy
	}

	// Only NATS needs to be healthy if JetStream is disabled
	return natsHealthy
}
