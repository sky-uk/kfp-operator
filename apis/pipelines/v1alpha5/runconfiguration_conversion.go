package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = pipelines.Map(src.Spec.Run.RuntimeParameters, convertToRuntimeParameter)
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	dst.Spec.Triggers = hub.Triggers{
		Schedules: src.Spec.Triggers.Schedules,
		OnChange: pipelines.Map(src.Spec.Triggers.OnChange, func(r OnChangeType) hub.OnChangeType {
			return hub.OnChangeType(r)
		}),
		RunConfigurations: src.Spec.Triggers.RunConfigurations,
	}

	dst.Spec.Run.Artifacts = pipelines.Map(src.Spec.Run.Artifacts, convertToOutputArtifact)
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName

	dst.Status = hub.RunConfigurationStatus{
		Dependencies: hub.Dependencies{
			Pipeline: hub.PipelineReference{
				Version: src.Status.ObservedPipelineVersion,
			},
			RunConfigurations: pipelines.MapValues(src.Status.Dependencies.RunConfigurations, convertToRunReference),
		},
		Triggers: hub.TriggersStatus{
			Pipeline: hub.PipelineReference{
				Version: src.Status.TriggeredPipelineVersion,
			},
			RunConfigurations: pipelines.MapValues(src.Status.Triggers.RunConfigurations, convertToTriggeredRunReference),
			RunSpec: hub.RunSpecTriggerStatus{
				Version: src.Status.Triggers.RunSpec.Version,
			},
		},
		SynchronizationState: src.Status.SynchronizationState,
		Provider:             src.Status.Provider,
		ObservedGeneration:   src.Status.ObservedGeneration,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = pipelines.Map(src.Spec.Run.RuntimeParameters, convertFromRuntimeParameter)
	dst.Spec.Run.Pipeline = PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	dst.Spec.Triggers = Triggers{
		Schedules: src.Spec.Triggers.Schedules,
		OnChange: pipelines.Map(src.Spec.Triggers.OnChange, func(r hub.OnChangeType) OnChangeType {
			return OnChangeType(r)
		}),
		RunConfigurations: src.Spec.Triggers.RunConfigurations,
	}
	dst.Spec.Run.Artifacts = pipelines.Map(src.Spec.Run.Artifacts, convertFromOutputArtifact)
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName

	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion:  src.Status.Dependencies.Pipeline.Version,
		TriggeredPipelineVersion: src.Status.Triggers.Pipeline.Version,
		Dependencies: Dependencies{
			RunConfigurations: pipelines.MapValues(src.Status.Dependencies.RunConfigurations, convertFromRunReference),
		},
		Triggers: TriggersStatus{
			RunConfigurations: pipelines.MapValues(src.Status.Triggers.RunConfigurations, convertFromTriggeredRunReference),
			RunSpec: RunSpecTriggerStatus{
				Version: src.Status.Triggers.RunSpec.Version,
			},
		},
		SynchronizationState: src.Status.SynchronizationState,
		Provider:             src.Status.Provider,
		ObservedGeneration:   src.Status.ObservedGeneration,
	}

	return nil
}
