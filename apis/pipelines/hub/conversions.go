package v1beta1

import (
	"encoding/json"
	"errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func ToTFXPipelineFramework(tfxComponents string) PipelineFramework {
	marshal, _ := json.Marshal(tfxComponents)
	return PipelineFramework{
		Type:       "tfx",
		Parameters: map[string]*apiextensionsv1.JSON{"components": {Raw: marshal}},
	}
}

func FromPipelineFramework(pf PipelineFramework) (string, *PipelineConversionRemainder, error) {
	if pf.Type != "tfx" {
		return "", &PipelineConversionRemainder{
			Framework: pf,
		}, nil
	} else if pf.Type == "tfx" && pf.Parameters != nil {
		components, componentsExists := pf.Parameters["components"]
		if componentsExists {
			var res string
			if err := json.Unmarshal(components.Raw, &res); err != nil {
				return "", nil, err
			}
			return res, nil, nil
		} else {
			return "", nil, errors.New("missing components in tfx framework parameters")
		}
	} else {
		return "", nil, errors.New("missing tfx framework parameters")
	}

}
