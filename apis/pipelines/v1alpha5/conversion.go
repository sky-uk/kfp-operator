package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultProvider string

var ResourceAnnotations = struct {
	Provider string
}{
	Provider: apis.Group + "/provider",
}

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

func getProviderAnnotation(resource v1.Object) string {
	if provider, hasProvider := resource.GetAnnotations()[ResourceAnnotations.Provider]; hasProvider {
		return provider
	}
	return DefaultProvider
}

func setProviderAnnotation(provider string, resource *v1.ObjectMeta) {
	v1.SetMetaDataAnnotation(resource, ResourceAnnotations.Provider, provider)
}

func removeProviderAnnotation(resource v1.Object) {
	delete(resource.GetAnnotations(), ResourceAnnotations.Provider)
}

func convertProviderAndIdTo(providerAndId ProviderAndId) hub.ProviderAndId {
	return hub.ProviderAndId{
		Name: providerAndId.Provider,
		Id:   providerAndId.Id,
	}
}

func convertProviderAndIdFrom(providerAndId hub.ProviderAndId) ProviderAndId {
	return ProviderAndId{
		Provider: providerAndId.Name,
		Id:       providerAndId.Id,
	}
}
