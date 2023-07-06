package pipelines

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
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

	runtimeParameters := make(map[string]string)

	for _, parameter := range runSchedule.Spec.RuntimeParameters {
		if parameter.Value == "" {
			return providers.RunScheduleDefinition{}, fmt.Errorf("runSchedules only supports Named/Value RuntimeParameters")

		}

		runtimeParameters[parameter.Name] = parameter.Value
	}

	return providers.RunScheduleDefinition{
		Name:                 runSchedule.ObjectMeta.Name,
		RunConfigurationName: runConfigurationNameForRunSchedule(runSchedule),
		Version:              runSchedule.ComputeVersion(),
		PipelineName:         common.NamespacedName{Name: runSchedule.Spec.Pipeline.Name, Namespace: runSchedule.Namespace},
		PipelineVersion:      runSchedule.Spec.Pipeline.Version,
		ExperimentName:       experimentName,
		Schedule:             runSchedule.Spec.Schedule,
		RuntimeParameters:    NamedValuesToMap(runSchedule.Spec.RuntimeParameters),
		Artifacts:            runSchedule.Spec.Artifacts,
	}, nil
}

func runConfigurationNameForRunSchedule(runSchedule *pipelinesv1.RunSchedule) (runConfigurationName common.NamespacedName) {
	rc := pipelinesv1.RunConfiguration{}

	owner := metav1.GetControllerOf(runSchedule)
	if owner == nil {
		return
	}

	ownerGroupVersion, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return
	}

	if ownerGroupVersion.Group == apis.Group && strings.ToLower(owner.Kind) == rc.GetKind() {
		runConfigurationName.Name = owner.Name
		runConfigurationName.Namespace = runSchedule.Namespace
	}

	return
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
