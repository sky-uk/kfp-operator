package v1beta1

import (
	"encoding/json"
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func AddComponentsToFrameworkParams(tfxComponents string, framework *PipelineFramework) error {
	marshal, err := json.Marshal(tfxComponents)
	if err != nil {
		return err
	}
	framework.Parameters["components"] = &apiextensionsv1.JSON{Raw: marshal}
	return nil
}

func AddBeamArgsToFrameworkParams(beamArgs []apis.NamedValue, framework *PipelineFramework) error {
	marshal, err := json.Marshal(beamArgs)
	if err != nil {
		return err
	}
	framework.Parameters["beamArgs"] = &apiextensionsv1.JSON{Raw: marshal}
	return nil
}

func ComponentsFromFramework(framework *PipelineFramework) (string, error) {
	if framework.Parameters == nil {
		return "", nil
	}
	components, componentsExists := framework.Parameters["components"]
	if componentsExists {
		var res string
		if err := json.Unmarshal(components.Raw, &res); err != nil {
			return "", err
		}
		return res, nil
	}

	return "", nil
}

func BeamArgsFromFramework(framework *PipelineFramework) ([]apis.NamedValue, error) {
	if framework.Parameters == nil {
		return []apis.NamedValue{}, nil
	}
	beamArgs, beamArgsExists := framework.Parameters["beamArgs"]
	if beamArgsExists {
		var res []apis.NamedValue
		if err := json.Unmarshal(beamArgs.Raw, &res); err != nil {
			return nil, err
		}

		return res, nil
	}

	return []apis.NamedValue{}, nil
}

func BeamArgsFromJsonPatches(patches []Patch) ([]apis.NamedValue, error) {
	var namedValues []apis.NamedValue
	for _, patch := range patches {
		var patchOps []apis.JsonPatchOperation
		if err := json.Unmarshal([]byte(patch.Patch), &patchOps); err != nil {
			return nil, err
		}
		for _, patchOp := range patchOps {
			nameValue, ok := patchOp.Value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected map[string]any, got %v", patchOp.Value)
			}

			name, ok := nameValue["name"].(string)
			if !ok {
				return nil, fmt.Errorf("expected string for 'name', got %v", nameValue["name"])
			}

			value, ok := nameValue["value"].(string)
			if !ok {
				return nil, fmt.Errorf("expected string for 'value', got %v", nameValue["value"])
			}

			namedValues = append(namedValues, apis.NamedValue{
				Name:  name,
				Value: value,
			})
		}
	}
	return namedValues, nil
}
