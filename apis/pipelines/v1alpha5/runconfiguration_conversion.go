package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersToHub(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsToHub(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersToHub(src.Spec.Triggers)

	dst.Status = hub.RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersFromHub(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsFromHub(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersFromHub(src.Spec.Triggers)
	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}
	return nil
}
