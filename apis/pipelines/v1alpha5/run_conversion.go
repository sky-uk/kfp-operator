package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = convertRuntimeParametersToHub(src.Spec.RuntimeParameters)
	dst.Spec.Artifacts = convertArtifactsToHub(src.Spec.Artifacts)

	dst.Status.ProviderId.Provider = src.Status.ProviderId.Provider
	dst.Status.ProviderId.Id = src.Status.ProviderId.Id

	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = hub.CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Conditions = hub.Conditions(src.Status.Conditions)
	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = convertRuntimeParametersFromHub(src.Spec.RuntimeParameters)
	dst.Spec.Artifacts = convertArtifactsFromHub(src.Spec.Artifacts)

	dst.Status.ProviderId.Provider = src.Status.ProviderId.Provider
	dst.Status.ProviderId.Id = src.Status.ProviderId.Id

	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Conditions = Conditions(src.Status.Conditions)

	return nil
}
