package label

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type LabelGen interface {
	GenerateLabels(value any) (map[string]string, error)
}

type DefaultLabelGen struct {
	ProviderName common.NamespacedName
}

// GenerateLabels generates labels for vertex ai runs and schedules to show
// which run configuration it originated from.
func (lg DefaultLabelGen) GenerateLabels(value any) (map[string]string, error) {
	var labels map[string]string

	switch v := value.(type) {
	case base.RunDefinition:
		labels = lg.runLabelsFromRunDefinition(v)
	case base.RunScheduleDefinition:
		labels = lg.runLabelsFromSchedule(v)
	default:
		return nil, fmt.Errorf(
			"unexpected definition received [%T], expected %T or %T",
			value,
			base.RunDefinition{},
			base.RunScheduleDefinition{},
		)
	}

	labels[ProviderName] = lg.ProviderName.Name
	labels[ProviderNamespace] = lg.ProviderName.Namespace

	return labels, nil
}

func (lg DefaultLabelGen) runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		PipelineName:      pipelineName.Name,
		PipelineNamespace: pipelineName.Namespace,
		PipelineVersion:   pipelineVersion,
	}
}

func (lg DefaultLabelGen) runLabelsFromRunDefinition(
	rd base.RunDefinition,
) map[string]string {
	runLabels := lg.runLabelsFromPipeline(
		rd.PipelineName,
		rd.PipelineVersion,
	)

	if !rd.RunConfigurationName.Empty() {
		runLabels[RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[RunName] = rd.Name.Name
		runLabels[RunNamespace] = rd.Name.Namespace
	}

	if rd.TriggerIndicator != nil {
		runLabels[triggers.Type] = rd.TriggerIndicator.Type
		runLabels[triggers.Source] = rd.TriggerIndicator.Source
		runLabels[triggers.SourceNamespace] = rd.TriggerIndicator.SourceNamespace
	}

	return runLabels
}

func (lg DefaultLabelGen) runLabelsFromSchedule(
	rsd base.RunScheduleDefinition,
) map[string]string {
	runLabels := lg.runLabelsFromPipeline(rsd.PipelineName, rsd.PipelineVersion)

	if !rsd.RunConfigurationName.Empty() {
		runLabels[RunConfigurationName] = rsd.RunConfigurationName.Name
		runLabels[RunConfigurationNamespace] = rsd.RunConfigurationName.Namespace
	}

	runLabels[triggers.Type] = rsd.TriggerIndicator.Type
	runLabels[triggers.Source] = rsd.TriggerIndicator.Source
	runLabels[triggers.SourceNamespace] = rsd.TriggerIndicator.SourceNamespace

	return runLabels
}
