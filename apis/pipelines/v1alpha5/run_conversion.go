package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status.ProviderId = src.Status.ProviderId
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Dependencies.Pipeline.Version = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = hub.CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Dependencies = hub.Dependencies{
		Pipeline: hub.PipelineReference{
			Version: src.Status.ObservedPipelineVersion,
		},
		RunConfigurations: src.Status.Dependencies.RunConfigurations,
	}

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status.ProviderId = src.Status.ProviderId
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.Dependencies.Pipeline.Version
	dst.Status.CompletionState = CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.Dependencies = Dependencies{
		RunConfigurations: src.Status.Dependencies.RunConfigurations,
	}

	return nil
}
