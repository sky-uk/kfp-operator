package v1beta1

import (
	"encoding/json"
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
