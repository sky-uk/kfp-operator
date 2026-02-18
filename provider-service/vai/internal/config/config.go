package config

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/pkg/common"
)

type VAIProviderConfig struct {
	ProviderName        common.NamespacedName `mapstructure:"providerName" yaml:"providerName"`
	PipelineRootStorage string                `mapstructure:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          Parameters            `mapstructure:"parameters" yaml:"parameters"`
}

type Parameters struct {
	VaiProject                            string `mapstructure:"vaiProject" yaml:"vaiProject"`
	VaiLocation                           string `mapstructure:"vaiLocation" yaml:"vaiLocation"`
	VaiJobServiceAccount                  string `mapstructure:"vaiJobServiceAccount" yaml:"vaiJobServiceAccount"`
	GcsEndpoint                           string `mapstructure:"gcsEndpoint" yaml:"gcsEndpoint"`
	PipelineBucket                        string `mapstructure:"pipelineBucket" yaml:"pipelineBucket"`
	EventsourcePipelineEventsSubscription string `mapstructure:"eventsourcePipelineEventsSubscription" yaml:"eventsourcePipelineEventsSubscription"`
	MaxConcurrentRunCount                 int64  `mapstructure:"maxConcurrentRunCount" yaml:"maxConcurrentRunCount"`
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
