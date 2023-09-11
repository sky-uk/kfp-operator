package v1alpha5

import hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"

func convertToRuntimeParameter(srt RuntimeParameter) hub.RuntimeParameter {
	drt := hub.RuntimeParameter{
		Name:  srt.Name,
		Value: srt.Value,
	}

	if srt.ValueFrom != nil {
		drt.ValueFrom = &hub.ValueFrom{
			RunConfigurationRef: hub.RunConfigurationRef{
				Name:           srt.ValueFrom.RunConfigurationRef.Name,
				OutputArtifact: srt.ValueFrom.RunConfigurationRef.OutputArtifact,
			},
		}
	}

	return drt
}

func convertFromRuntimeParameter(srt hub.RuntimeParameter) RuntimeParameter {
	drt := RuntimeParameter{
		Name:  srt.Name,
		Value: srt.Value,
	}

	if srt.ValueFrom != nil {
		drt.ValueFrom = &ValueFrom{
			RunConfigurationRef: RunConfigurationRef{
				Name:           srt.ValueFrom.RunConfigurationRef.Name,
				OutputArtifact: srt.ValueFrom.RunConfigurationRef.OutputArtifact,
			},
		}
	}

	return drt
}

func convertToOutputArtifact(sa OutputArtifact) hub.OutputArtifact {
	return hub.OutputArtifact{
		Name: sa.Name,
		Path: hub.ArtifactPath{
			Locator: hub.ArtifactLocator{
				Component: sa.Path.Locator.Component,
				Artifact:  sa.Path.Locator.Artifact,
				Index:     sa.Path.Locator.Index,
			},
			Filter: sa.Path.Filter,
		},
	}
}

func convertFromOutputArtifact(sa hub.OutputArtifact) OutputArtifact {
	return OutputArtifact{
		Name: sa.Name,
		Path: ArtifactPath{
			Locator: ArtifactLocator{
				Component: sa.Path.Locator.Component,
				Artifact:  sa.Path.Locator.Artifact,
				Index:     sa.Path.Locator.Index,
			},
			Filter: sa.Path.Filter,
		},
	}
}

func convertToRunReference(r RunReference) hub.RunReference {
	return hub.RunReference{
		ProviderId: r.ProviderId,
		Artifacts:  r.Artifacts,
	}
}

func convertFromRunReference(r hub.RunReference) RunReference {
	return RunReference{
		ProviderId: r.ProviderId,
		Artifacts:  r.Artifacts,
	}
}

func convertToTriggeredRunReference(r TriggeredRunReference) hub.TriggeredRunReference {
	return hub.TriggeredRunReference{
		ProviderId: r.ProviderId,
	}
}

func convertFromTriggeredRunReference(r hub.TriggeredRunReference) TriggeredRunReference {
	return TriggeredRunReference{
		ProviderId: r.ProviderId,
	}
}
