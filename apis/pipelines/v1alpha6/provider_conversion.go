package v1alpha6

import (
	"encoding/json"
	"errors"
	common "github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Provider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	var patches []hub.Patch

	for _, nv := range src.Spec.DefaultBeamArgs {
		patchOp := common.PatchOperation{
			Op:   "add",
			Path: "/framework/parameters/beamArgs/-",
			Value: map[string]string{
				"name":  nv.Name,
				"value": nv.Value,
			},
		}
	remainder := ProviderConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}
	remainder.Image = src.Spec.Image

	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters

	return pipelines.SetConversionAnnotations(dst, &remainder)
}

func (dst *Provider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	remainder := ProviderConversionRemainder{}

		patchBytes, err := json.Marshal(patchOp)
		if err != nil {
			return errors.New("failed to marshal patch operation to JSON")
		}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}

		patches = append(patches, hub.Patch{
			Type:  "json",
			Patch: string(patchBytes),
		})
	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Frameworks = []hub.Framework{{
		Name:    "tfx",
		Image:   src.Spec.Image,
		Patches: patches,
	}}
	status := src.Status.Conditions.GetSyncStateFromReason()

	dst.Status.SynchronizationState = status
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Image = remainder.Image
	dst.Spec.Parameters = src.Spec.Parameters

	return nil
}
