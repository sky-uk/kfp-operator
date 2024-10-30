package config

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"go.uber.org/zap/zapcore"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ProviderName    string `mapstructure:"providerName"`
	OperatorWebhook string `mapstructure:"operatorWebhook"`
	Pod             Pod    `mapstructure:"pod"`
}

type Pod struct {
	Namespace string `mapstructure:"namespace"`
}

func LoadConfig() (*Config, error) {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		logger.Error(err, "Failed to create zap logger")
		return nil, err
	}

	config, err := load()

	if err != nil {
		logger.Error(err, "Failed to load config file")
		return nil, err
	}

	logger.Info(fmt.Sprintf("Loaded config: %+v", config))
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
