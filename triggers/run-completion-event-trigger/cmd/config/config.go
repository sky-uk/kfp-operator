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
	Subject      string          `mapstructure:"subject"`
	ServerConfig ServerConfig    `mapstructure:"serverConfig"`
	JetStream    JetStreamConfig `mapstructure:"jetstream"`
}

type JetStreamConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	StreamName string `mapstructure:"streamName"`
	Subject    string `mapstructure:"subject"`
	MaxAge     string `mapstructure:"maxAge"`
	MaxMsgs    int64  `mapstructure:"maxMsgs"`
	MaxBytes   int64  `mapstructure:"maxBytes"`
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
