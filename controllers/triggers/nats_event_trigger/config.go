package nats_event_trigger

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NATSConfig   NATSConfig   `yaml:"natsConfig,omitempty"`
	ServerConfig ServerConfig `yaml:"serverConfig,omitempty"`
}

type NATSConfig struct {
	Subject string `yaml:"subject,omitempty"`
	Url     string `yaml:"url,omitempty"`
}

type ServerConfig struct {
	Host string `yaml:"host,omitempty"`
	Port string `yaml:"port,omitempty"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at (%s): %w", filename, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
