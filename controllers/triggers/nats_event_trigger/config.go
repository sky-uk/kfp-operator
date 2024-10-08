package nats_event_trigger

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NATSConfig   *NATSConfig   `yaml:"natsConfig,omitempty"`
	ServerConfig *ServerConfig `yaml:"serverConfig,omitempty"`
}

type NATSConfig struct {
	Subject *string `yaml:"subject,omitempty"`
	Url     *string `yaml:"url,omitempty"`
}

type ServerConfig struct {
	Host *string `yaml:"host,omitempty"`
	Port *string `yaml:"port,omitempty"`
}

func LoadConfig(file io.Reader) (*Config, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func loadConfigFromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file at (%s): %w", filename, err)
	}
	defer file.Close()
	return LoadConfig(file)
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Config
	if err := unmarshal((*alias)(c)); err != nil {
		return err
	}

	var missingConfigErrors []error

	if c.NATSConfig == nil {
		missingConfigErrors = append(missingConfigErrors, errors.New("missing NATSConfig"))
	} else {
		if c.NATSConfig.Subject == nil {
			missingConfigErrors = append(missingConfigErrors, errors.New("missing NATSConfig field: Subject"))
		}

		if c.NATSConfig.Url == nil {
			missingConfigErrors = append(missingConfigErrors, errors.New("missing NATSConfig field: Url"))
		}
	}

	if c.ServerConfig == nil {
		missingConfigErrors = append(missingConfigErrors, errors.New("missing ServerConfig"))
	} else {
		if c.ServerConfig.Host == nil {
			missingConfigErrors = append(missingConfigErrors, errors.New("missing ServerConfig field: Host"))
		}

		if c.ServerConfig.Port == nil {
			missingConfigErrors = append(missingConfigErrors, errors.New("missing ServerConfig field: Port"))
		}
	}
	if len(missingConfigErrors) > 0 {
		return fmt.Errorf("missing config errors: %v", missingConfigErrors)
	}

	return nil
}
