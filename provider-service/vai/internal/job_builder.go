package internal

import (
	"fmt"
	"strings"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type JobBuilder struct {
	serviceAccount string
	pipelineBucket string
}

func (b JobBuilder) MkPipelineJob(
	rd resource.RunDefinition,
) (*aiplatformpb.PipelineJob, error) {
	params := make(map[string]*aiplatformpb.Value, len(rd.RuntimeParameters))
	for name, value := range rd.RuntimeParameters {
		params[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	// TODO: see if pipelinePath can be passed in instead.
	templateUri, err := b.pipelineUri(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return nil, err
	}

	job := &aiplatformpb.PipelineJob{
		Labels: b.runLabelsFromRunDefinition(rd),
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: params,
		},
		ServiceAccount: b.serviceAccount,
		TemplateUri:    templateUri,
	}

	return job, nil
}

// returns namespaceName/pipelineVersion
// e.g. namespace/name/pipelineVersion
func (b JobBuilder) pipelineStorageObject(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) (string, error) {
	namespaceName, err := pipelineName.String()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", namespaceName, pipelineVersion), nil
}

func (b JobBuilder) pipelineUri(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) (string, error) {
	pipelineUri, err := b.pipelineStorageObject(pipelineName, pipelineVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gs://%s/%s", b.pipelineBucket, pipelineUri), nil
}

func (b JobBuilder) runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		labels.PipelineName:      pipelineName.Name,
		labels.PipelineNamespace: pipelineName.Namespace,
		labels.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
	}
}

func (b JobBuilder) runLabelsFromRunDefinition(
	rd resource.RunDefinition,
) map[string]string {
	runLabels := b.runLabelsFromPipeline(
		rd.PipelineName,
		rd.PipelineVersion,
	)

	if !rd.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[labels.RunName] = rd.Name.Name
		runLabels[labels.RunNamespace] = rd.Name.Namespace
	}

	return runLabels
}
