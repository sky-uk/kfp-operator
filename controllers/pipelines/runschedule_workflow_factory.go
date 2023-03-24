package pipelines

import (
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type RunScheduleDefinitionCreator struct {
	Config config.Configuration
}

func (rcdc RunScheduleDefinitionCreator) runScheduleDefinition(runSchedule *pipelinesv1.RunSchedule) (providers.RunScheduleDefinition, error) {
	var experimentName string

	if runSchedule.Spec.ExperimentName == "" {
		experimentName = rcdc.Config.DefaultExperiment
	} else {
		experimentName = runSchedule.Spec.ExperimentName
	}

	return providers.RunScheduleDefinition{
		Name:                 runSchedule.ObjectMeta.Name,
		RunConfigurationName: runConfigurationNameForRunSchedule(runSchedule),
		Version:              runSchedule.ComputeVersion(),
		PipelineName:         runSchedule.Spec.Pipeline.Name,
		PipelineVersion:      runSchedule.Spec.Pipeline.Version,
		ExperimentName:       experimentName,
		Schedule:             runSchedule.Spec.Schedule,
		RuntimeParameters:    NamedValuesToMap(runSchedule.Spec.RuntimeParameters),
	}, nil
}

func runConfigurationNameForRunSchedule(runSchedule *pipelinesv1.RunSchedule) string {
	rc := pipelinesv1.RunConfiguration{}

	owner := metav1.GetControllerOf(runSchedule)
	if owner == nil {
		return ""
	}

	ownerGroupVersion, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return ""
	}

	if ownerGroupVersion.Group == apis.Group && strings.ToLower(owner.Kind) == rc.GetKind() {
		return owner.Name
	}

	return ""
}

func RunScheduleWorkflowFactory(config config.Configuration) *ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition]{
		DefinitionCreator: RunScheduleDefinitionCreator{
			Config: config,
		}.runScheduleDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
