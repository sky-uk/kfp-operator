package jsonutil

import (
	"fmt"
	jsonpatch "github.com/evanphx/json-patch/v5"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

const (
	JsonPatch  = "json"
	MergePatch = "merge"
)

func PatchJson(patches []pipelineshub.Patch, json []byte) (string, error) {
	for _, patch := range patches {
		switch patch.Type {
		case JsonPatch:
			{
				patch, err := jsonpatch.DecodePatch([]byte(patch.Patch))
				if err != nil {
					return "", err
				}
				patchedJson, err := patch.ApplyWithOptions(json, &jsonpatch.ApplyOptions{EnsurePathExistsOnAdd: true})
				if err != nil {
					return "", err
				}
				json = patchedJson
			}
		case MergePatch:
			{
				patchedJson, err := jsonpatch.MergePatch(json, []byte(patch.Patch))
				if err != nil {
					return "", err
				}
				json = patchedJson
			}
		default:
			return "", fmt.Errorf("invalid patch type: %s", patch.Type)
		}

	}
	return string(json), nil
}
