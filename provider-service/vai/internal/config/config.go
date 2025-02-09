package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type VAIProviderConfig struct {
	Name       string     `yaml:"name"`
	Parameters Parameters `yaml:"parameters"`
}

type Parameters struct {
	VaiProject                            string `yaml:"vaiProject"`
	VaiLocation                           string `yaml:"vaiLocation"`
	VaiJobServiceAccount                  string `yaml:"vaiJobServiceAccount"`
	GcsEndpoint                           string `yaml:"gcsEndpoint"`
	PipelineBucket                        string `yaml:"pipelineBucket"`
	EventsourcePipelineEventsSubscription string `yaml:"eventsourcePipelineEventsSubscription"`
	MaxConcurrentRunCount                 int64  `yaml:"maxConcurrentRunCount"`
}

func LoadVAIProviderConfig(providerName string) (*VAIProviderConfig, error) {
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	jsonBytes, err := json.Marshal(VAIProviderConfig{Name: providerName})
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(jsonBytes)
	if err = viper.ReadConfig(reader); err != nil {
		return nil, err
	}

	var config VAIProviderConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (vaipc VAIProviderConfig) VaiEndpoint() string {
	return fmt.Sprintf("%s-aiplatform.googleapis.com:443", vaipc.Parameters.VaiLocation)
}

func (vaipc VAIProviderConfig) Parent() string {
	return fmt.Sprintf(`projects/%s/locations/%s`, vaipc.Parameters.VaiProject, vaipc.Parameters.VaiLocation)
}

func (vaipc VAIProviderConfig) PipelineJobName(name string) string {
	return fmt.Sprintf("%s/pipelineJobs/%s", vaipc.Parent(), name)
}

func (vaipc VAIProviderConfig) GetMaxConcurrentRunCountOrDefault() int64 {
	const defaultMaxConcurrentRunCount = 10
	if vaipc.Parameters.MaxConcurrentRunCount <= 0 {
		return defaultMaxConcurrentRunCount
	} else {
		return vaipc.Parameters.MaxConcurrentRunCount
	}
}
