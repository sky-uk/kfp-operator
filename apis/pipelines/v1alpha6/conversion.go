package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultProviderNamespace string

var ResourceAnnotations = struct {
	ProviderNamespace string
}{
	ProviderNamespace: apis.Group + "/providerNamespace",
}

func namespaceToProvider(provider string, namespace string) common.NamespacedName {
	if namespace == "" {
		namespace = DefaultProviderNamespace
	}

	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
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
