package config

import (
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

type Config struct {
	Server baseConfig.Server `mapstructure:"server"`
}
