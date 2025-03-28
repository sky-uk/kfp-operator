package workflowfactory

import (
	"strings"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RunScheduleDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (rsdc RunScheduleDefinitionCreator) runScheduleDefinition(
	rs *pipelinesv1.RunSchedule,
) (providers.RunScheduleDefinition, error) {
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

	return providers.RunScheduleDefinition{
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
		PipelineVersion:   rs.Spec.Pipeline.Version,
		ExperimentName:    experimentName,
		Schedule:          rs.Spec.Schedule,
		RuntimeParameters: NamedValuesToMap(rs.Spec.RuntimeParameters),
		Artifacts:         rs.Spec.Artifacts,
	}, nil
}

func runConfigurationNameForRunSchedule(
	rs *pipelinesv1.RunSchedule,
) (rcn common.NamespacedName) {
	rc := pipelinesv1.RunConfiguration{}

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
) *ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition]{
		DefinitionCreator: RunScheduleDefinitionCreator{
			Config: config,
		}.runScheduleDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
		WorkflowParamsCreator: WorkflowParamsCreatorNoop[*pipelinesv1.RunSchedule],
	}
}
