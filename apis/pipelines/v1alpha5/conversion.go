package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

func convertRuntimeParametersTo(rtp []RuntimeParameter) []hub.RuntimeParameter {
	convertValueFromTo := func(v *ValueFrom) *hub.ValueFrom {
		if v != nil {
			return &hub.ValueFrom{
				RunConfigurationRef: hub.RunConfigurationRef{
					Name:           v.RunConfigurationRef.Name,
					OutputArtifact: v.RunConfigurationRef.OutputArtifact,
				},
			}
		}
		return nil
	}
	var hubRtp []hub.RuntimeParameter
	for _, namedValue := range rtp {
		hubRtp = append(hubRtp, hub.RuntimeParameter{
			Name:      namedValue.Name,
			Value:     namedValue.Value,
			ValueFrom: convertValueFromTo(namedValue.ValueFrom),
		})
	}
	return hubRtp
}

func convertArtifactsTo(oa []OutputArtifact) []hub.OutputArtifact {
	var hubOa []hub.OutputArtifact
	for _, artifact := range oa {
		hubOa = append(hubOa, hub.OutputArtifact{
			Name: artifact.Name,
			Path: hub.ArtifactPath{
				Locator: hub.ArtifactLocator{
					Component: artifact.Path.Locator.Component,
					Artifact:  artifact.Path.Locator.Artifact,
					Index:     artifact.Path.Locator.Index,
				},
				Filter: artifact.Path.Filter,
			},
		})
	}
	return hubOa
}

func convertRuntimeParametersFrom(hubRtp []hub.RuntimeParameter) []RuntimeParameter {
	convertValueFromFrom := func(v *hub.ValueFrom) *ValueFrom {
		if v != nil {
			return &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					Name:           v.RunConfigurationRef.Name,
					OutputArtifact: v.RunConfigurationRef.OutputArtifact,
				},
			}
		}
		return nil
	}
	var rtp []RuntimeParameter
	for _, namedValue := range hubRtp {
		rtp = append(rtp, RuntimeParameter{
			Name:      namedValue.Name,
			Value:     namedValue.Value,
			ValueFrom: convertValueFromFrom(namedValue.ValueFrom),
		})
	}
	return rtp
}

func convertArtifactsFrom(hubArtifacts []hub.OutputArtifact) []OutputArtifact {
	var artifacts []OutputArtifact
	for _, artifact := range hubArtifacts {
		artifacts = append(artifacts, OutputArtifact{
			Name: artifact.Name,
			Path: ArtifactPath{
				Locator: ArtifactLocator{
					artifact.Path.Locator.Component,
					artifact.Path.Locator.Artifact,
					artifact.Path.Locator.Index,
				},
				Filter: artifact.Path.Filter,
			},
		})
	}
	return artifacts
}
