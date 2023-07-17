package v1alpha4

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha5remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.RuntimeParameters = hub.MergeRuntimeParameters(src.Spec.RuntimeParameters, v1alpha5remainder.ValueFromParameters)
	dst.Spec.Artifacts = v1alpha5remainder.Artifacts
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = hub.CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	v1alpha5remainder.Artifacts = src.Spec.Artifacts
	dst.Spec.RuntimeParameters, v1alpha5remainder.ValueFromParameters = hub.SplitRunTimeParameters(src.Spec.RuntimeParameters)
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.Version = src.Status.Version
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.CompletionState = CompletionState(src.Status.CompletionState)
	dst.Status.MarkedCompletedAt = src.Status.MarkedCompletedAt

	return pipelines.SetConversionAnnotations(dst, &v1alpha5remainder)
}
