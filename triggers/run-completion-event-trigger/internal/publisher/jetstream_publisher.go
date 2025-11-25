package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sky-uk/kfp-operator/pkg/common"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
)

type JetStreamPublisher struct {
	NatsConn  natsConn
	Subject   string
	JetStream jetstream.JetStream
	Config    *configLoader.NATSConfig
}

func NewJetStreamPublisher(ctx context.Context, config *configLoader.NATSConfig) (*JetStreamPublisher, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Create NATS connection with authentication and TLS support
	nc, err := createNATSConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS connection: %w", err)
	}

	// Create JetStream instance
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream instance: %w", err)
	}

	publisher := &JetStreamPublisher{
		NatsConn:  nc,
		Subject:   config.Subject,
		JetStream: js,
		Config:    config,
	}

	// Ensure the JetStream stream exists if configured
	if config.JetStream != nil && config.JetStream.Stream != "" {
		if err := publisher.ensureStreamExists(ctx); err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to ensure stream exists: %w", err)
		}
	}

	logger.Info("jetstream setup", "nc servers", nc.Servers(), "connected to", nc.ConnectedAddr())

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if publisher.NatsConn != nil {
					conn := publisher.NatsConn
					fmt.Printf("[DEBUG] NATS Status: connected=%v, url=%s, status=%v, last_error=%v\n",
						nc.IsConnected(),
						nc.ConnectedUrl(),
						nc.Status(),
						nc.LastError())

					// Quick JetStream test
					if publisher.JetStream != nil && conn.IsConnected() {
						testCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
						_, err := publisher.JetStream.AccountInfo(testCtx)
						cancel()
						fmt.Printf("[DEBUG] JetStream test: %v\n", err)
					}
				} else {
					fmt.Printf("[DEBUG] NATS connection is nil\n")
				}
			}
		}
	}()

	logger.Info("New JetStream publisher:", "Subject", config.Subject, "Server", nc.ConnectedUrl(), "Stream", config.JetStream.Stream)
	return publisher, nil
}

func (nc *JetStreamPublisher) Publish(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
	dataWrapper := DataWrapper{Data: runCompletionEvent}
	eventData, err := json.Marshal(dataWrapper)
	if err != nil {
		return &MarshallingError{err.Error()}
	}

	_, err = nc.JetStream.Publish(ctx, nc.Subject, eventData)
	if err != nil {
		return &ConnectionError{fmt.Sprintf("JetStream publish %s failed: %v", runCompletionEvent.RunId, err)}
	}

	return nil
}

func (nc *JetStreamPublisher) Name() string {
	return "nats-jetstream-publisher"
}

func (nc *JetStreamPublisher) IsHealthy() bool {
	return nc.NatsConn.IsConnected()
}

func (nc *JetStreamPublisher) Close() {
	if conn, ok := nc.NatsConn.(*nats.Conn); ok {
		conn.Close()
	}
}

func (nc *JetStreamPublisher) ensureStreamExists(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)
	streamName := nc.Config.JetStream.Stream

	_, err := nc.JetStream.Stream(ctx, streamName)
	if err == nil {
		logger.Info("JetStream stream already exists", "stream", streamName)
		return nil
	}

	// If error is not "stream not found", return the error
	if !isStreamNotFoundError(err) {
		return fmt.Errorf("failed to check stream '%s' existence: %w", streamName, err)
	}

	// Create the stream
	logger.Info("Creating JetStream stream", "stream", streamName)
	streamConfig, err := nc.createStreamConfig()
	if err != nil {
		return fmt.Errorf("failed to create stream config for '%s': %w", streamName, err)
	}

	_, err = nc.JetStream.CreateStream(ctx, streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create stream '%s': %w", streamName, err)
	}

	logger.Info("Successfully created JetStream", "stream", streamName)
	return nil
}

// createStreamConfig creates a stream configuration based on the publisher configuration
func (nc *JetStreamPublisher) createStreamConfig() (jetstream.StreamConfig, error) {
	subjects := []string{nc.Subject}

	// Determine storage type (default to file storage)
	storage := jetstream.FileStorage
	if nc.Config.JetStream.Storage != "" {
		switch strings.ToLower(nc.Config.JetStream.Storage) {
		case "file":
			storage = jetstream.FileStorage
		case "memory":
			storage = jetstream.MemoryStorage
		default:
			return jetstream.StreamConfig{}, fmt.Errorf("invalid storage type '%s': must be 'file' or 'memory'", nc.Config.JetStream.Storage)
		}
	}

	// Parse max age duration (default to 24 hours)
	maxAge := 24 * time.Hour // Default
	if nc.Config.JetStream.MaxAge != "" {
		duration, err := time.ParseDuration(nc.Config.JetStream.MaxAge)
		if err != nil {
			return jetstream.StreamConfig{}, fmt.Errorf("invalid maxAge duration '%s': %w", nc.Config.JetStream.MaxAge, err)
		}
		maxAge = duration
	}

	return jetstream.StreamConfig{
		Name:     nc.Config.JetStream.Stream,
		Subjects: subjects,
		Storage:  storage,
		MaxAge:   maxAge,
	}, nil
}

// isStreamNotFoundError checks if the error indicates that a stream was not found
func isStreamNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for JetStream API error first (most specific)
	var apiErr *jetstream.APIError
	if errors.As(err, &apiErr) {
		// JetStream API error codes for stream not found
		return apiErr.ErrorCode == jetstream.JSErrCodeStreamNotFound
	}

	// Check for standard NATS errors
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		return true
	}

	// Fallback to error message checking (for compatibility with older versions)
	errMsg := err.Error()
	return errMsg == "nats: stream not found" || errMsg == "stream not found"
}
