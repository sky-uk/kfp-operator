package v1alpha5

import (
	"encoding/json"
	"errors"
	"fmt"
	common "github.com/sky-uk/kfp-operator/apis"

	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Provider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	var beamArgsPatchOps []common.JsonPatchOperation

	remainderFromConvertFrom := ProviderConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainderFromConvertFrom); err != nil {
		return err
	}

	for _, namedValue := range src.Spec.DefaultBeamArgs {
		patchOp := common.JsonPatchOperation{
			Op:   "add",
			Path: "/framework/parameters/beamArgs/0",
			Value: map[string]string{
				"name":  namedValue.Name,
				"value": namedValue.Value,
			},
		}

		beamArgsPatchOps = append(beamArgsPatchOps, patchOp)
	}

	tfxFramework := hub.Framework{
		Name:  "tfx",
		Image: DefaultTfxImage,
	}

	if len(beamArgsPatchOps) > 0 {
		patchOpsBytes, err := json.Marshal(beamArgsPatchOps)
		if err != nil {
			return fmt.Errorf("failed to marshal patch operations: %w", err)
		}

		patch := hub.Patch{
			Type:    "json",
			Payload: string(patchOpsBytes),
		}

		tfxFramework.Patches = []hub.Patch{patch}
	}

	dst.Spec.Frameworks = []hub.Framework{tfxFramework}
	dst.Spec.AllowedNamespaces = remainderFromConvertFrom.AllowedNamespaces

	if remainderFromConvertFrom.ServiceImage == "" {
		return errors.New("ServiceImage not set in remainder when converting to hub from v1alpha5")
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.ServiceImage = remainderFromConvertFrom.ServiceImage

	remainder := ProviderConversionRemainder{}
	remainder.ServiceImage = ""
	remainder.Image = src.Spec.Image

	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters
	return pipelines.SetConversionAnnotations(dst, &remainder)
}

func (dst *Provider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	remainderFromConvertTo := ProviderConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainderFromConvertTo); err != nil {
		return err
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	tfxFramework, found := common.Find(src.Spec.Frameworks, func(framework hub.Framework) bool {
		return framework.Name == "tfx"
	})

	if !found {
		return fmt.Errorf("tfx framework not in provider frameworks: %+v", src.Spec.Frameworks)
	}

	beamArgs, err := hub.BeamArgsFromJsonPatches(tfxFramework.Patches)
	if err != nil {
		return err
	}

	dst.Spec.DefaultBeamArgs = beamArgs
	dst.Spec.Image = tfxFramework.Image

	dst.Status.SynchronizationState = src.Status.Conditions.GetSyncStateFromReason()
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters
	dst.Spec.Image = remainderFromConvertTo.Image

	remainder := ProviderConversionRemainder{}
	remainder.ServiceImage = src.Spec.ServiceImage
	remainder.Image = ""
	remainder.AllowedNamespaces = src.Spec.AllowedNamespaces

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
