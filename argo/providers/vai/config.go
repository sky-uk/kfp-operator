package vai

import (
	"fmt"
	"strings"
)

type VAIProviderConfig struct {
	Name                                  string `yaml:"name"`
	VaiProject                            string `yaml:"vaiProject"`
	VaiLocation                           string `yaml:"vaiLocation"`
	VaiJobServiceAccount                  string `yaml:"vaiJobServiceAccount"`
	GcsEndpoint                           string `yaml:"gcsEndpoint"`
	PipelineBucket                        string `yaml:"pipelineBucket"`
	RunIntentsTopic                       string `yaml:"runIntentsTopic"`
	RunsTopic                             string `yaml:"runsTopic"`
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

func (vaipc VAIProviderConfig) schedulerJobName(name string) string {
	return fmt.Sprintf("%s/jobs/%s", vaipc.parent(), name)
}

func (vaipc VAIProviderConfig) runIntentsTopicFullName() string {
	return vaipc.topicFullName(vaipc.RunIntentsTopic)
}

func (vaipc VAIProviderConfig) topicFullName(topicName string) string {
	return fmt.Sprintf("projects/%s/topics/%s", vaipc.VaiProject, topicName)
}

func (vaipc VAIProviderConfig) pipelineStorageObject(pipelineName string, pipelineVersion string) string {
	return fmt.Sprintf("%s/%s", pipelineName, pipelineVersion)
}

func (vaipc VAIProviderConfig) gcsUri(bucket string, pathSegments ...string) string {
	return fmt.Sprintf("gs://%s/%s", bucket, strings.Join(pathSegments, "/"))
}

func (vaipc VAIProviderConfig) pipelineUri(pipelineName string, pipelineVersion string) string {
	return vaipc.gcsUri(vaipc.PipelineBucket, vaipc.pipelineStorageObject(pipelineName, pipelineVersion))
}

func (vaipc VAIProviderConfig) getMaxConcurrentRunCount() int64 {
	const DefaultMaxConcurrentRunCount = 10
	if vaipc.MaxConcurrentRunCount <= 0 {
		return DefaultMaxConcurrentRunCount
	} else {
		return vaipc.MaxConcurrentRunCount
	}
}
