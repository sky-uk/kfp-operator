package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.RuntimeParameters = pipelines.Map(src.Spec.RuntimeParameters, convertToRuntimeParameter)

	dst.Spec.Artifacts = pipelines.Map(src.Spec.Artifacts, convertToOutputArtifact)
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Dependencies.Pipeline.Version = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = hub.CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Dependencies = hub.Dependencies{
		Pipeline: hub.PipelineReference{
			Version: src.Status.ObservedPipelineVersion,
		},
		RunConfigurations: pipelines.MapValues(src.Status.Dependencies.RunConfigurations, convertToRunReference),
	}

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}

	dst.Spec.RuntimeParameters = pipelines.Map(src.Spec.RuntimeParameters, convertFromRuntimeParameter)

	dst.Spec.Artifacts = pipelines.Map(src.Spec.Artifacts, convertFromOutputArtifact)
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.Dependencies.Pipeline.Version
	dst.Status.CompletionState = CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Dependencies = Dependencies{
		RunConfigurations: pipelines.MapValues(src.Status.Dependencies.RunConfigurations, convertFromRunReference),
	}

	return nil
}
