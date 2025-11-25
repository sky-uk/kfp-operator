package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/pkg/common"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
)

type PublisherHandler interface {
	Publish(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error
	Name() string
	IsHealthy() bool
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

type DataWrapper struct {
	Data common.RunCompletionEvent `json:"data"`
}

// NewPublisherFromConfig creates the appropriate publisher based on configuration
func NewPublisherFromConfig(ctx context.Context, config *configLoader.NATSConfig) (PublisherHandler, error) {
	if isJetStreamEnabled(config) {
		return NewJetStreamPublisher(ctx, config)
	}

	nc, err := createNATSConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS connection: %w", err)
	}

	return NewNatsPublisher(ctx, nc, config.Subject), nil
}

// createNATSConnection creates a NATS connection based on the configuration
func createNATSConnection(ctx context.Context, config *configLoader.NATSConfig) (*nats.Conn, error) {
	logger := logr.FromContextOrDiscard(ctx)
	// Build connection options with explicit configuration
	opts := []nats.Option{
		nats.Name("kfp-operator-run-completion-event-trigger"),
		nats.RetryOnFailedConnect(true),
	}

	// Add TLS configuration if enabled
	if config.Auth != nil && config.Auth.TLS != nil && config.Auth.TLS.Enabled {
		tlsConfig, err := buildTLSConfig(config.Auth.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts = append(opts,
			nats.Secure(tlsConfig),
			nats.TLSHandshakeFirst(),
		)
	}

	// Add authentication options
	if config.Auth != nil {
		authOpts, err := buildAuthOptions(config.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to build auth options: %w", err)
		}
		opts = append(opts, authOpts...)
	}

	// Create the connection
	serverURL := buildServerURL(config)

	logger.Info("attempting to connect to NATS server", "serverURL", serverURL)

	conn, err := nats.Connect(serverURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	return conn, nil
}

// buildServerURL constructs the server URL from configuration
func buildServerURL(config *configLoader.NATSConfig) string {
	// Use TLS scheme if TLS is enabled
	scheme := "nats"

	return fmt.Sprintf("%s://%s", scheme, config.ServerConfig.ToAddr())
}

// buildAuthOptions creates authentication options based on configuration (excluding TLS)
func buildAuthOptions(auth *configLoader.AuthConfig) ([]nats.Option, error) {
	var opts []nats.Option

	// Username/password authentication
	username := auth.Username()
	password := auth.Password()
	if username != "" && password != "" {
		opts = append(opts, nats.UserInfo(username, password))
	}

	return opts, nil
}

// buildTLSConfig creates a TLS configuration from the auth config
func buildTLSConfig(tlsConfig *configLoader.TLSConfig) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}

	// Load client certificate if provided
	if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if tlsConfig.CAFile != "" {
		caCert, err := os.ReadFile(tlsConfig.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate from %s", tlsConfig.CAFile)
		}
		config.RootCAs = caCertPool
	}

	return config, nil
}

// isJetStreamEnabled returns true if JetStream is enabled in the configuration
func isJetStreamEnabled(config *configLoader.NATSConfig) bool {
	return config.JetStream != nil && config.JetStream.Enabled
}
