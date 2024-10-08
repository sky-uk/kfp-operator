package pipelines

import (
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
	Config config.KfpControllerConfigSpec
}

func (rcdc RunScheduleDefinitionCreator) runScheduleDefinition(runSchedule *pipelinesv1.RunSchedule) (providers.RunScheduleDefinition, error) {
	var experimentName common.NamespacedName
	if runSchedule.Spec.ExperimentName == "" {
		experimentName = common.NamespacedName{Name: rcdc.Config.DefaultExperiment}
	} else {
		experimentName = common.NamespacedName{Name: runSchedule.Spec.ExperimentName, Namespace: runSchedule.Namespace}
	}

	return providers.RunScheduleDefinition{
		Name:                 common.NamespacedName{Name: runSchedule.ObjectMeta.Name, Namespace: runSchedule.Namespace},
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

func RunScheduleWorkflowFactory(config config.KfpControllerConfigSpec) *ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.RunSchedule, providers.RunScheduleDefinition]{
		DefinitionCreator: RunScheduleDefinitionCreator{
			Config: config,
		}.runScheduleDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
