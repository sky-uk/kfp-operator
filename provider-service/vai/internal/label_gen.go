package internal

import (
	"fmt"
	"strings"

	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type LabelGen interface {
	GenerateLabels(value any) (map[string]string, error)
}

type DefaultLabelGen struct{}

func (lg DefaultLabelGen) GenerateLabels(value any) (map[string]string, error) {
	switch v := value.(type) {
	case resource.RunDefinition:
		return lg.runLabelsFromRunDefinition(v), nil
	case resource.RunScheduleDefinition:
		return lg.runLabelsFromSchedule(v), nil
	default:
		return nil, fmt.Errorf(
			"Unexpected definition received [%T], expected %T or %T",
			value,
			resource.RunDefinition{},
			resource.RunScheduleDefinition{},
		)
	}
}

func (lg DefaultLabelGen) runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		labels.PipelineName:      pipelineName.Name,
		labels.PipelineNamespace: pipelineName.Namespace,
		labels.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
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

func (lg DefaultLabelGen) runLabelsFromSchedule(
	rsd resource.RunScheduleDefinition,
) map[string]string {
	runLabels := lg.runLabelsFromPipeline(rsd.PipelineName, rsd.PipelineVersion)

	if !rsd.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] = rsd.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] = rsd.RunConfigurationName.Namespace
	}

	return runLabels
}
