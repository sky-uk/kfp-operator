package config

import (
	"context"
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
		logger.Error(err, "failed to load config file")
		return nil, err
	}

	logger.Info(fmt.Sprintf("loaded config: %+v", config))
	return config, nil
}

func load() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/provider-service")
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
