package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/sky-uk/kfp-operator/argo/common"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/spf13/viper"
)

type Config struct {
	Server baseConfig.Server `mapstructure:"server"`
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
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	var config Config
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error reading config: %w", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("fatal error unmarshalling config %w", err)
	}
	return &config, nil
}
