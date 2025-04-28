package v1alpha6

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

	for _, nv := range src.Spec.DefaultBeamArgs {
		patchOp := common.JsonPatchOperation{
			Op:   "add",
			Path: "/framework/parameters/beamArgs/-",
			Value: map[string]string{
				"name":  nv.Name,
				"value": nv.Value,
			},
		}

		beamArgsPatchOps = append(beamArgsPatchOps, patchOp)
	}

	tfxFramework := hub.Framework{
		Name:  "tfx",
		Image: src.Spec.Image,
	}

	if len(beamArgsPatchOps) > 0 {
		patchOpsBytes, err := json.Marshal(beamArgsPatchOps)
		if err != nil {
			return errors.New("failed to marshal patch operations to JSON")
		}

		patch := hub.Patch{
			Type:  "json",
			Patch: string(patchOpsBytes),
		}

		tfxFramework.Patches = []hub.Patch{patch}
	}

	dst.Spec.Frameworks = []hub.Framework{tfxFramework}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	remainder := ProviderConversionRemainder{}

	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters

	return pipelines.SetConversionAnnotations(dst, &remainder)
}

func (dst *Provider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	remainder := ProviderConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	tfxFramework, found := common.Find(src.Spec.Frameworks, func(framework hub.Framework) bool {
		return framework.Name == "tfx"
	})

	if found {
		beamArgs, err := hub.BeamArgsFromJsonPatches(tfxFramework.Patches)
		if err != nil {
			return err
		}
		dst.Spec.DefaultBeamArgs = beamArgs
		dst.Spec.Image = tfxFramework.Image
	} else {
		return fmt.Errorf("tfx framework not in provider frameworks: %v", src.Spec.Frameworks)
	}

	status := src.Status.Conditions.GetSyncStateFromReason()

	dst.Status.SynchronizationState = status
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters

	return nil
}
