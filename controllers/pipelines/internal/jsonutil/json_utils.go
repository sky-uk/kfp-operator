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
	for _, p := range patches {
		switch p.Type {
		case JsonPatch:
			{
				patch, err := jsonpatch.DecodePatch([]byte(p.Patch))
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
				patchedJson, err := jsonpatch.MergePatch(json, []byte(p.Patch))
				if err != nil {
					return "", err
				}
				json = patchedJson
			}
		default:
			return "", fmt.Errorf("invalid patch type: %s", p.Type)
		}

	}
	return string(json), nil
}
