package v1alpha3

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Experiment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = src.Status.KfpId
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.KfpId = src.Status.ProviderId
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}
