package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = src.Status.ProviderId
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Status.Conditions = src.Status.Conditions

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = src.Spec

	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = src.Status.ProviderId
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Status.Conditions = src.Status.Conditions

	return nil
}
