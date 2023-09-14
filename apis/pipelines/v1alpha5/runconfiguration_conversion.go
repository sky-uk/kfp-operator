package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status = hub.RunConfigurationStatus{
		Dependencies: hub.Dependencies{
			Pipeline: hub.PipelineReference{
				Version: src.Status.ObservedPipelineVersion,
			},
			RunConfigurations: src.Status.Dependencies.RunConfigurations,
		},
		Triggers: hub.TriggersStatus{
			Pipeline: hub.PipelineReference{
				Version: src.Status.TriggeredPipelineVersion,
			},
			RunConfigurations: src.Status.Triggers.RunConfigurations,
			RunSpec: hub.RunSpecTriggerStatus{
				Version: src.Status.Triggers.RunSpec.Version,
			},
		},
		SynchronizationState: src.Status.SynchronizationState,
		Provider:             src.Status.Provider,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           src.Status.Conditions,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion:  src.Status.Dependencies.Pipeline.Version,
		TriggeredPipelineVersion: src.Status.Triggers.Pipeline.Version,
		Dependencies: Dependencies{
			RunConfigurations: src.Status.Dependencies.RunConfigurations,
		},
		Triggers: TriggersStatus{
			RunConfigurations: src.Status.Triggers.RunConfigurations,
			RunSpec:           src.Status.Triggers.RunSpec,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Provider:             src.Status.Provider,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           src.Status.Conditions,
	}

	return nil
}
