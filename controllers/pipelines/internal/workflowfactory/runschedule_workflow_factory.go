package workflowfactory

import (
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
	"strings"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RunScheduleDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (rsdc RunScheduleDefinitionCreator) runScheduleDefinition(
	_ pipelineshub.Provider,
	rs *pipelineshub.RunSchedule,
) ([]pipelineshub.Patch, providers.RunScheduleDefinition, error) {
	var experimentName common.NamespacedName
	if rs.Spec.ExperimentName == "" {
		experimentName = common.NamespacedName{
			Name: rsdc.Config.DefaultExperiment,
		}
	} else {
		experimentName = common.NamespacedName{
			Name:      rs.Spec.ExperimentName,
			Namespace: rs.Namespace,
		}
	}

	return nil, providers.RunScheduleDefinition{
		Name: common.NamespacedName{
			Name:      rs.ObjectMeta.Name,
			Namespace: rs.Namespace,
		},
		RunConfigurationName: runConfigurationNameForRunSchedule(rs),
		Version:              rs.ComputeVersion(),
		PipelineName: common.NamespacedName{
			Name:      rs.Spec.Pipeline.Name,
			Namespace: rs.Namespace,
		},
		PipelineVersion: rs.Spec.Pipeline.Version,
		ExperimentName:  experimentName,
		Schedule:        rs.Spec.Schedule,
		Parameters:      NamedValuesToMap(rs.Spec.Parameters),
		Artifacts:       rs.Spec.Artifacts,
		TriggerIndicator: triggers.Indicator{
			Type:            triggers.Schedule,
			Source:          rs.Name,
			SourceNamespace: rs.Namespace,
		},
	}, nil
}

func runConfigurationNameForRunSchedule(
	rs *pipelineshub.RunSchedule,
) (rcn common.NamespacedName) {
	rc := pipelineshub.RunConfiguration{}

	owner := metav1.GetControllerOf(rs)
	if owner == nil {
		return
	}

	ownerGroupVersion, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return
	}

	if ownerGroupVersion.Group == apis.Group && strings.ToLower(owner.Kind) == rc.GetKind() {
		rcn.Name = owner.Name
		rcn.Namespace = rs.Namespace
	}

	return
}

func RunScheduleWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.RunSchedule, providers.RunScheduleDefinition] {
	return &ResourceWorkflowFactory[*pipelineshub.RunSchedule, providers.RunScheduleDefinition]{
		DefinitionCreator: RunScheduleDefinitionCreator{
			Config: config,
		}.runScheduleDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
		WorkflowParamsCreator: WorkflowParamsCreatorNoop[*pipelineshub.RunSchedule],
	}
}
