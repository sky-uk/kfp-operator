package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultProvider string
var DefaultProviderNamespace string

var ResourceAnnotations = struct {
	Provider          string
	ProviderNamespace string
}{
	Provider:          apis.Group + "/provider",
	ProviderNamespace: apis.Group + "/providerNamespace",
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
	if provider, ok := resource.GetAnnotations()[ResourceAnnotations.Provider]; ok {
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

func getProviderNamespaceAnnotation(resource v1.Object) string {
	if providerNamespace, ok := resource.GetAnnotations()[ResourceAnnotations.ProviderNamespace]; ok {
		return providerNamespace
	}
	return ""
}

func setProviderNamespaceAnnotation(namespace string, resource *v1.ObjectMeta) {
	v1.SetMetaDataAnnotation(resource, ResourceAnnotations.ProviderNamespace, namespace)
}

func removeProviderNamespaceAnnotation(resource v1.Object) {
	delete(resource.GetAnnotations(), ResourceAnnotations.ProviderNamespace)
}

func convertProviderAndIdTo(providerAndId ProviderAndId, namespace string) hub.ProviderAndId {
	return hub.ProviderAndId{
		Name: common.NamespacedName{
			Name:      providerAndId.Provider,
			Namespace: namespace,
		},
		Id: providerAndId.Id,
	}
}

func convertProviderAndIdFrom(providerAndId hub.ProviderAndId) ProviderAndId {
	return ProviderAndId{
		Provider: providerAndId.Name.Name,
		Id:       providerAndId.Id,
	}
}

func namespaceToProvider(resource v1.Object) common.NamespacedName {
	provider := getProviderAnnotation(resource)
	providerNamespace := getProviderNamespaceAnnotation(resource)
	if providerNamespace == "" {
		providerNamespace = DefaultProviderNamespace
	}
	return common.NamespacedName{
		Name:      provider,
		Namespace: providerNamespace,
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
