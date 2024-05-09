package vai

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type VAIProviderConfig struct {
	Name                                  string `yaml:"name"`
	VaiProject                            string `yaml:"vaiProject"`
	VaiLocation                           string `yaml:"vaiLocation"`
	VaiJobServiceAccount                  string `yaml:"vaiJobServiceAccount"`
	GcsEndpoint                           string `yaml:"gcsEndpoint"`
	PipelineBucket                        string `yaml:"pipelineBucket"`
	EventsourcePipelineEventsSubscription string `yaml:"eventsourcePipelineEventsSubscription"`
	MaxConcurrentRunCount                 int64  `yaml:"maxConcurrentRunCount"`
}

func (vaipc VAIProviderConfig) vaiEndpoint() string {
	return fmt.Sprintf("%s-aiplatform.googleapis.com:443", vaipc.VaiLocation)
}

func (vaipc VAIProviderConfig) parent() string {
	return fmt.Sprintf(`projects/%s/locations/%s`, vaipc.VaiProject, vaipc.VaiLocation)
}

func (vaipc VAIProviderConfig) pipelineJobName(name string) string {
	return fmt.Sprintf("%s/pipelineJobs/%s", vaipc.parent(), name)
}

func (vaipc VAIProviderConfig) pipelineStorageObject(pipelineName common.NamespacedName, pipelineVersion string) (string, error) {
	namespaceName, err := pipelineName.String()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", namespaceName, pipelineVersion), nil
}

func (vaipc VAIProviderConfig) pipelineUri(pipelineName common.NamespacedName, pipelineVersion string) (string, error) {
	pipelineUri, err := vaipc.pipelineStorageObject(pipelineName, pipelineVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gs://%s/%s", vaipc.PipelineBucket, pipelineUri), nil
}

func (vaipc VAIProviderConfig) getMaxConcurrentRunCountOrDefault() int64 {
	const DefaultMaxConcurrentRunCount = 10
	if vaipc.MaxConcurrentRunCount <= 0 {
		return DefaultMaxConcurrentRunCount
	} else {
		return vaipc.MaxConcurrentRunCount
	}
}
