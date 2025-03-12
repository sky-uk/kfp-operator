package v1beta1

import (
	"encoding/json"
	"errors"
	"github.com/sky-uk/kfp-operator/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func AddComponentsToFrameworkParams(tfxComponents string, framework *PipelineFramework) {
	marshal, _ := json.Marshal(tfxComponents)
	framework.Parameters["components"] = &apiextensionsv1.JSON{Raw: marshal}
}

func AddBeamArgsToFrameworkParams(beamArgs []apis.NamedValue, framework *PipelineFramework) error {
	beamArgsMap := make(map[string]string, len(beamArgs))
	for _, arg := range beamArgs {
		beamArgsMap[arg.Name] = arg.Value
	}

	marshal, err := json.Marshal(beamArgsMap)
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
	} else {
		return "", errors.New("missing components in tfx framework parameters")
	}
}

func BeamArgsFromFramework(framework *PipelineFramework) ([]apis.NamedValue, error) {
	if framework.Parameters == nil {
		return []apis.NamedValue{}, nil
	}
	beamArgs, beamArgsExists := framework.Parameters["beamArgs"]
	if beamArgsExists {
		var res map[string]string
		if err := json.Unmarshal(beamArgs.Raw, &res); err != nil {
			return nil, err
		}

		var beamArgs []apis.NamedValue
		for name, value := range res {
			beamArgs = append(beamArgs, apis.NamedValue{
				Name:  name,
				Value: value,
			})
		}

		return beamArgs, nil
	} else {
		return nil, errors.New("missing beamArgs in tfx framework parameters")
	}
}
