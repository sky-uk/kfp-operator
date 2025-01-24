package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ProviderName    string `mapstructure:"providerName"`
	OperatorWebhook string `mapstructure:"operatorWebhook"`
	Pod             Pod    `mapstructure:"pod"`
	Server          Server `mapstructure:"server"`
}

type Pod struct {
	Namespace string `mapstructure:"namespace"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	logger := common.LoggerFromContext(ctx)
	config, err := load()

	if err != nil {
		logger.Error(err, "failed to load config")
		return nil, err
	}

	logger.Info(fmt.Sprintf("loaded config: %+v", config))
	return config, nil
}

func load() (*Config, error) {
	if err := initConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialise viper config %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("fatal error unmarshalling config %w", err)
	}

	return &config, nil
}

// this initialises the config in Viper with default empty values so that they can be overridden with env vars
func initConfig() error {
	var config Config
	viper.SetConfigType("json")

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jsonBytes)
	if err = viper.ReadConfig(reader); err != nil {
		return err
	}

	return nil
}
