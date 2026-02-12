package config

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/pkg/common"
)

type VAIProviderConfig struct {
	ProviderName        common.NamespacedName `yaml:"providerName"`
	PipelineRootStorage string                `yaml:"pipelineRootStorage"`
	Parameters          Parameters            `yaml:"parameters"`
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
