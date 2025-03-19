package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var DefaultProvider string
var DefaultProviderNamespace string

func convertArtifactsTo(outputArtifact []OutputArtifact) []hub.OutputArtifact {
	var hubOutputArtifact []hub.OutputArtifact
	for _, artifact := range outputArtifact {
		hubOutputArtifact = append(hubOutputArtifact, hub.OutputArtifact{
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
	return hubOutputArtifact
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

func convertProviderAndIdFrom(providerAndId hub.ProviderAndId) ProviderAndId {
	return ProviderAndId{
		Provider: providerAndId.Name.Name,
		Id:       providerAndId.Id,
	}
}

func convertProviderTo(
	provider string,
	namespace string,
) common.NamespacedName {
	if provider == "" {
		provider = DefaultProvider
	}
	if namespace == "" {
		namespace = DefaultProviderNamespace
	}
	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
}
