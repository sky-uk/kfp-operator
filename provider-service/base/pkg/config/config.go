package config

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %w", err)
	}

	return &config, nil
}
