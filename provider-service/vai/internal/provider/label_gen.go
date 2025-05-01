package provider

import (
	"fmt"
	"strings"

	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
)

type LabelGen interface {
	GenerateLabels(value any) (map[string]string, error)
}

type DefaultLabelGen struct {
	providerName common.NamespacedName
}

// GenerateLabels generates labels for vertex ai runs and schedules to show
// which run configuration it originated from.
func (lg DefaultLabelGen) GenerateLabels(value any) (map[string]string, error) {
	var labels map[string]string
	switch v := value.(type) {
	case resource.RunDefinition:
		labels = lg.runLabelsFromRunDefinition(v)
	case resource.RunScheduleDefinition:
		labels = lg.runLabelsFromSchedule(v)
	default:
		return nil, fmt.Errorf(
			"Unexpected definition received [%T], expected %T or %T",
			value,
			resource.RunDefinition{},
			resource.RunScheduleDefinition{},
		)
	}
	labels[label.ProviderName] = lg.providerName.Name
	labels[label.ProviderNamespace] = lg.providerName.Namespace

	return labels, nil
}

func (lg DefaultLabelGen) runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		label.PipelineName:      pipelineName.Name,
		label.PipelineNamespace: pipelineName.Namespace,
		label.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
	}
}

func (lg DefaultLabelGen) runLabelsFromRunDefinition(
	rd resource.RunDefinition,
) map[string]string {
	runLabels := lg.runLabelsFromPipeline(
		rd.PipelineName,
		rd.PipelineVersion,
	)

	if !rd.RunConfigurationName.Empty() {
		runLabels[label.RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[label.RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[label.RunName] = rd.Name.Name
		runLabels[label.RunNamespace] = rd.Name.Namespace
	}

	return runLabels
}

func (lg DefaultLabelGen) runLabelsFromSchedule(
	rsd resource.RunScheduleDefinition,
) map[string]string {
	runLabels := lg.runLabelsFromPipeline(rsd.PipelineName, rsd.PipelineVersion)

	if !rsd.RunConfigurationName.Empty() {
		runLabels[label.RunConfigurationName] = rsd.RunConfigurationName.Name
		runLabels[label.RunConfigurationNamespace] = rsd.RunConfigurationName.Namespace
	}

	return runLabels
}
