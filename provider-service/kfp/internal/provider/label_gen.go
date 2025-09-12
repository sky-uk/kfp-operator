package provider

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/label"
)

type LabelGen interface {
	GenerateLabels(value any) (map[string]string, error)
}

type NoopLabelGen struct{}

func (lg NoopLabelGen) GenerateLabels(value any) (map[string]string, error) {
	return map[string]string{}, nil
}

type DefaultLabelGen struct {
	providerName common.NamespacedName
}

// GenerateLabels generates labels for runs and schedules to show
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
		label.PipelineVersion:   pipelineVersion,
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
		runLabels[label.RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[label.RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[label.RunName] = rd.Name.Name
		runLabels[label.RunNamespace] = rd.Name.Namespace
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
		runLabels[label.RunConfigurationName] = rsd.RunConfigurationName.Name
		runLabels[label.RunConfigurationNamespace] = rsd.RunConfigurationName.Namespace
	}

	runLabels[triggers.Type] = rsd.TriggerIndicator.Type
	runLabels[triggers.Source] = rsd.TriggerIndicator.Source
	runLabels[triggers.SourceNamespace] = rsd.TriggerIndicator.SourceNamespace

	return runLabels
}
