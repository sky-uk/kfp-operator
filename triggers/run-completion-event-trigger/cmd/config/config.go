package run_completion_event_trigger

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	NATSConfig    NATSConfig   `mapstructure:"natsConfig"`
	ServerConfig  ServerConfig `mapstructure:"serverConfig"`
	MetricsConfig ServerConfig `mapstructure:"metricsConfig"`
}

type NATSConfig struct {
	Subject      string           `mapstructure:"subject"`
	ServerConfig ServerConfig     `mapstructure:"serverConfig"`
	JetStream    *JetStreamConfig `mapstructure:"jetstream"`
	Auth         *AuthConfig      `mapstructure:"auth"`
}

// JetStreamConfig contains JetStream-specific configuration
type JetStreamConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Stream  string `mapstructure:"stream"`
	Storage string `mapstructure:"storage"` // "file" or "memory", defaults to "file"
	MaxAge  string `mapstructure:"maxAge"`  // Duration string (e.g., "24h", "7d"), defaults to "24h"
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	TLS        *TLSConfig `mapstructure:"tls"`
	ClientAuth string     `mapstructure:"clientAuth"`
}

func (ac AuthConfig) Username() string {
	if ac.ClientAuth != "" {
		for line := range strings.Lines(ac.ClientAuth) {
			if strings.HasPrefix(line, "username:") {
				return strings.TrimPrefix(line, "username:")
			}
		}
	}
	return "not-found"
}

func (ac AuthConfig) Password() string {
	if ac.ClientAuth != "" {
		for line := range strings.Lines(ac.ClientAuth) {
			if strings.HasPrefix(line, "password:") {
				return strings.TrimPrefix(line, "password:")
			}
		}
	}
	return "not-found"
}

// TLSConfig contains TLS authentication configuration
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	CertFile           string `mapstructure:"certFile"`
	KeyFile            string `mapstructure:"keyFile"`
	CAFile             string `mapstructure:"caFile"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
}

// UserPasswordConfig contains username/password authentication configuration
type UserPasswordConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

func (sc ServerConfig) ToAddr() string {
	return net.JoinHostPort(sc.Host, sc.Port)
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/run-completion-event-trigger")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error loading config %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("fatal error unmarshalling config %w", err)
	}

	return &config, nil
}
