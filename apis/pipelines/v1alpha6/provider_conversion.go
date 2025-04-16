package v1alpha6

import (
	"encoding/json"
	"errors"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Provider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	type patchOperation struct {
		Op    string            `json:"op"`
		Path  string            `json:"path"`
		Value map[string]string `json:"value"`
	}

	var patches []hub.Patch

	for _, nv := range src.Spec.DefaultBeamArgs {
		patchOp := patchOperation{
			Op:   "add",
			Path: "/Framework/parameters/beamArgs/-",
			Value: map[string]string{
				"name":  nv.Name,
				"value": nv.Value,
			},
		}

		patchBytes, err := json.Marshal(patchOp)
		if err != nil {
			return errors.New("failed to marshal patch operation to JSON")
		}

		patches = append(patches, hub.Patch{
			Type:  "json",
			Patch: string(patchBytes),
		})
	}

	dst.Spec.Frameworks = []hub.Framework{{
		Name:    "tfx",
		Image:   src.Spec.Image,
		Patches: patches,
	}}

	dst.TypeMeta.APIVersion = dstApiVersion
	return nil
}
