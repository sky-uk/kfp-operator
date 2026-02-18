package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/spf13/viper"
)

type Config struct {
	ProviderName        common.NamespacedName `mapstructure:"providerName"`
	PipelineRootStorage string                `mapstructure:"pipelineRootStorage"`
	OperatorWebhook     string                `mapstructure:"operatorWebhook"`
	Server              Server                `mapstructure:"server"`
	Metrics             MetricsConfig         `mapstructure:"metrics"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MetricsConfig struct {
	Port int `mapstructure:"port"`
}

func LoadConfig[T any](initConfig T) (*T, error) {
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	jsonBytes, err := json.Marshal(initConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise config %w", err)
	}

	reader := bytes.NewReader(jsonBytes)
	if err = viper.ReadConfig(reader); err != nil {
		return nil, fmt.Errorf("viper failed to read config %w", err)
	}

	var config T
	if err := viper.Unmarshal(&config, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %w", err)
	}

	return &config, nil
}
