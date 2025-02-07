package config

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
)

type KfpProviderConfig struct {
	Name       string     `yaml:"name"`
	Parameters Parameters `yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}

func LoadKfpProviderConfig(providerName string) (*KfpProviderConfig, error) {
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	jsonBytes, err := json.Marshal(KfpProviderConfig{Name: providerName})
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(jsonBytes)
	if err = viper.ReadConfig(reader); err != nil {
		return nil, err
	}

	var config KfpProviderConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
